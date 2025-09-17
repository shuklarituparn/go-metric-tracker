package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

type MetricsHandler struct {
	storage repository.Storage
}

func NewMetricHandler(storage repository.Storage) *MetricsHandler {
	return &MetricsHandler{
		storage: storage,
	}
}

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
		c.JSON(http.StatusOK, metric)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})
		return
	}
}

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
