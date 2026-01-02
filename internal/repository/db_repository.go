package repository

import (
	"database/sql"
	"fmt"
	"log"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (r *DBRepository) InitSchema() error {
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

	_, err := r.db.Exec(query)
	return err
}

func (r *DBRepository) UpdateGauge(name string, value float64) error {
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
	result, err := r.db.Exec(query, name, models.Gauge, value)
	log.Printf("got the following result: %v", result)
	if err != nil {
		return fmt.Errorf("err: while updating the gauage in db: %w", err)
	}
	return nil
}

func (r *DBRepository) UpdateCounter(name string, inc int64) error {
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
	if _, err := r.db.Exec(q, name, models.Counter, inc); err != nil {
		return fmt.Errorf("update counter %q: %w", name, err)
	}
	return nil
}

func (r *DBRepository) GetCounter(name string) (int64, bool, error) {

	var value sql.NullInt64
	const query = `SELECT delta FROM metrics WHERE id=$1 and type=$2`
	err := r.db.QueryRow(query, name, models.Counter).Scan(&value)

	if err != nil || !value.Valid {
		return 0, false, err
	}
	return value.Int64, true, nil
}

func (r *DBRepository) GetGauge(name string) (float64, bool, error) {
	var value sql.NullFloat64
	const query = `SELECT value FROM metrics WHERE id=$1 and type =$2`
	err := r.db.QueryRow(query, name, models.Gauge).Scan(&value)
	if err != nil || !value.Valid {
		return 0, false, err
	}
	return value.Float64, true, nil
}

func (r *DBRepository) GetMetric(name string) (*models.Metrics, bool, error) {

	var metric models.Metrics
	var value sql.NullFloat64
	var delta sql.NullInt64

	const query = `SELECT id, type, delta, value FROM metrics WHERE id=$1`

	err := r.db.QueryRow(query, name).Scan(&metric.ID, &metric.MType, &delta, &value)
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

func (r *DBRepository) GetAllMetrics() ([]models.Metrics, error) {
	var metrics []models.Metrics
	const query = `SELECT id, type, delta, value FROM metrics ORDER BY id`

	rows, err := r.db.Query(query)
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
func (r *DBRepository) UpdateBatch(metrics []models.Metrics) error {

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction :%w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("failed to rollback transaction: %v", rollbackErr)
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
		if closeErr := gaugeStmt.Close(); closeErr != nil {
			log.Printf("error: problem closing gauge statement: %v", closeErr)
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
		if closeErr := counterStatement.Close(); closeErr != nil {
			log.Printf("error: problem closing counter statement: %v", closeErr)
		}
	}()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				if _, execErr := gaugeStmt.Exec(metric.ID, models.Gauge, *metric.Value); execErr != nil {
					return fmt.Errorf("failed to update gauge %s: %w", metric.ID, execErr)
				}
			}
		case models.Counter:
			if metric.Delta != nil {
				if _, execErr := counterStatement.Exec(metric.ID, models.Counter, *metric.Delta); execErr != nil {
					return fmt.Errorf("failed to update counter %s: %w", metric.ID, execErr)
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *DBRepository) Close() error {
	return r.db.Close()
}

func (r *DBRepository) Ping() error {
	return r.db.Ping()
}