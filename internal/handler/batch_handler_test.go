package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/audit"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler_BatchUpdateWithAudit(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if closeErr := tmpFile.Close(); closeErr != nil {
		t.Fatalf("Failed to close temp file: %v", closeErr)
	}

	auditManager := audit.NewAuditManager()
	fileAuditor, err := audit.NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}
	auditManager.AddObserver(fileAuditor)

	var storage repository.Storage = repository.NewMemStorage()
	handler := NewMetricHandlerWithAudit(storage, auditManager)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/updates/", handler.UpdateMetricsBatch)

	body := `[
		{"id":"Alloc","type":"gauge","value":123.45},
		{"id":"Frees","type":"counter","delta":10},
		{"id":"Temperature","type":"gauge","value":25.5}
	]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.0.42:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(100 * time.Millisecond)

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	var auditEvent audit.AuditEvent
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if err := json.Unmarshal([]byte(lines[0]), &auditEvent); err != nil {
		t.Fatalf("Failed to unmarshal audit event: %v", err)
	}

	if len(auditEvent.Metrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(auditEvent.Metrics))
	}

	expectedMetrics := map[string]bool{
		"Alloc":       false,
		"Frees":       false,
		"Temperature": false,
	}

	for _, metric := range auditEvent.Metrics {
		if _, exists := expectedMetrics[metric]; exists {
			expectedMetrics[metric] = true
		}
	}

	for metric, found := range expectedMetrics {
		if !found {
			t.Errorf("Expected to find metric '%s' in audit", metric)
		}
	}
}

func TestMetricsHandler_BatchUpdateMultipleAuditors(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if closeErr := tmpFile.Close(); closeErr != nil {
		t.Fatalf("Failed to close temp file: %v", closeErr)
	}

	var httpReceived audit.AuditEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if unmarshalErr := json.Unmarshal(body, &httpReceived); unmarshalErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	auditManager := audit.NewAuditManager()

	fileAuditor, err := audit.NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}
	auditManager.AddObserver(fileAuditor)

	httpAuditor, err := audit.NewHTTPAuditor(server.URL)
	if err != nil {
		t.Fatalf("Failed to create HTTP auditor: %v", err)
	}
	auditManager.AddObserver(httpAuditor)

	var storage repository.Storage = repository.NewMemStorage()
	handler := NewMetricHandlerWithAudit(storage, auditManager)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/updates/", handler.UpdateMetricsBatch)

	body := `[{"id":"metric1","type":"gauge","value":100.5}]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected audit file to contain data")
	}

	if len(httpReceived.Metrics) != 1 || httpReceived.Metrics[0] != "metric1" {
		t.Errorf("Expected HTTP auditor to receive metric1, got %v", httpReceived.Metrics)
	}
}