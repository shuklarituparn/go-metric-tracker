package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r := router.NewRouterWithFS(cfg)
	go func() {
		if err := r.Run(cfg.Endpoint); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down server...")
}
