package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type BatchUpdater interface {
	UpdateBatch([]models.Metrics) error
}

func (h *MetricsHandler) UpdateMetricsBatch(c *gin.Context) {
	var metrics []models.Metrics
	if err := c.ShouldBindJSON(&metrics); err != nil {
		log.Printf("Failed to bind json: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	for i, metric := range metrics {
		if metric.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "metric id is required", "index": i})
			return
		}
		if metric.MType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "metric type is required", "index": i})
			return
		}

		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "metric value is required"})
				return
			}
		case models.Counter:
			if metric.Delta == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "counter "})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric type"})

		}
	}
	if batchUpdater, ok := h.storage.(BatchUpdater); ok {
		if err := batchUpdater.UpdateBatch(metrics); err != nil {
			log.Printf("Failed to update batch: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update metrics"})
			return
		}
	} else {
		for _, metric := range metrics {
			switch metric.MType {
			case models.Gauge:
				if err := h.storage.UpdateGauge(metric.ID, *metric.Value); err != nil {
					log.Printf("Failed to update gauge %s: %v", metric.ID, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			case models.Counter:
				if err := h.storage.UpdateCounter(metric.ID, *metric.Delta); err != nil {
					log.Printf("Failed to update counter %s: %v", metric.ID, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
	}

	var response []models.Metrics
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			response = append(response, metric)
		case models.Counter:
			if newValue, ok := h.storage.GetCounter(metric.ID); ok {
				metric.Delta = &newValue
				response = append(response, metric)
			} else {
				response = append(response, metric)
			}
		}
	}

	c.JSON(http.StatusOK, response)

}
