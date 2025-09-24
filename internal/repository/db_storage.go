package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type DBStorage struct {
	dr          *DBRepository
	retryConfig retry.RetryConfig
}

func NewDBStorage(dr *DBRepository) (*DBStorage, error) {
	storage := &DBStorage{
		dr:          dr,
		retryConfig: retry.DefaultConfig(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := retry.ExecuteWithRetry(ctx, storage.retryConfig, func() (struct{}, error) {
		return struct{}{}, dr.InitSchema()
	})
	if err != nil {
		return nil, fmt.Errorf("error: problem while initializing schema: %w", err)
	}
	return storage, nil
}

func (ds *DBStorage) UpdateGauge(name string, value float64) error {
	if name == "" {
		return fmt.Errorf("name can't be empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() (struct{}, error) {
		return struct{}{}, ds.dr.UpdateGauge(name, value)
	})
	return err
}

func (ds *DBStorage) UpdateCounter(name string, inc int64) error {
	if name == "" {
		return fmt.Errorf("counter metric name cannot be empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() (struct{}, error) {
		return struct{}{}, ds.dr.UpdateCounter(name, inc)
	})
	return err
}

func (ds *DBStorage) GetCounter(name string) (int64, bool) {
	if name == "" {
		return 0, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() (struct {
		value int64
		found bool
	}, error) {
		value, found, err := ds.dr.GetCounter(name)
		return struct {
			value int64
			found bool
		}{value, found}, err
	})
	if err != nil {
		log.Printf("Error getting counter %s: %v", name, err)
		return 0, false
	}
	return result.value, result.found
}

func (ds *DBStorage) GetGauge(name string) (float64, bool) {
	if name == "" {
		return 0, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() (struct {
		value float64
		found bool
	}, error) {
		value, found, err := ds.dr.GetGauge(name)
		return struct {
			value float64
			found bool
		}{value, found}, err
	})
	if err != nil {
		log.Printf("Error getting gauge %s: %v", name, err)
		return 0, false
	}
	return result.value, result.found
}

func (ds *DBStorage) GetMetric(name string) (*models.Metrics, bool) {
	if name == "" {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() (struct {
		metric *models.Metrics
		found  bool
	}, error) {
		metric, found, err := ds.dr.GetMetric(name)
		return struct {
			metric *models.Metrics
			found  bool
		}{metric, found}, err
	})
	if err != nil {
		log.Printf("Error getting metric %s: %v", name, err)
		return nil, false
	}
	return result.metric, result.found
}

func (ds *DBStorage) GetAllMetrics() []models.Metrics {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	metrics, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() ([]models.Metrics, error) {
		return ds.dr.GetAllMetrics()
	})
	if err != nil {
		log.Printf("Error getting all metrics: %v", err)
		return []models.Metrics{}
	}
	return metrics
}

func (ds *DBStorage) UpdateBatch(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := retry.ExecuteWithRetry(ctx, ds.retryConfig, func() (struct{}, error) {
		return struct{}{}, ds.dr.UpdateBatch(metrics)
	})
	return err
}
