package agent

import (
	"fmt"
	"log"
	"net/http"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type Sender struct {
	client          *http.Client
	serverAddress   string
	reportInterval  time.Duration
	metricCollector *MetricCollector
}

func NewSender(serverAddress string, reportInterval time.Duration, collector *MetricCollector) *Sender {
	return &Sender{
		client: &http.Client{},
		serverAddress:   serverAddress,
		reportInterval:  reportInterval,
		metricCollector: collector,
	}
}

func (s *Sender) Start() {
	reportAfter := time.NewTicker(s.reportInterval)
	go func() {
		for range reportAfter.C {
			s.SendMetrics()
		}
	}()
}
func (s *Sender) SendMetrics() {
	metrics := s.metricCollector.storage.GetAllMetrics()
	for _, metric := range metrics {
		if err := s.SendMetric(metric); err != nil {
			log.Printf("Failed to send metric %s: %v", metric.ID, err)
		}
	}
	log.Printf("Sent %d metrics to server", len(metrics))
}

func (s *Sender) SendMetric(metric models.Metrics) error {
	var url string

	switch metric.MType {
	case models.Gauge:
		url = fmt.Sprintf("%s/update/gauge/%s/%g", s.serverAddress, metric.ID, *metric.Value)
	case models.Counter:
		url = fmt.Sprintf("%s/update/counter/%s/%d", s.serverAddress, metric.ID, *metric.Delta)
	default:
		return fmt.Errorf("unknown metric type: %v", metric.MType)

	}
	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		return fmt.Errorf("err: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("err: failed to create request: %w", err)

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
