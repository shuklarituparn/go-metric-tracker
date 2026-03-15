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
	grpcclient "github.com/shuklarituparn/go-metric-tracker/internal/grpc"
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

	var serverSender agent.ServerSender
	var grpcSender *grpcclient.GRPCSender

	if cfg.GRPCEndpoint != "" {
		log.Printf("Using gRPC transport to %s", cfg.GRPCEndpoint)
		grpcSender, err = grpcclient.NewGRPCSender(cfg.GRPCEndpoint, cfg.ReportInterval, storage)
		if err != nil {
			log.Fatalf("Failed to create gRPC sender: %v", err)
		}
		grpcSender.Start()
	} else {
		fullendpoint := fmt.Sprintf("http://%s", cfg.Endpoint)
		log.Printf("Using HTTP transport to %s", fullendpoint)
		if cfg.CryptoKey != "" {
			serverSender = agent.NewSenderWithCrypto(fullendpoint, cfg.ReportInterval, storage, cfg.Key, cfg.Ratelimit, cfg.CryptoKey)
		} else {
			serverSender = agent.NewSender(fullendpoint, cfg.ReportInterval, storage, cfg.Key, cfg.Ratelimit)
		}
		serverSender.Start()
	}

	<-ctx.Done()
	stop()
	log.Println("Shutting down agent...")

	metricCollector.Stop()

	storageSender.Stop()

	if err := storageSender.SendToStorage(); err != nil {
		log.Printf("Error sending final metrics to storage: %v", err)
	}

	if grpcSender != nil {
		grpcSender.Stop()
		log.Println("Sending final metrics to server via gRPC...")
		if err := grpcSender.SendMetricsBatch(); err != nil {
			log.Printf("Error sending final metrics via gRPC: %v", err)
		}
		if err := grpcSender.Close(); err != nil {
			log.Printf("Error closing gRPC connection: %v", err)
		}
	} else if serverSender != nil {
		serverSender.Stop()
		log.Println("Sending final metrics to server via HTTP...")
		serverSender.SendMetrics()
	}

	log.Println("Agent shutdown complete")
}