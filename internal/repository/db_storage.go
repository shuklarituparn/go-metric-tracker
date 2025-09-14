package repository

import (
	"database/sql"
	"fmt"
	"log"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type DBStorage struct {
	DB *sql.DB
}

func NewDBStorage(db *sql.DB) (*DBStorage, error) {
	storage := &DBStorage{
		DB: db,
	}
	if err := storage.initSchema(); err != nil {
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
	if name == "" {
		return 0, false
	}
	var value sql.NullInt64
	const query = `SELECT delta FROM metrics WHERE id=$1 and type=$2`
	err := ds.DB.QueryRow(query, name, models.Counter).Scan(&value)

	if err != nil || !value.Valid {
		return 0, false
	}
	return value.Int64, true

}

func (ds *DBStorage) GetGauge(name string) (float64, bool) {
	if name == "" {
		return 0, false
	}
	var value sql.NullFloat64
	const query = `SELECT value FROM metrics WHERE id=$1 and type =$2`
	err := ds.DB.QueryRow(query, name, models.Gauge).Scan(&value)
	if err != nil || !value.Valid {
		return 0, false
	}
	return value.Float64, true
}

func (ds *DBStorage) GetMetric(name string) (*models.Metrics, bool) {
	if name == "" {
		return nil, false
	}
	var metric models.Metrics
	var value sql.NullFloat64
	var delta sql.NullInt64

	const query = `SELECT id, type, delta, value FROM metrics WHERE id=$1`

	err := ds.DB.QueryRow(query, name).Scan(&metric.ID, &metric.MType, &delta, &value)
	if err != nil {
		return nil, false
	}

	if delta.Valid {
		metric.Delta = &delta.Int64
	}

	if value.Valid {
		metric.Value = &value.Float64
	}
	return &metric, true
}

func (ds *DBStorage) GetAllMetrics() []models.Metrics {
	var metrics []models.Metrics
	const query = `SELECT id, type, delta, value FROM metrics ORDER BY id`

	rows, err := ds.DB.Query(query)
	if err != nil {
		log.Printf("failed to get all metrics: %v", err)
		return metrics
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
	
	return metrics
}

func (ds *DBStorage) Close() error {
	return ds.DB.Close()
}

func (ds *DBStorage) Ping() error {
	return ds.DB.Ping()
}