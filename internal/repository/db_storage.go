package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type DBStorage struct {
	DB          *sql.DB
	retryConfig retry.RetryConfig
}

func NewDBStorage(db *sql.DB) (*DBStorage, error) {
	storage := &DBStorage{
		DB:          db,
		retryConfig: retry.DefaultConfig(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := retry.ExecuteWithRetry(ctx, storage.retryConfig, func() error {
		return storage.initSchema()
	})
	if err != nil {
		return nil, fmt.Errorf("error: problem while initializing schema: %w", err)
	}
	return storage, nil
}

func (ds *DBStorage) initSchema() error {
	const query = `
	CREATE TABLE IF NOT EXISTS metrics(
	id VARCHAR(255) PRIMARY KEY,
	type VARCHAR(255) NOT NULL,
	delta BIGINT,
	value DOUBLE PRECISION ,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_metric_type ON metrics(type);
	`

	val, err := ds.DB.Exec(query)
	log.Printf("got following result: %v", val)
	return err
}
func (ds *DBStorage) UpdateGauge(name string, value float64) error {
	if name == "" {
		return fmt.Errorf("name can't be empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		return ds.updateGaugeInternal(name, value)
	})
}

func (ds *DBStorage) updateGaugeInternal(name string, value float64) error {
	const query = `
	INSERT INTO metrics (id, type, delta, value, updated_at)
	VALUES ($1, $2, NULL, $3, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE 
	SET
	value = EXCLUDED.value,
	delta = NULL,
	type = EXCLUDED.type,
	updated_at= CURRENT_TIMESTAMP;
	`
	result, err := ds.DB.Exec(query, name, models.Gauge, value)
	log.Printf("got the following result: %v", result)
	if err != nil {
		return fmt.Errorf("err: while updating the gauage in db: %w", err)
	}
	return nil
}

func (ds *DBStorage) UpdateCounter(name string, inc int64) error {
	if name == "" {
		return fmt.Errorf("counter metric name cannot be empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		return ds.updateCounterInternal(name, inc)
	})
}
func (ds *DBStorage) updateCounterInternal(name string, inc int64) error {
	if name == "" {
		return fmt.Errorf("counter metric name cannot be empty")
	}

	const q = `
	INSERT INTO metrics (id, type, delta, value, updated_at)
	VALUES ($1, $2, $3, NULL, CURRENT_TIMESTAMP)
	ON CONFLICT (id) DO UPDATE
	SET delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
	    value = NULL,
	    type  = EXCLUDED.type,
	    updated_at = CURRENT_TIMESTAMP;
	`
	if _, err := ds.DB.Exec(q, name, models.Counter, inc); err != nil {
		return fmt.Errorf("update counter %q: %w", name, err)
	}
	return nil
}
func (ds *DBStorage) GetCounter(name string) (int64, bool) {
	var value int64
	var found bool

	if name == "" {
		return 0, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		var err error
		value, found, err = ds.getCounterInternal(name)
		return err
	})
	if err != nil {
		log.Printf("Error getting counter %s: %v", name, err)
		return 0, false
	}
	return value, found
}

func (ds *DBStorage) getCounterInternal(name string) (int64, bool, error) {

	var value sql.NullInt64
	const query = `SELECT delta FROM metrics WHERE id=$1 and type=$2`
	err := ds.DB.QueryRow(query, name, models.Counter).Scan(&value)

	if err != nil || !value.Valid {
		return 0, false, err
	}
	return value.Int64, true, nil
}

func (ds *DBStorage) GetGauge(name string) (float64, bool) {
	var value float64
	var found bool
	if name == "" {
		return 0, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		var err error
		value, found, err = ds.getGaugeInternal(name)
		return err
	})
	if err != nil {
		log.Printf("Error getting gauge %s: %v", name, err)
		return 0, false
	}

	return value, found
}

func (ds *DBStorage) getGaugeInternal(name string) (float64, bool, error) {
	var value sql.NullFloat64
	const query = `SELECT value FROM metrics WHERE id=$1 and type =$2`
	err := ds.DB.QueryRow(query, name, models.Gauge).Scan(&value)
	if err != nil || !value.Valid {
		return 0, false, err
	}
	return value.Float64, true, nil
}

