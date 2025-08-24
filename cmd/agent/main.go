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

func main() {
	cfg := config.Load()
	storage := repository.NewMemStorage()

	metricCollector := agent.NewMetricCollector(cfg.PollInterval, storage)
	metricCollector.Start()
	time.Sleep(2 * time.Second)
	fullendpoint :=fmt.Sprintf("http://%s", cfg.Endpoint)

	metricSender := agent.NewSender(fullendpoint, cfg.ReportInterval, storage)
	metricSender.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down agent...")

	metricSender.SendMetrics()
}
