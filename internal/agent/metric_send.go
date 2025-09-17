package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"github.com/shuklarituparn/go-metric-tracker/internal/utils"
)

type StorageSender interface {
	Start()
	SendToStorage() error
}
type ServerSender interface {
	Start()
	SendMetrics()
	SendMetricsBatch() error
}

type MetricStorageSender struct {
	collector Collector
	storage   repository.Storage
	ticker    *time.Ticker
	interval  time.Duration
}

type Sender struct {
	client         *http.Client
	serverAddress  string
	reportInterval time.Duration
	storage        repository.Storage
	ticker         *time.Ticker
	useBatch       bool
}

func NewStorageSender(collector Collector, storage repository.Storage, interval time.Duration) StorageSender {
	return &MetricStorageSender{
		collector: collector,
		storage:   storage,
		interval:  interval,
	}
}

func (mss *MetricStorageSender) Start() {
	mss.ticker = time.NewTicker(mss.interval)

	go func() {
		if err := mss.SendToStorage(); err != nil {
			log.Printf("problem sending metric to storage: %v", err)
		}
		for range mss.ticker.C {
			if err := mss.SendToStorage(); err != nil {
				log.Printf("error sending metrics to storage: %v", err)
			}
		}
	}()

}

func (mss *MetricStorageSender) SendToStorage() error {

	metrics := mss.collector.GetMetrics()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				if err := mss.storage.UpdateGauge(metric.ID, *metric.Value); err != nil {
					return fmt.Errorf("error updating gauge metric %s: %w", metric.ID, err)
				}
			}
		case models.Counter:
			if metric.Delta != nil {
				if err := mss.storage.UpdateCounter(metric.ID, *metric.Delta); err != nil {
					return fmt.Errorf("error updating counter metric %s: %w", metric.ID, err)
				}
			}
		}
	}
	return nil

}

func NewSender(serverAddress string, reportInterval time.Duration, storage repository.Storage) ServerSender {
	return &Sender{
		client:         &http.Client{},
		serverAddress:  serverAddress,
		reportInterval: reportInterval,
		storage:        storage,
		useBatch:       true,
	}
}

func (s *Sender) Start() {
	s.ticker = time.NewTicker(s.reportInterval)
	go func() {
		for range s.ticker.C {
			s.SendMetrics()
		}
	}()
}

func (s *Sender) SendMetrics() {
	metrics := s.storage.GetAllMetrics()
	if len(metrics) == 0 {
		log.Println("No metrics to send")
		return
	}

	for _, metric := range metrics {
		if err := s.SendMetric(metric); err != nil {
			log.Printf("Failed to send metric %s: %v", metric.ID, err)
		}
	}
	log.Printf("Sent %d metrics to server", len(metrics))
}

func (s *Sender) SendMetric(metric models.Metrics) error {
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	compressedData, err := utils.CompressedData(metricJSON)
	if err != nil {
		return fmt.Errorf("err: failed to compress data: %s", err.Error())
	}
	url := fmt.Sprintf("%s/update/", s.serverAddress)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *Sender) SendMetricsBatch() error {
	metrics := s.storage.GetAllMetrics()

	if len(metrics) == 0 {
		return fmt.Errorf("error: there are no metrics")

	}
	metricJSON, err := json.Marshal(metrics)
	if err != nil {
		s.SendMetrics()
		return fmt.Errorf("problem marshalling metrics: %v", err)
	}
	compressedData, err := utils.CompressedData(metricJSON)
	if err != nil {
		if err := s.sendBatchUncompressed(metricJSON); err != nil {
			return fmt.Errorf("error: problem sending uncompressed metrics: %v", err)
		}
		return fmt.Errorf("error: problem compressing the metrics: %v", err)
	}
	url := fmt.Sprintf("%s/updates/", s.serverAddress)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if err != nil {
		return fmt.Errorf("error: problem sending the metrics:%v", err)
	}
	resp, err := s.client.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: problem with server code %v", err)
	}

	log.Printf("Successfully sent %d metrics in batch", len(metrics))
	return nil
}

func (s *Sender) sendBatchUncompressed(data []byte) error {
	url := fmt.Sprintf("%s/updates/", s.serverAddress)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}
