// internal/handler/metric_handler_test.go
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func TestMetricsHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		checkValue     func(storage *repository.MemStorage) bool
	}{
		{
			name:           "Correct gauge metric",
			method:         http.MethodPost,
			url:            "/update/gauge/temperature/30.7",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				v, ok := storage.GetGauge("temperature")
				return ok && v == 30.7
			},
		},
		{
			name:           "Correct counter metric",
			method:         http.MethodPost,
			url:            "/update/counter/test/100",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				v, ok := storage.GetCounter("test")
				return ok && v == 100
			},
		},
		{
			name:           "Counter accumulation",
			method:         http.MethodPost,
			url:            "/update/counter/accumulate/50",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				// This test assumes we send another request after the first
				v, ok := storage.GetCounter("accumulate")
				return ok && v == 50
			},
		},
		{
			name:           "Incorrect method type",
			method:         http.MethodGet,
			url:            "/update/gauge/temperature/31",
			expectedStatus: http.StatusMethodNotAllowed,
			checkValue: func(storage *repository.MemStorage) bool {
				return true
			},
		},
		{
			name:           "Very large gauge value",
			method:         http.MethodPost,
			url:            "/update/gauge/large/999999999.999999",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				v, ok := storage.GetGauge("large")
				return ok && v == 999999999.999999
			},
		},
		{
			name:           "Very large counter value",
			method:         http.MethodPost,
			url:            "/update/counter/bigcount/9999999999",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				v, ok := storage.GetCounter("bigcount")
				return ok && v == 9999999999
			},
		},
		{
			name:           "Invalid metric type",
			method:         http.MethodPost,
			url:            "/update/metric/test2/100",
			expectedStatus: http.StatusBadRequest,
			checkValue: func(storage *repository.MemStorage) bool {
				return true
			},
		},
		{
			name:           "Missing metric name",
			method:         http.MethodPost,
			url:            "/update/gauge//30.7",
			expectedStatus: http.StatusNotFound,
			checkValue: func(storage *repository.MemStorage) bool {
				return true
			},
		},
		{
			name:           "Missing metric value",
			method:         http.MethodPost,
			url:            "/update/gauge/temperature/",
			expectedStatus: http.StatusBadRequest,
			checkValue: func(storage *repository.MemStorage) bool {
				return true
			},
		},
		{
			name:           "Invalid gauge value (not a number)",
			method:         http.MethodPost,
			url:            "/update/gauge/temperature/invalid",
			expectedStatus: http.StatusBadRequest,
			checkValue: func(storage *repository.MemStorage) bool {
				_, ok := storage.GetGauge("temperature")
				return !ok
			},
		},
		{
			name:           "Invalid counter value (not an integer)",
			method:         http.MethodPost,
			url:            "/update/counter/requests/10.5",
			expectedStatus: http.StatusBadRequest,
			checkValue: func(storage *repository.MemStorage) bool {
				return true
			},
		},
		{
			name:           "Negative gauge value",
			method:         http.MethodPost,
			url:            "/update/gauge/balance/-15.5",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				v, ok := storage.GetGauge("balance")
				return ok && v == -15.5
			},
		},
		{
			name:           "Negative counter value",
			method:         http.MethodPost,
			url:            "/update/counter/debt/-50",
			expectedStatus: http.StatusOK,
			checkValue: func(storage *repository.MemStorage) bool {
				v, ok := storage.GetCounter("debt")
				return ok && v == -50
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := repository.NewMemStorage()
			handler := NewMetricHandler(storage)

			req := httptest.NewRequest(test.method, test.url, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			if !test.checkValue(storage) {
				t.Errorf("Value check failed for test: %s", test.name)
			}
		})
	}
}

func TestMetricsHandler_CounterAccumulation(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := NewMetricHandler(storage)

	req1 := httptest.NewRequest(http.MethodPost, "/update/counter/visits/10", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	v, ok := storage.GetCounter("visits")
	if !ok || v != 10 {
		t.Errorf("Expected counter value 10, got %v", v)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/update/counter/visits/25", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	v, ok = storage.GetCounter("visits")
	if !ok || v != 35 {
		t.Errorf("Expected counter value 35 after accumulation, got %v", v)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/update/counter/visits/5", nil)
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)

	v, ok = storage.GetCounter("visits")
	if !ok || v != 40 {
		t.Errorf("Expected counter value 40 after accumulation, got %v", v)
	}
}

func TestMetricsHandler_GaugeOverwrite(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := NewMetricHandler(storage)

	req1 := httptest.NewRequest(http.MethodPost, "/update/gauge/temp/20.5", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	v, ok := storage.GetGauge("temp")
	if !ok || v != 20.5 {
		t.Errorf("Expected gauge value 20.5, got %v", v)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/update/gauge/temp/30.5", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	v, ok = storage.GetGauge("temp")
	if !ok || v != 30.5 {
		t.Errorf("Expected gauge value 30.5 after overwrite, got %v", v)
	}
}
