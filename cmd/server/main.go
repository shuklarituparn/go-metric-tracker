package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/router"
)

func main() {
	cfg := config.LoadServerConfig()
	router := router.NewRouterWithFS(cfg)
	go func() {
		if err := router.Run(cfg.Endpoint); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shut down server ")
}
