package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/agent"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

func main() {
	storage := repository.NewMemStorage()
	endpoint := flag.String("a", "localhost:8080", "endpoint address")
	reportInterval := flag.String("r", "10", "report interval")
	pollInterval := flag.String("p", "2", "poll interval")
	flag.Parse()
	fullendpoint := "http://" + *endpoint
	correctPollDuration, _ := strconv.Atoi(*pollInterval)
	correctReportDuration, _ := strconv.Atoi(*reportInterval)
	metricCollector := agent.NewMetricCollector(time.Duration(correctPollDuration)*time.Second, storage)
	metricCollector.Start()
	time.Sleep(2 * time.Second)

	metricSender := agent.NewSender(fullendpoint, time.Duration(correctReportDuration)*time.Second, metricCollector)
	metricSender.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down agent...")

	metricSender.SendMetrics()
}
