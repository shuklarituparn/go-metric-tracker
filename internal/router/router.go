package router

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/middleware"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"go.uber.org/zap"
)

func NewRouterWithFS(cfg *config.ServerConfig) *gin.Engine {
	storage := repository.NewFileStorage(cfg.FileStoragePath, cfg.StoreIntervalDuration, cfg.Restore)
	router := CreateRouter(storage, cfg)
	if cfg.StoreIntervalDuration > 0 {
		storage.AutoSave()
	}

	return router
}

func CreateRouter(storage repository.Storage, cfg *config.ServerConfig) *gin.Engine {
	var db *sql.DB

	if cfg.DatabaseDSN != "" && cfg.DBConfig.DSN != "" {
		var err error
		db, err = cfg.DBConfig.Connect()
		if err != nil {
			log.Printf("warning: failed to connect DB: %v (continuing without DB)", err)
		}
	}
	metricsHandler := handler.NewMetricHandler(storage, db)
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
	router.Use(middleware.CompressionMiddleware())
	router.Use(middleware.DecompressionMiddleware())

	router.POST("/update/", metricsHandler.UpdateMetricJSON)
	router.POST("/value/", metricsHandler.GetMetricJSON)
	router.POST("/update/:type/:name/:value", metricsHandler.UpdateMetric)
	router.GET("/value/:type/:name", metricsHandler.GetMetric)
	router.GET("/", handler.DefaultHandle)
	router.GET("/ping", metricsHandler.DBHandler)
	router.POST("/update/:type/:name/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}
