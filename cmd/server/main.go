package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/router"
)
var buildVersion string
var buildDate string
var buildCommit string

func na(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", na(buildVersion))
	fmt.Printf("Build date: %s\n", na(buildDate))
	fmt.Printf("Build commit: %s\n", na(buildCommit))
}


func main() {
	printBuildInfo()
	cfg, err := config.NewConfig().LoadServerConfig()
	if err != nil {
		log.Fatalf("error loading server config :%v", err)
	}
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
