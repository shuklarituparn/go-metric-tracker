package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/agent"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func main() {
	storage := repository.NewMemStorage()
	metricCollector := agent.NewMetricCollector(2*time.Second, storage)
	metricCollector.Start()
	time.Sleep(2 * time.Second)
	metricSender := agent.NewSender("http://localhost:8080", 10*time.Second, metricCollector)
	metricSender.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down agent...")

	metricSender.SendMetrics()
}
