package agent

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

type MetricCollector struct {
	metrics      map[string]models.Metrics
	mu           sync.RWMutex
	pollCount    int64
	pollInterval time.Duration
	storage      *repository.MemStorage
}

func NewMetricCollector(pollInterval time.Duration, storage *repository.MemStorage) *MetricCollector {
	return &MetricCollector{
		metrics:      make(map[string]models.Metrics),
		pollCount:    0,
		pollInterval: pollInterval,
		storage:      storage,
	}

}

func (mc *MetricCollector) Start() {
	ticker := time.NewTicker(mc.pollInterval)
	go func() {
		for range ticker.C {
			mc.Collect()
			if err := mc.SendToStorage(); err != nil {
				log.Printf("error: problem sending metric in metric collector")
			}
		}
	}()
	mc.Collect()
}

func (mc *MetricCollector) Collect() {
	var metricStats runtime.MemStats
	runtime.ReadMemStats(&metricStats)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.pollCount++

	allocValue := float64(metricStats.Alloc)
	mc.metrics["Alloc"] = models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &allocValue,
	}

	buckHashValue := float64(metricStats.BuckHashSys)
	mc.metrics["BuckHashSys"] = models.Metrics{
		ID:    "BuckHashSys",
		MType: "gauge",
		Value: &buckHashValue,
	}
	FreesVal := float64(metricStats.Frees)
	mc.metrics["Frees"] = models.Metrics{
		ID:    "Frees",
		MType: "gauge",
		Value: &FreesVal,
	}
	gpuFractionValue := float64(metricStats.GCCPUFraction)
	mc.metrics["GCCPUFraction"] = models.Metrics{
		ID:    "GCCPUFraction",
		MType: "gauge",
		Value: &gpuFractionValue,
	}
	gSysVal := float64(metricStats.GCSys)
	mc.metrics["GCSys"] = models.Metrics{
		ID:    "GCSys",
		MType: "gauge",
		Value: &gSysVal,
	}
	heapAllocValue := float64(metricStats.HeapAlloc)
	mc.metrics["HeapAlloc"] = models.Metrics{
		ID:    "HeapAlloc",
		MType: "gauge",
		Value: &heapAllocValue,
	}
	heapIdleValue := float64(metricStats.HeapIdle)
	mc.metrics["HeapIdle"] = models.Metrics{
		ID:    "HeapIdle",
		MType: "gauge",
		Value: &heapIdleValue,
	}
	heapInUseVal := float64(metricStats.HeapInuse)
	mc.metrics["HeapInuse"] = models.Metrics{
		ID:    "HeapInuse",
		MType: "gauge",
		Value: &heapInUseVal,
	}
	heapObjectVal := float64(metricStats.HeapObjects)
	mc.metrics["HeapObjects"] = models.Metrics{
		ID:    "HeapObjects",
		MType: "gauge",
		Value: &heapObjectVal,
	}
	heapReleasedVal := float64(metricStats.HeapReleased)
	mc.metrics["HeapReleased"] = models.Metrics{
		ID:    "HeapReleased",
		MType: "gauge",
		Value: &heapReleasedVal,
	}
	heapSysVal := float64(metricStats.HeapSys)
	mc.metrics["HeapSys"] = models.Metrics{
		ID:    "HeapSys",
		MType: "gauge",
		Value: &heapSysVal,
	}
	LastGCVal := float64(metricStats.LastGC)
	mc.metrics["LastGC"] = models.Metrics{
		ID:    "LastGC",
		MType: "gauge",
		Value: &LastGCVal,
	}
	LookupsVal := float64(metricStats.Lookups)
	mc.metrics["Lookups"] = models.Metrics{
		ID:    "Lookups",
		MType: "gauge",
		Value: &LookupsVal,
	}
	MCacheInuseVal := float64(metricStats.MCacheInuse)
	mc.metrics["MCacheInuse"] = models.Metrics{
		ID:    "MCacheInuse",
		MType: "gauge",
		Value: &MCacheInuseVal,
	}
	MCacheSysVal := float64(metricStats.MCacheSys)
	mc.metrics["MCacheSys"] = models.Metrics{
		ID:    "MCacheSys",
		MType: "gauge",
		Value: &MCacheSysVal,
	}
	MSpanInuseVal := float64(metricStats.MSpanInuse)
	mc.metrics["MSpanInuse"] = models.Metrics{
		ID:    "MSpanInuse",
		MType: "gauge",
		Value: &MSpanInuseVal,
	}
	MSpanSysVal := float64(metricStats.MSpanSys)
	mc.metrics["MSpanSys"] = models.Metrics{
		ID:    "MSpanSys",
		MType: "gauge",
		Value: &MSpanSysVal,
	}
	MallocsVal := float64(metricStats.Mallocs)
	mc.metrics["Mallocs"] = models.Metrics{
		ID:    "Mallocs",
		MType: "gauge",
		Value: &MallocsVal,
	}
	NextGCVal := float64(metricStats.NextGC)
	mc.metrics["NextGC"] = models.Metrics{
		ID:    "NextGC",
		MType: "gauge",
		Value: &NextGCVal,
	}
	NumForcedGCVal := float64(metricStats.NumForcedGC)
	mc.metrics["NumForcedGC"] = models.Metrics{
		ID:    "NumForcedGC",
		MType: "gauge",
		Value: &NumForcedGCVal,
	}
	NumGCVal := float64(metricStats.NumGC)
	mc.metrics["NumGC"] = models.Metrics{
		ID:    "NumGC",
		MType: "gauge",
		Value: &NumGCVal,
	}
	OtherSysVal := float64(metricStats.OtherSys)
	mc.metrics["OtherSys"] = models.Metrics{
		ID:    "OtherSys",
		MType: "gauge",
		Value: &OtherSysVal,
	}
	PauseTotalNsVal := float64(metricStats.PauseTotalNs)
	mc.metrics["PauseTotalNs"] = models.Metrics{
		ID:    "PauseTotalNs",
		MType: "gauge",
		Value: &PauseTotalNsVal,
	}
	StackInuseVal := float64(metricStats.StackInuse)
	mc.metrics["StackInuse"] = models.Metrics{
		ID:    "StackInuse",
		MType: "gauge",
		Value: &StackInuseVal,
	}
	StackSysVal := float64(metricStats.StackSys)
	mc.metrics["StackSys"] = models.Metrics{
		ID:    "StackSys",
		MType: "gauge",
		Value: &StackSysVal,
	}
	SysVal := float64(metricStats.Sys)
	mc.metrics["Sys"] = models.Metrics{
		ID:    "Sys",
		MType: "gauge",
		Value: &SysVal,
	}
	TotalAllocVal := float64(metricStats.TotalAlloc)
	mc.metrics["TotalAlloc"] = models.Metrics{
		ID:    "TotalAlloc",
		MType: "gauge",
		Value: &TotalAllocVal,
	}
	pollCountDelta := mc.pollCount
	mc.metrics["PollCount"] = models.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &pollCountDelta,
	}
	randomVal := rand.Float64()
	mc.metrics["RandomValue"] = models.Metrics{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &randomVal,
	}
}

func (mc *MetricCollector) SendToStorage() error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, metric := range mc.metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				if err := mc.storage.UpdateGauge(metric.ID, *metric.Value); err != nil {
					return fmt.Errorf("error: updating gauge metric in collector")
				}
			}
		case models.Counter:
			if metric.Delta != nil {
				if err := mc.storage.UpdateCounter(metric.ID, *metric.Delta); err != nil {
					return fmt.Errorf("error: updating counter metric in collector")

				}
			}
		}
	}
	return nil
}
