// Package handler предоставляет HTTP обработчики для работы с метриками.
package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/audit"
	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

// MetricsHandler обрабатывает HTTP запросы для работы с метриками.
type MetricsHandler struct {
	storage      repository.Storage
	auditManager *audit.AuditManager
}

// NewMetricHandler создает новый экземпляр обработчика метрик.
// Принимает storage - хранилище метрик, реализующее интерфейс repository.Storage.
func NewMetricHandler(storage repository.Storage) *MetricsHandler {
	return &MetricsHandler{
		storage:      storage,
		auditManager: nil,
	}
}

func NewMetricHandlerWithAudit(storage repository.Storage, auditManager *audit.AuditManager) *MetricsHandler {
	return &MetricsHandler{
		storage:      storage,
		auditManager: auditManager,
	}
}
func (h *MetricsHandler) sendAudit(c *gin.Context, metricNames []string) {
	if h.auditManager != nil && h.auditManager.IsEnabled() {
		event := audit.AuditEvent{
			Timestamp: time.Now().Unix(),
			Metrics:   metricNames,
			IPAddress: c.ClientIP(),
		}
		h.auditManager.Notify(event)
	}
}

// UpdateMetric обрабатывает запрос на обновление метрики через URL параметры.
// Маршрут: POST /update/:type/:name/:value
// Параметры:
//   - type: тип метрики (gauge или counter)
//   - name: имя метрики
//   - value: значение метрики
//
// Возвращает HTTP 200 при успехе, HTTP 400 при некорректных данных,
// HTTP 404 если не указано имя метрики, HTTP 500 при ошибке сервера.
func (h *MetricsHandler) UpdateMetric(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	log.Printf("Received request: %s %s", c.Request.Method, c.Request.URL.Path)
	log.Printf("Path params: type=%s, name=%s, value=%s", metricType, metricName, metricValue)

	if metricName == "" {
		c.String(http.StatusNotFound, "Metric name required")
		return
	}
	if metricValue == "" {
		c.String(http.StatusBadRequest, "Metric value required")
		return
	}

	switch metricType {
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid gauge value")
			return
		}

		if err := h.storage.UpdateGauge(metricName, value); err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		c.String(http.StatusOK, "")
	case models.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid counter value")
			return
		}

		if err := h.storage.UpdateCounter(metricName, value); err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		c.String(http.StatusOK, "")
	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
		return
	}
}

// GetMetric обрабатывает запрос на получение значения метрики через URL параметры.
// Маршрут: GET /value/:type/:name
// Параметры:
//   - type: тип метрики (gauge или counter)
//   - name: имя метрики
//
// Возвращает значение метрики в виде строки при успехе (HTTP 200),
// HTTP 404 если метрика не найдена, HTTP 400 при некорректном типе.
func (h *MetricsHandler) GetMetric(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	log.Printf("Getting metric: type=%s, name=%s", metricType, metricName)

	if metricName == "" {
		c.String(http.StatusNotFound, "Metric name required")
		return
	}

	switch metricType {
	case models.Gauge:
		value, exists := h.storage.GetGauge(metricName)
		if !exists {
			c.String(http.StatusNotFound, "Metric not found")
			return
		}
		c.String(http.StatusOK, strconv.FormatFloat(value, 'g', -1, 64))
	case models.Counter:
		value, exists := h.storage.GetCounter(metricName)
		if !exists {
			c.String(http.StatusNotFound, "Metric not found")
			return
		}
		c.String(http.StatusOK, strconv.FormatInt(value, 10))
	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
		return
	}
}

// UpdateMetricJSON обрабатывает запрос на обновление метрики в формате JSON.
// Маршрут: POST /update/
// Ожидает JSON тело запроса с полями:
//   - id: имя метрики
//   - type: тип метрики (gauge или counter)
//   - value: значение для gauge метрики
//   - delta: значение для counter метрики
//
// Возвращает обновленную метрику в формате JSON при успехе (HTTP 200),
// HTTP 400 при некорректных данных, HTTP 500 при ошибке сервера.
func (h *MetricsHandler) UpdateMetricJSON(c *gin.Context) {
	var metric models.Metrics

	if err := c.ShouldBindJSON(&metric); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if metric.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric id is required"})
		return
	}

	if metric.MType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric type is required"})
		return
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gauge value is required"})
			return
		}
		if err := h.storage.UpdateGauge(metric.ID, *metric.Value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.sendAudit(c, []string{metric.ID})
		c.JSON(http.StatusOK, metric)

	case models.Counter:
		if metric.Delta == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "counter delta is required"})
			return
		}
		if err := h.storage.UpdateCounter(metric.ID, *metric.Delta); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		newValue, _ := h.storage.GetCounter(metric.ID)
		metric.Delta = &newValue
		h.sendAudit(c, []string{metric.ID})
		c.JSON(http.StatusOK, metric)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}
}

// GetMetricJSON обрабатывает запрос на получение метрики в формате JSON.
// Маршрут: POST /value/
// Ожидает JSON тело запроса с полями:
//   - id: имя метрики
//
// Возвращает метрику в формате JSON при успехе (HTTP 200),
// HTTP 404 если метрика не найдена, HTTP 400 при некорректном JSON.
func (h *MetricsHandler) GetMetricJSON(c *gin.Context) {
	var metric models.Metrics

	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if metric.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric id is required"})
		return
	}

	storedMetric, exists := h.storage.GetMetric(metric.ID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
		return
	}

	c.JSON(http.StatusOK, storedMetric)
}

// DBHandler обрабатывает запрос проверки подключения к базе данных.
// Маршрут: GET /ping
// Возвращает HTTP 200 если подключение к БД успешно, HTTP 500 при ошибке.
func (h *MetricsHandler) DBHandler(c *gin.Context) {
	if dbStorage, ok := h.storage.(*repository.DBStorage); ok {
		if err := dbStorage.Ping(); err != nil {
			log.Printf("database ping failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.Status(http.StatusInternalServerError)
}
