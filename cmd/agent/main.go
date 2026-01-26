// Package agent creates a new agent

package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/agent"
	"github.com/shuklarituparn/go-metric-tracker/internal/buildinfo"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	buildinfo.Info{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	}.Print()

	cfg, err := config.NewConfig().LoadAgentConfig()
	if err != nil {
		log.Fatalf("can't load agent config :%v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	storage := repository.NewMemStorage()

	metricCollector := agent.NewMetricCollector(cfg.PollInterval)
	metricCollector.Start()
	storageSender := agent.NewStorageSender(metricCollector, storage, cfg.PollInterval)
	storageSender.Start()

	time.Sleep(2 * time.Second)
	fullendpoint := fmt.Sprintf("http://%s", cfg.Endpoint)
	var serverSender agent.ServerSender
	if cfg.CryptoKey != "" {
		serverSender = agent.NewSenderWithCrypto(fullendpoint, cfg.ReportInterval, storage, cfg.Key, cfg.Ratelimit, cfg.CryptoKey)
	} else {
		serverSender = agent.NewSender(fullendpoint, cfg.ReportInterval, storage, cfg.Key, cfg.Ratelimit)
	}
	serverSender.Start()

	<-ctx.Done()
	stop()
	log.Println("Shutting down agent...")

	metricCollector.Stop()

	storageSender.Stop()

	serverSender.Stop()

	storageSender.SendToStorage()
	log.Println("Sending final metrics to server...")
	serverSender.SendMetrics()
	log.Println("Agent shutdown complete")
}