package config

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type DefaultValues struct {
	Endpoint             string
	ReportInterval       string
	PollInterval         string
	StoreInterval        string
	FileStoragePath      string
	Restore              string
	AgentRetryIntervals  string
	ServerRetryIntervals string
}

type ConfigParser struct {
	defaults *DefaultValues
}

type agentFlags struct {
	endpoint       string
	reportInterval string
	pollInterval   string
	retryIntervals string
}

type serverFlags struct {
	endpoint        string
	storeInterval   string
	fileStoragePath string
	restore         string
	databaseDSN     string
	retryIntervals  string
}

func NewConfigParser() *ConfigParser {
	return &ConfigParser{
		defaults: &DefaultValues{
			Endpoint:             "localhost:8080",
			ReportInterval:       "10",
			PollInterval:         "2",
			StoreInterval:        "300",
			FileStoragePath:      "/tmp/metrics-store.json",
			Restore:              "true",
			AgentRetryIntervals:  "1,3,5",
			ServerRetryIntervals: "2,4,8",
		},
	}
}

func (p *ConfigParser) parseRetryIntervals(intervals string) []time.Duration {
	if intervals == "" {
		return nil
	}

	parts := strings.Split(intervals, ",")
	durations := make([]time.Duration, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if seconds, err := strconv.Atoi(part); err == nil && seconds > 0 {
			durations = append(durations, time.Duration(seconds)*time.Second)
		}
	}

	return durations
}

func (p *ConfigParser) createRetryConfig(intervals []time.Duration) retry.RetryConfig {
	if len(intervals) == 0 {
		return retry.DefaultConfig()
	}

	return retry.RetryConfig{
		MaxAttempts: len(intervals) + 1,
		Intervals:   intervals,
	}
}

func (p *ConfigParser) ParseAgentConfig() (*AgentConfig, error) {

	pollTimeDuration := p.parseDurationWithDefault(p.defaults.PollInterval, 2)
	reportTimeDuration := p.parseDurationWithDefault(p.defaults.ReportInterval, 10)

	defaultRetryIntervals := p.parseRetryIntervals(p.defaults.AgentRetryIntervals)

	appConfig := &AgentConfig{
		Endpoint:       p.defaults.Endpoint,
		PollInterval:   pollTimeDuration,
		ReportInterval: reportTimeDuration,
		RetryConfig:    p.createRetryConfig(defaultRetryIntervals),
	}
	var flags agentFlags
	flag.StringVar(&flags.endpoint, "a", p.defaults.Endpoint, "endpoint address")
	flag.StringVar(&flags.reportInterval, "r",p.defaults.ReportInterval, "report interval (seconds)")
	flag.StringVar(&flags.pollInterval, "p", p.defaults.PollInterval, "poll interval (seconds)")
	flag.StringVar(&flags.retryIntervals, "retry-intervals", p.defaults.AgentRetryIntervals, "retry intervals in seconds, comma-separated (e.g., '1,3,5')")
	flag.Parse()

	if flags.endpoint != "" {
		appConfig.Endpoint = flags.endpoint
	}
	if flags.reportInterval != "" {
		if duration, err := strconv.Atoi(flags.reportInterval); err == nil {
			appConfig.ReportInterval = time.Duration(duration) * time.Second
		}
	}
	if flags.pollInterval != "" {
		if duration, err := strconv.Atoi(flags.pollInterval); err == nil {
			appConfig.PollInterval = time.Duration(duration) * time.Second
		}
	}

	if flags.retryIntervals != "" {
		customRetryIntervals := p.parseRetryIntervals(flags.retryIntervals)
		if len(customRetryIntervals) > 0 {
			appConfig.RetryConfig = p.createRetryConfig(customRetryIntervals)
		}
	}

	if err := env.Parse(appConfig); err != nil {
		return appConfig, err
	}

	return appConfig, nil
}

func (p *ConfigParser) ParseServerConfig() (*ServerConfig, error) {

	storeIntervalDuration := p.parseDurationWithDefault(p.defaults.StoreInterval, 300)

	restoreValue := p.parseBoolWithDefault(p.defaults.Restore, true)

	defaultRetryIntervals := p.parseRetryIntervals(p.defaults.ServerRetryIntervals)

	config := &ServerConfig{
		Endpoint:              p.defaults.Endpoint,
		StoreInterval:         p.defaults.StoreInterval,
		StoreIntervalDuration: storeIntervalDuration,
		FileStoragePath:       p.defaults.FileStoragePath,
		Restore:               restoreValue,
		DBConfig:              db.NewDefaultConfig(),
		DatabaseDSN:           "",
		RetryConfig:           p.createRetryConfig(defaultRetryIntervals),
	}

	var flags serverFlags
	flag.StringVar(&flags.endpoint, "a", p.defaults.Endpoint, "endpoint address")
	flag.StringVar(&flags.storeInterval, "i", p.defaults.StoreInterval, "store interval (seconds)")
	flag.StringVar(&flags.fileStoragePath, "f", p.defaults.FileStoragePath, "file storage path")
	flag.StringVar(&flags.restore, "r", p.defaults.Restore, "restore from file (true/false)")
	flag.StringVar(&flags.databaseDSN, "d", "", "database DSN")
	flag.StringVar(&flags.retryIntervals, "retry-intervals", "", "retry intervals in seconds, comma-separated (e.g., '2,4,8')")
	flag.Parse()

	if flags.endpoint != "" {
		config.Endpoint = flags.endpoint
	}
	if flags.storeInterval != "" {
		config.StoreInterval = flags.storeInterval
		if duration, err := strconv.Atoi(flags.storeInterval); err == nil {
			config.StoreIntervalDuration = time.Duration(duration) * time.Second
		}
	}
	if flags.restore != "" {
		if restoreVal, err := strconv.ParseBool(flags.restore); err == nil {
			config.Restore = restoreVal
		}
	}
	if flags.fileStoragePath != "" {
		config.FileStoragePath = flags.fileStoragePath
	}
	if flags.databaseDSN != "" {
		config.DatabaseDSN = flags.databaseDSN
		config.DBConfig.DSN = flags.databaseDSN
	}

	if flags.retryIntervals != "" {
		customRetryIntervals := p.parseRetryIntervals(flags.retryIntervals)
		if len(customRetryIntervals) > 0 {
			config.RetryConfig = p.createRetryConfig(customRetryIntervals)
		}
	}

	if err := env.Parse(config); err != nil {
		return config, err
	}

	if config.DatabaseDSN != "" {
		config.DBConfig.DSN = config.DatabaseDSN
	}
	if config.StoreInterval != "" {
		if duration, err := strconv.Atoi(config.StoreInterval); err == nil {
			config.StoreIntervalDuration = time.Duration(duration) * time.Second
		}
	}

	return config, nil
}

func (p *ConfigParser) parseDurationWithDefault(value string, defaultSeconds int) time.Duration {
	if duration, err := strconv.Atoi(value); err == nil {
		return time.Duration(duration) * time.Second
	}
	return time.Duration(defaultSeconds) * time.Second
}

func (p *ConfigParser) parseBoolWithDefault(value string, defaultValue bool) bool {
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}
	return defaultValue
}
