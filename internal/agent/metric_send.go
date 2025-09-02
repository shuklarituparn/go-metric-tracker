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
)

type Sender struct {
	client         *http.Client
	serverAddress  string
	reportInterval time.Duration
	storage        repository.Storage
	ticker         *time.Ticker
}

func NewSender(serverAddress string, reportInterval time.Duration, storage repository.Storage) *Sender {
	return &Sender{
		client:         &http.Client{},
		serverAddress:  serverAddress,
		reportInterval: reportInterval,
		storage:        storage,
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
	
	url := fmt.Sprintf("%s/update", s.serverAddress)
	
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metricJSON))
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