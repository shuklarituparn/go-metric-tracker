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
		return fmt.Errorf("err: failed to marshal metric: %w", err)

	}
	compressedData, err := utils.CompressedData(metricJSON)
	if err != nil {
		return fmt.Errorf("err: failed to compress data: %s", err.Error())
	}
	urlJSON := fmt.Sprintf("%s/update", s.serverAddress)
	jsonReq, err := http.NewRequest(http.MethodPost, urlJSON, bytes.NewBuffer(compressedData))
	if err != nil {
		return fmt.Errorf("err: failed to create request: %w", err)

	}
	jsonReq.Header.Set("Content-Type", "application/json")
	jsonReq.Header.Set("Content-Encoding", "gzip")
	jsonReq.Header.Set("Accept-Encoding", "gzip")
	Jsonresp, err := s.client.Do(jsonReq)

	if err != nil {
		return fmt.Errorf("err: failed to create request: %w", err)

	}

	defer func() {
		if err := Jsonresp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if Jsonresp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", Jsonresp.StatusCode)
	}

	return nil

}
