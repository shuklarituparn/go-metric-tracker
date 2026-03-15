package config

import (
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type AgentConfig struct {
	Endpoint       string `env:"ADDRESS"`
	GRPCEndpoint   string `env:"GRPC_ADDRESS"`
	Key            string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	RetryConfig    retry.RetryConfig
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Ratelimit      int           `env:"RATE_LIMIT"`
}

type ServerConfig struct {
	Endpoint              string `env:"ADDRESS"`
	GRPCEndpoint          string `env:"GRPC_ADDRESS"`
	StoreInterval         string `env:"STORE_INTERVAL"`
	FileStoragePath       string `env:"FILE_STORAGE_PATH"`
	Key                   string `env:"KEY"`
	CryptoKey             string `env:"CRYPTO_KEY"`
	DatabaseDSN           string `env:"DATABASE_DSN"`
	AuditFile             string `env:"AUDIT_FILE"`
	AuditURL              string `env:"AUDIT_URL"`
	TrustedSubnet         string `env:"TRUSTED_SUBNET"`
	DBConfig              *db.Config
	RetryConfig           retry.RetryConfig
	StoreIntervalDuration time.Duration
	Restore               bool `env:"RESTORE"`
}