package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func main() {

	storage := repository.NewMemStorage()
	metricsHandler := handler.NewMetricHandler(storage)
	debugHandler := handler.NewDebugHandler(storage)

	router:= gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST("/update/:type/:name/:value/", metricsHandler.UpdateMetric)
	router.GET("/debug",debugHandler.DebugHandler )


	addr := "localhost:8080"
	log.Printf("Starting metrics server on %s", addr)
	log.Printf("Update metrics: POST http://%s/update/<type>/<name>/<value>", addr)
	log.Printf("View metrics: GET http://%s/debug", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
