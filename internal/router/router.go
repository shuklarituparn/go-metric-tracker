package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/middleware"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"go.uber.org/zap"
)

func NewRouter() *gin.Engine {
	storage := repository.NewMemStorage()
	metricsHandler := handler.NewMetricHandler(storage)
	debugHandler := handler.NewDebugHandler(storage)

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("err: problem starting logger")
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("Failed to sync logger: %v", syncErr)
		}
	}()

	router := gin.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	router.POST("/update", metricsHandler.UpdateMetricJSON)
	router.POST("/value", metricsHandler.GetMetricJSON)
	router.POST("/update/:type/:name/:value", metricsHandler.UpdateMetric)
	router.GET("/value/:type/:name", metricsHandler.GetMetric)
	router.GET("/", debugHandler.DebugHandler)
	router.GET("/debug", debugHandler.DebugHandler)

	router.POST("/update/:type/:name/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}
