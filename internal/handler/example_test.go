package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

// Example демонстрирует базовое использование обработчиков метрик.
// Показывает, как создать обработчик и обновить метрику типа gauge.
func Example() {
	// Создаем хранилище в памяти
	storage := repository.NewMemStorage()

	// Создаем обработчик метрик
	handler := handler.NewMetricHandler(storage)

	// Настраиваем Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/update/:type/:name/:value", handler.UpdateMetric)

	// Создаем тестовый запрос для обновления gauge метрики
	req := httptest.NewRequest("POST", "/update/gauge/temperature/23.5", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// ExampleMetricsHandler_UpdateMetricJSON демонстрирует обновление метрики через JSON API.
// Показывает, как отправить метрику типа gauge в формате JSON.
func ExampleMetricsHandler_UpdateMetricJSON() {
	storage := repository.NewMemStorage()
	handler := handler.NewMetricHandler(storage)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/update/", handler.UpdateMetricJSON)

	// Создаем метрику gauge
	value := 42.5
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: models.Gauge,
		Value: &value,
	}

	body, _ := json.Marshal(metric)
	req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// ExampleMetricsHandler_UpdateMetricJSON_counter демонстрирует обновление counter метрики.
// Показывает, как отправить метрику типа counter в формате JSON.
func ExampleMetricsHandler_UpdateMetricJSON_counter() {
	storage := repository.NewMemStorage()
	handler := handler.NewMetricHandler(storage)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/update/", handler.UpdateMetricJSON)

	// Создаем метрику counter
	delta := int64(10)
	metric := models.Metrics{
		ID:    "request_count",
		MType: models.Counter,
		Delta: &delta,
	}

	body, _ := json.Marshal(metric)
	req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// ExampleMetricsHandler_GetMetricJSON демонстрирует получение метрики через JSON API.
// Показывает, как запросить значение сохраненной метрики.
func ExampleMetricsHandler_GetMetricJSON() {
	storage := repository.NewMemStorage()
	handler := handler.NewMetricHandler(storage)

	// Сначала сохраняем метрику
	value := 100.0
	storage.UpdateGauge("memory_usage", value)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/value/", handler.GetMetricJSON)

	// Создаем запрос для получения метрики
	metric := models.Metrics{
		ID:    "memory_usage",
		MType: models.Gauge,
	}

	body, _ := json.Marshal(metric)
	req := httptest.NewRequest("POST", "/value/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var response models.Metrics
	json.Unmarshal(w.Body.Bytes(), &response)

	fmt.Printf("Status: %d, Value: %.0f\n", w.Code, *response.Value)
	// Output:
	// Status: 200, Value: 100
}

// ExampleMetricsHandler_UpdateMetricsBatch демонстрирует пакетное обновление метрик.
// Показывает, как отправить несколько метрик одним запросом.
func ExampleMetricsHandler_UpdateMetricsBatch() {
	storage := repository.NewMemStorage()
	handler := handler.NewMetricHandler(storage)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/updates/", handler.UpdateMetricsBatch)

	// Создаем массив метрик
	gaugeValue := 75.5
	counterDelta := int64(5)

	metrics := []models.Metrics{
		{
			ID:    "cpu",
			MType: models.Gauge,
			Value: &gaugeValue,
		},
		{
			ID:    "requests",
			MType: models.Counter,
			Delta: &counterDelta,
		},
	}

	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest("POST", "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// ExampleNewMetricHandler демонстрирует создание нового обработчика метрик.
// Показывает базовую инициализацию обработчика с хранилищем.
func ExampleNewMetricHandler() {
	// Создаем хранилище
	storage := repository.NewMemStorage()

	// Создаем обработчик
	handler := handler.NewMetricHandler(storage)

	fmt.Printf("Handler created: %T\n", handler)
	// Output:
	// Handler created: *handler.MetricsHandler
}
