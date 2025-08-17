package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func NewRouter() *gin.Engine {
	storage := repository.NewMemStorage()

	metricsHandler := handler.NewMetricHandler(storage)
	debugHandler := handler.NewDebugHandler(storage)

	router := gin.Default()

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

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
