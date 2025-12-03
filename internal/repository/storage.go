// Package repository предоставляет интерфейсы и реализации для хранения метрик.
package repository

import (
	"fmt"
	"sync"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

// MetricReader определяет интерфейс для чтения метрик из хранилища.
type MetricReader interface {
	// GetCounter возвращает значение метрики типа counter и флаг ее существования.
	GetCounter(name string) (int64, bool)
	// GetGauge возвращает значение метрики типа gauge и флаг ее существования.
	GetGauge(name string) (float64, bool)
	// GetAllMetrics возвращает все метрики из хранилища.
	GetAllMetrics() []models.Metrics
	// GetMetric возвращает метрику по имени и флаг ее существования.
	GetMetric(name string) (*models.Metrics, bool)
}

// MetricWriter определяет интерфейс для записи метрик в хранилище.
type MetricWriter interface {
	// UpdateGauge обновляет значение метрики типа gauge.
	UpdateGauge(name string, value float64) error
	// UpdateCounter увеличивает значение метрики типа counter на указанную величину.
	UpdateCounter(name string, value int64) error
}

// Storage объединяет интерфейсы для чтения и записи метрик.
type Storage interface {
	MetricReader
	MetricWriter
}

// MemStorage реализует хранилище метрик в оперативной памяти.
type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage создает новое хранилище метрик в памяти.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge обновляет значение метрики типа gauge в памяти.
// Возвращает ошибку, если имя метрики пустое.
func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	if name == "" {
		return fmt.Errorf("error: gauge metric name cannot be empty")
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = value
	return nil
}

// UpdateCounter увеличивает значение метрики типа counter на указанную величину.
// Возвращает ошибку, если имя метрики пустое.
func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	if name == "" {
		return fmt.Errorf("error: counter metric name cannot be empty")
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += value
	return nil
}

// GetCounter возвращает значение метрики типа counter и флаг ее существования.
// Возвращает 0 и false, если имя метрики пустое или метрика не найдена.
func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	if name == "" {
		return 0, false
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.counters[name]
	return value, ok
}

// GetGauge возвращает значение метрики типа gauge и флаг ее существования.
// Возвращает 0 и false, если имя метрики пустое или метрика не найдена.
func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	if name == "" {
		return 0, false
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.gauges[name]
	return value, ok
}

// GetMetric возвращает метрику по имени (gauge или counter) и флаг ее существования.
// Возвращает nil и false, если имя метрики пустое или метрика не найдена.
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

// GetAllMetrics возвращает все метрики (gauge и counter) из хранилища.
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
