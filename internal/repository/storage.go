package repository

import (
	"fmt"
	"sync"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetCounter(name string) (int64, bool)
	GetGauge(name string) (float64, bool)
	GetAllMetrics() []models.Metrics
	GetMetric(name string) (*models.Metrics, bool)
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	if name == "" {
		return fmt.Errorf("error: gauge metric name cannot be empty")
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = value
	return nil
}

func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	if name == "" {
		return fmt.Errorf("error: counter metric name cannot be empty")
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += value
	return nil
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	if name == "" {
		return 0, false
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.counters[name]
	return value, ok
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	if name == "" {
		return 0, false
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.gauges[name]
	return value, ok
}

func (ms *MemStorage) GetMetric(name string) (*models.Metrics, bool) {
	if name == "" {
		return nil, false
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if value, ok := ms.gauges[name]; ok {
		return &models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}, true
	}

	if value, ok := ms.counters[name]; ok {
		return &models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &value,
		}, true
	}
	return nil, false
}

func (ms *MemStorage) GetAllMetrics() []models.Metrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	var metric []models.Metrics

	for name, value := range ms.gauges {
		v := value
		metric = append(metric, models.Metrics{
			ID:    name,
			Value: &v,
			MType: models.Gauge,
		})
	}

	for name, value := range ms.counters {
		delta := value
		metric = append(metric, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}
	return metric
}
