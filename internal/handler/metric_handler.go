package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

type MetricsHandler struct {
	storage repository.Storage
}
type DebugHandler struct {
	storage repository.Storage
}

func NewMetricHandler(storage repository.Storage) *MetricsHandler {
	return &MetricsHandler{
		storage: storage,
	}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestURL := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(requestURL, "/")
	log.Printf("Path parts: %v (length: %d)", parts, len(parts))
	if len(parts) < 3 || parts[0] != "update" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Metric name required", http.StatusNotFound)
		return
	}
	if len(parts) < 4 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	typeOfMetric := parts[1]
	metricName := parts[2]
	valueMetric := parts[3]

	switch typeOfMetric {
	case models.Gauge:
		value, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}

		if err := h.storage.UpdateGauge(metricName, value); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
	case models.Counter:
		value, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}

		if err := h.storage.UpdateCounter(metricName, value); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

}

func NewDebugHandler(storage repository.Storage) *DebugHandler {
	return &DebugHandler{
		storage: storage,
	}
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := h.storage.GetAllMetrics()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if len(metrics) == 0 {
		if _, err := fmt.Fprintln(w, "No metrics stored yet"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
		return
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if _, err := fmt.Fprintf(w, "%s (gauge): %v\n", metric.ID, *metric.Value); err != nil {
				log.Printf("failed to write response: %v", err)
			}
		case models.Counter:
			if _, err := fmt.Fprintf(w, "%s (counter): %v\n", metric.ID, *metric.Delta); err != nil {
				log.Printf("failed to write response: %v", err)
			}
		}
	}
}
