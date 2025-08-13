package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks,omitempty"`
}

type HealthHandler struct {
	storage   repository.Storage
	startTime time.Time
}

func NewHealthHandler(storage repository.Storage) *HealthHandler {
	return &HealthHandler{
		storage:   storage,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(h.startTime)

	storageHealth := "healthy"
	if h.storage == nil {
		storageHealth = "unhealthy"
	} else {
		metrics := h.storage.GetAllMetrics()
		if metrics == nil {
			storageHealth = "degraded"
		}
	}

	overallStatus := "healthy"
	if storageHealth != "healthy" {
		overallStatus = "unhealthy"
	}

	status := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    uptime.String(),
		Checks: map[string]string{
			"storage": storageHealth,
		},
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}
