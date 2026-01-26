// Package router предоставляет настройку маршрутизации HTTP-сервера метрик.
package router

import (
	"crypto/rsa"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/audit"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/middleware"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"github.com/shuklarituparn/go-metric-tracker/internal/utils"
	"go.uber.org/zap"
)

type StorageCloser interface {
	repository.Storage
	Close()
}

// NewRouterWithFS создает новый роутер с файловым хранилищем или базой данных.
// Автоматически настраивает обработчик сигналов для корректного завершения работы.
// Принимает cfg - конфигурацию сервера.
// Возвращает настроенный роутер Gin.
func NewRouterWithFS(cfg *config.ServerConfig) *gin.Engine {
	router, cleanup := NewRouterWithStorage(cfg)
	if cleanup != nil {
		go func() {
			signChan := make(chan os.Signal, 1)
			signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			<-signChan
			cleanup()
		}()
	}
	return router
}

// NewRouterWithCleanup создает новый роутер и возвращает функцию очистки для graceful shutdown.
// Эта функция используется для более точного контроля над процессом завершения работы.
func NewRouterWithCleanup(cfg *config.ServerConfig) (*gin.Engine, func()) {
	return NewRouterWithStorage(cfg)
}

// NewRouterWithStorage создает новый роутер и инициализирует хранилище на основе конфигурации.
// Поддерживает как базу данных (PostgreSQL), так и файловое хранилище.
// Возвращает настроенный роутер и функцию очистки ресурсов.
func NewRouterWithStorage(cfg *config.ServerConfig) (*gin.Engine, func()) {
	var storage repository.Storage
	var cleanup func()
	var router *gin.Engine

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
			dbRepository := repository.NewDBRepository(db)
			dbStorage, err := repository.NewDBStorage(dbRepository)
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
					if err := dbRepository.Close(); err != nil {
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
	auditManager := createAuditManager(cfg)

	var privateKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		var err error
		privateKey, err = utils.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			log.Printf("Failed to load private key from %s: %v", cfg.CryptoKey, err)
		} else {
			log.Printf("Loaded private key for decryption from %s", cfg.CryptoKey)
		}
	}

	if cfg.Key != "" {
		router = CreateRouter(storage, cfg.Key, auditManager, privateKey, cfg.TrustedSubnet)

	} else {
		router = CreateRouter(storage, "", auditManager, privateKey, cfg.TrustedSubnet)
	}
	return router, cleanup
}

func createAuditManager(cfg *config.ServerConfig) *audit.AuditManager {
	manager := audit.NewAuditManager()

	if cfg.AuditFile != "" {
		fileAuditor, err := audit.NewFileAuditor(cfg.AuditFile)
		if err != nil {
			log.Printf("Failed to create file auditor: %v", err)
		} else {
			manager.AddObserver(fileAuditor)
			log.Printf("Audit logging to file: %s", cfg.AuditFile)
		}
	}

	if cfg.AuditURL != "" {
		httpAuditor, err := audit.NewHTTPAuditor(cfg.AuditURL)
		if err != nil {
			log.Printf("Failed to create HTTP auditor: %v", err)
		} else {
			manager.AddObserver(httpAuditor)
			log.Printf("Audit logging to URL: %s", cfg.AuditURL)
		}
	}

	if !manager.IsEnabled() {
		log.Printf("Audit is disabled (no audit-file or audit-url configured)")
	}

	return manager
}

// CreateRouter создает и настраивает роутер Gin с обработчиками метрик и middleware.
// Настраивает маршруты для обновления и получения метрик, поддерживает HMAC аутентификацию,
// асимметричное шифрование и сжатие данных.
// Параметры:
//   - storage: хранилище метрик
//   - key: ключ для HMAC аутентификации (если пустой, HMAC не используется)
//   - auditManager: менеджер аудита
//   - privateKey: приватный RSA ключ для расшифровки (если nil, расшифровка не используется)
//   - trustedSubnet: CIDR нотация доверенной подсети (если пустая, ограничения не применяются)
//
// Возвращает настроенный роутер Gin.
func CreateRouter(storage repository.Storage, key string, auditManager *audit.AuditManager, privateKey *rsa.PrivateKey, trustedSubnet string) *gin.Engine {
	var metricsHandler *handler.MetricsHandler
	if auditManager != nil && auditManager.IsEnabled() {
		metricsHandler = handler.NewMetricHandlerWithAudit(storage, auditManager)
	} else {
		metricsHandler = handler.NewMetricHandler(storage)
	}

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
	router.Use(middleware.TrustedSubnetMiddleware(trustedSubnet))
	if key != "" {
		router.Use(middleware.HMACMiddleware(key))
	}
	if privateKey != nil {
		router.Use(middleware.DecryptionMiddleware(privateKey))
	}
	router.Use(middleware.CompressionMiddleware())
	router.Use(middleware.DecompressionMiddleware())

	router.POST("/update/", metricsHandler.UpdateMetricJSON)
	router.POST("/updates/", metricsHandler.UpdateMetricsBatch)
	router.POST("/value/", metricsHandler.GetMetricJSON)
	router.POST("/update/:type/:name/:value", metricsHandler.UpdateMetric)
	router.GET("/value/:type/:name", metricsHandler.GetMetric)
	router.GET("/", handler.DefaultHandle)
	router.GET("/ping", metricsHandler.DBHandler)
	router.POST("/update/:type/:name/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})
	pprofGroup := router.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}