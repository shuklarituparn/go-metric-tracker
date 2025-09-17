package agent

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type Collector interface {
	Start()
	Collect()
	GetMetrics() map[string]models.Metrics
}

type MetricCollector struct {
	metrics      map[string]models.Metrics
	mu           sync.RWMutex
	pollCount    int64
	pollInterval time.Duration
	ticker       *time.Ticker
}

func NewMetricCollector(pollInterval time.Duration) Collector {
	return &MetricCollector{
		metrics:      make(map[string]models.Metrics),
		pollCount:    0,
		pollInterval: pollInterval,
	}

}
func (mc *MetricCollector) Start() {
	mc.Collect()
	mc.ticker = time.NewTicker(mc.pollInterval)
	go func() {
		for range mc.ticker.C {
			mc.Collect()
		}
	}()
}

func CreateGuageMetric(id string, value float64) *models.Metrics {
	return &models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &value,
	}
}

func CreateCounterMetric(id string, value int64) *models.Metrics {
	return &models.Metrics{
		ID:    id,
		MType: models.Counter,
		Delta: &value,
	}
}

func (mc *MetricCollector) Collect() {
	var metricStats runtime.MemStats
	runtime.ReadMemStats(&metricStats)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.pollCount++

	gaugeMetrics := map[string]float64{
		"Alloc":         float64(metricStats.Alloc),
		"BuckHashSys":   float64(metricStats.BuckHashSys),
		"Frees":         float64(metricStats.Frees),
		"GCCPUFraction": metricStats.GCCPUFraction,
		"GCSys":         float64(metricStats.GCSys),
		"HeapAlloc":     float64(metricStats.HeapAlloc),
		"HeapIdle":      float64(metricStats.HeapIdle),
		"HeapInuse":     float64(metricStats.HeapInuse),
		"HeapObjects":   float64(metricStats.HeapObjects),
		"HeapReleased":  float64(metricStats.HeapReleased),
		"HeapSys":       float64(metricStats.HeapSys),
		"LastGC":        float64(metricStats.LastGC),
		"Lookups":       float64(metricStats.Lookups),
		"MCacheInuse":   float64(metricStats.MCacheInuse),
		"MCacheSys":     float64(metricStats.MCacheSys),
		"MSpanInuse":    float64(metricStats.MSpanInuse),
		"MSpanSys":      float64(metricStats.MSpanSys),
		"Mallocs":       float64(metricStats.Mallocs),
		"NextGC":        float64(metricStats.NextGC),
		"NumForcedGC":   float64(metricStats.NumForcedGC),
		"NumGC":         float64(metricStats.NumGC),
		"OtherSys":      float64(metricStats.OtherSys),
		"PauseTotalNs":  float64(metricStats.PauseTotalNs),
		"StackInuse":    float64(metricStats.StackInuse),
		"StackSys":      float64(metricStats.StackSys),
		"Sys":           float64(metricStats.Sys),
		"TotalAlloc":    float64(metricStats.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}

	for name, value := range gaugeMetrics {
		mc.metrics[name] = *CreateGuageMetric(name, value)
	}
	mc.metrics["PollCount"] = *CreateCounterMetric("PollCount", mc.pollCount)

}

func (mc *MetricCollector) GetMetrics() map[string]models.Metrics {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metricsCopy := make(map[string]models.Metrics)
	for k, v := range mc.metrics {
		metricsCopy[k] = v

	}
	return metricsCopy
}
