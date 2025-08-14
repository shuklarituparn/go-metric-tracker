package main

import (
	"log"
	"net/http"

	"github.com/shuklarituparn/go-metric-tracker/internal/handler"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func main() {

	storage := repository.NewMemStorage()
	metricsHandler := handler.NewMetricHandler(storage)
	healthHandler := handler.NewHealthHandler(storage)

	mux := http.NewServeMux()
	mux.Handle("/update/", metricsHandler)

	debugHandler := handler.NewDebugHandler(storage)
	mux.Handle("/debug", debugHandler)
	mux.Handle("/health", healthHandler)

	addr := "localhost:8080"
	log.Printf("Starting metrics server on %s", addr)
	log.Printf("Update metrics: POST http://%s/update/<type>/<name>/<value>", addr)
	log.Printf("View metrics: GET http://%s/debug", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
