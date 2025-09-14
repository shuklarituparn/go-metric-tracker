package router

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/middleware"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"go.uber.org/zap"
)

type StorageCloser interface {
	repository.Storage
	Close()
}

func NewRouterWithFS(cfg *config.ServerConfig) *gin.Engine {
	router, cleanup := NewRouterWithStorage(cfg)
	if cleanup != nil {
		go func() {
			signChan := make(chan os.Signal, 1)
			signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)
			<-signChan
			cleanup()
		}()
	}
	return router
}

func NewRouterWithStorage(cfg *config.ServerConfig) (*gin.Engine, func()) {
	var storage repository.Storage
	var cleanup func()

	if cfg.DBConfig.DSN != "" {
		db, err := cfg.DBConfig.Connect()
		if err != nil {
			log.Printf("Failed to connect to database, falling back to file storage: %v", err)
			fileStorage := repository.NewFileStorage(cfg.FileStoragePath, cfg.StoreIntervalDuration, cfg.Restore)
			storage = fileStorage

			if cfg.StoreIntervalDuration > 0 {
				fileStorage.AutoSave()
			}
			cleanup = func() {
				log.Println("closing file storage")
				if err := fileStorage.Close(); err != nil {
					log.Printf("Error closing file storage: %v", err)
				}
			}
		} else {
			dbStorage, err := repository.NewDBStorage(db)
			if err != nil {
				log.Printf("failed to initialize db storage: %v", err)
				if err := db.Close(); err != nil {
					log.Printf("error: problem closing db: %v", err)
				}
				fileStorage := repository.NewFileStorage(cfg.FileStoragePath, cfg.StoreIntervalDuration, cfg.Restore)
				storage = fileStorage

				cleanup = func() {
					if err := fileStorage.Close(); err != nil {
						log.Printf("Error closing file storage: %v", err)
					}
				}
			} else {
				log.Printf("Using database storage")
				storage = dbStorage
				cleanup = func() {
					log.Println("Closing database storage...")
					if err := dbStorage.Close(); err != nil {
						log.Printf("Error closing database storage: %v", err)
					}
				}

			}
		}
	} else {
		fileStorage := repository.NewFileStorage(cfg.FileStoragePath, cfg.StoreIntervalDuration, cfg.Restore)
		storage = fileStorage

		if cfg.StoreIntervalDuration > 0 {
			fileStorage.AutoSave()
		}

		cleanup = func() {
			log.Println("Closing file storage...")
			if err := fileStorage.Close(); err != nil {
				log.Printf("Error closing file storage: %v", err)
			}
		}
	}

	router := CreateRouter(storage)
	return router, cleanup
}
func CreateRouter(storage repository.Storage) *gin.Engine {
	metricsHandler := handler.NewMetricHandler(storage)
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
