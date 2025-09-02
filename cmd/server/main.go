package main

import (
	"log"

	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/router"
)

func main() {

	cfg := config.Load()
	router := router.NewRouter()
	if err := router.Run(cfg.Endpoint); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
