package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/agent"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func main() {
	cfg := config.Load()
	storage := repository.NewMemStorage()

	metricCollector := agent.NewMetricCollector(cfg.PollInterval, storage)
	metricCollector.Start()
	time.Sleep(2 * time.Second)

	metricSender := agent.NewSender(cfg.Endpoint, cfg.ReportInterval, storage)
	metricSender.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down agent...")

	metricSender.SendMetrics()
}
