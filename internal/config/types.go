package config

import (
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type AgentConfig struct {
    Endpoint       string        `env:"ADDRESS"`
    PollInterval   time.Duration `env:"POLL_INTERVAL"`
    ReportInterval time.Duration `env:"REPORT_INTERVAL"`
    RetryConfig    retry.RetryConfig
}

type ServerConfig struct {
	Endpoint              string `env:"ADDRESS"`
	StoreInterval         string `env:"STORE_INTERVAL"`
	FileStoragePath       string `env:"FILE_STORAGE_PATH"`
	Restore               bool   `env:"RESTORE"`
	StoreIntervalDuration time.Duration
	DBConfig              *db.Config
	DatabaseDSN           string `env:"DATABASE_DSN"`
	RetryConfig           retry.RetryConfig
}
