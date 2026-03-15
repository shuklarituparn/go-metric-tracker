package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/shuklarituparn/go-metric-tracker/internal/buildinfo"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
	grpcserver "github.com/shuklarituparn/go-metric-tracker/internal/grpc"
	"github.com/shuklarituparn/go-metric-tracker/internal/router"
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

	cfg, err := config.NewConfig().LoadServerConfig()
	if err != nil {
		log.Fatalf("error loading server config :%v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	r, cleanup, storage := router.NewRouterWithStorageAndCleanup(cfg)

	srv := &http.Server{
		Addr:    cfg.Endpoint,
		Handler: r,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Endpoint)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	var grpcSrv *grpcserver.Server
	if cfg.GRPCEndpoint != "" {
		logger, err := zap.NewProduction()
		if err != nil {
			log.Printf("Failed to create zap logger for gRPC: %v", err)
			logger = zap.NewNop()
		}

		grpcSrv, err = grpcserver.NewServer(storage, cfg.TrustedSubnet, logger)
		if err != nil {
			log.Fatalf("Failed to create gRPC server: %v", err)
		}

		go func() {
			log.Printf("Starting gRPC server on %s", cfg.GRPCEndpoint)
			if err := grpcSrv.Start(cfg.GRPCEndpoint); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	}

	<-ctx.Done()
	stop()
	log.Println("Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if grpcSrv != nil {
		grpcSrv.Stop()
		log.Println("gRPC server stopped gracefully")
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	if cleanup != nil {
		cleanup()
	}

	log.Println("Server shutdown complete")
}