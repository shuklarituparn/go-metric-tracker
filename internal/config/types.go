package config

import (
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type AgentConfig struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	RetryConfig    retry.RetryConfig
	Key            string `env:"KEY"`
	Ratelimit      int    `env:"RATE_LIMIT"`
}

type ServerConfig struct {
	Endpoint              string `env:"ADDRESS"`
	StoreInterval         string `env:"STORE_INTERVAL"`
	FileStoragePath       string `env:"FILE_STORAGE_PATH"`
	Restore               bool   `env:"RESTORE"`
	StoreIntervalDuration time.Duration
	DBConfig              *db.Config
	Key                   string `env:"KEY"`
	DatabaseDSN           string `env:"DATABASE_DSN"`
	RetryConfig           retry.RetryConfig
	AuditFile             string `env:"AUDIT_FILE"`
	AuditURL              string `env:"AUDIT_URL"`
}
