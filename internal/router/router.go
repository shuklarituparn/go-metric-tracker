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

	router.Use(func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() == http.StatusNotFound {
			if c.Request.Method != "POST" &&
				(c.Request.URL.Path == "/update/" ||
					len(c.Request.URL.Path) > 8 && c.Request.URL.Path[:8] == "/update/") {
				c.AbortWithStatus(http.StatusMethodNotAllowed)
				return
			}
		}
	})

	router.POST("/update", metricsHandler.UpdateMetricJSON)

	router.POST("/update/:type/:name/:value", metricsHandler.UpdateMetric)

	router.GET("/value/:type/:name", metricsHandler.GetMetric)

	router.POST("/update/:type/:name/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})

	router.GET("/update/*path", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	})
	router.PUT("/update/*path", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	})
	router.DELETE("/update/*path", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	})
	router.PATCH("/update/*path", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	})

	router.GET("/debug", debugHandler.DebugHandler)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router

}
