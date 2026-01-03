//Package agent creates a new agent

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/agent"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
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
	cfg, err := config.NewConfig().LoadAgentConfig()
	if err != nil {
		log.Fatalf("can't load agent config :%v", err)
	}
	storage := repository.NewMemStorage()

	metricCollector := agent.NewMetricCollector(cfg.PollInterval)
	metricCollector.Start()
	storageSender := agent.NewStorageSender(metricCollector, storage, cfg.PollInterval)
	storageSender.Start()

	time.Sleep(2 * time.Second)
	fullendpoint := fmt.Sprintf("http://%s", cfg.Endpoint)
	serverSender := agent.NewSender(fullendpoint, cfg.ReportInterval, storage, cfg.Key, cfg.Ratelimit)
	serverSender.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down agent...")

	serverSender.SendMetrics()
}