func (ds *DBStorage) GetMetric(name string) (*models.Metrics, bool) {
	if name == "" {
		return nil, false
	}
	ctx, cancel:= context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	var metric *models.Metrics
	var found bool
	
	err:= retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		var err error
		metric, found, err= ds.getMetricInternal(name)
		return err
	})
	if err != nil {
			log.Printf("Error getting metric %s: %v", name, err)
			return nil, false
		}
		
		return metric, found
}

func (ds *DBStorage) getMetricInternal(name string) (*models.Metrics, bool, error) {

	var metric models.Metrics
	var value sql.NullFloat64
	var delta sql.NullInt64

	const query = `SELECT id, type, delta, value FROM metrics WHERE id=$1`

	err := ds.DB.QueryRow(query, name).Scan(&metric.ID, &metric.MType, &delta, &value)
	if err != nil {
		return nil, false, err
	}

	if delta.Valid {
		metric.Delta = &delta.Int64
	}

	if value.Valid {
		metric.Value = &value.Float64
	}
	return &metric, true, nil
}

func (ds *DBStorage) GetAllMetrics() []models.Metrics {
	ctx,cancel:= context.WithTimeout(context.Background(), 30* time.Second)
	defer cancel()
	
	var metrics []models.Metrics
	err:= retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		var err error
		metrics, err= ds.getAllMetricsInternal()
		return err
	})
	if err != nil {
			log.Printf("Error getting all metrics: %v", err)
			return []models.Metrics{}
		}
	return metrics 
	
}

func (ds *DBStorage) getAllMetricsInternal() ([]models.Metrics, error) {
	var metrics []models.Metrics
	const query = `SELECT id, type, delta, value FROM metrics ORDER BY id`

	rows, err := ds.DB.Query(query)
	if err != nil {
		log.Printf("failed to get all metrics: %v", err)
		return metrics, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error: error closing rows: %v", err)
		}
	}()

	for rows.Next() {
		var metric models.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64

		if err := rows.Scan(&metric.ID, &metric.MType, &delta, &value); err != nil {
			log.Printf("Failed to scan metric: %v", err)
			continue
		}
		if delta.Valid {
			metric.Delta = &delta.Int64
		}
		if value.Valid {
			metric.Value = &value.Float64
		}
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
	}

	return metrics, nil
}

func (ds *DBStorage) Close() error {
	return ds.DB.Close()
}

func (ds *DBStorage) Ping() error {
	return ds.DB.Ping()
}

func (ds *DBStorage) UpdateBatch(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	ctx, cancel:= context.WithTimeout(context.Background(), 30* time.Second)
	defer cancel()
	
	return retry.ExecuteWithRetry(ctx, ds.retryConfig, func() error {
		return ds.updateBatchInternal(metrics)
	})

}

func (ds *DBStorage) updateBatchInternal(metrics []models.Metrics) error {

	tx, err := ds.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction :%w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}
	}()
	gaugeStmt, err := tx.Prepare(`
			INSERT INTO metrics (id, type, delta, value, updated_at)
			VALUES ($1, $2, NULL, $3, CURRENT_TIMESTAMP)
			ON CONFLICT(id) DO UPDATE 
			SET
				value = EXCLUDED.value,
				delta = NULL,
				type = EXCLUDED.type,
				updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare gauge statement: %w", err)
	}
	defer func() {
		if err := gaugeStmt.Close(); err != nil {
			log.Printf("error: problem closing gauge statement: %v", err)
		}
	}()

	counterStatement, err := tx.Prepare(`
		INSERT INTO metrics (id, type, delta, value, updated_at)
				VALUES ($1, $2, $3, NULL, CURRENT_TIMESTAMP)
				ON CONFLICT (id) DO UPDATE
				SET 
					delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
					value = NULL,
					type = EXCLUDED.type,
					updated_at = CURRENT_TIMESTAMP
		`)
	if err != nil {
		return fmt.Errorf("failed to prepare counter statement: %w", err)
	}
	defer func() {
		if err := counterStatement.Close(); err != nil {
			log.Printf("error: problem closing counter statement: %v", err)
		}
	}()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				if _, err := gaugeStmt.Exec(metric.ID, models.Gauge, *metric.Value); err != nil {
					return fmt.Errorf("failed to update gauge %s: %w", metric.ID, err)
				}
			}
		case models.Counter:
			if metric.Delta != nil {
				if _, err := counterStatement.Exec(metric.ID, models.Counter, *metric.Delta); err != nil {
					return fmt.Errorf("failed to update counter %s: %w", metric.ID, err)
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
