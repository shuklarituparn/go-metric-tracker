package main

import (
	"flag"
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
	endpoint:= flag.String("a","localhost:8080", "endpoint address")
	reportInterval:= flag.Duration("r",10*time.Second, "report interval" )
	pollInterval:=flag.Duration("p",2*time.Second, "poll interval" )
	flag.Parse()

	metricCollector := agent.NewMetricCollector(*pollInterval, storage)
	metricCollector.Start()
	time.Sleep(2 * time.Second)
	metricSender := agent.NewSender(*endpoint, *reportInterval, metricCollector)
	metricSender.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down agent...")

	metricSender.SendMetrics()
}
