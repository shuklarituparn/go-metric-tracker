package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/buildinfo"
	"github.com/shuklarituparn/go-metric-tracker/internal/config"
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

	r, cleanup := router.NewRouterWithCleanup(cfg)

	srv := &http.Server{
		Addr:    cfg.Endpoint,
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on %s", cfg.Endpoint)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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