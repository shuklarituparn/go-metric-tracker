package config

import (
	"flag"
	"log"
	"os"
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
	Key                  string
}

type ConfigParser struct {
	defaults *DefaultValues
}

type agentFlags struct {
	endpoint       string
	reportInterval string
	pollInterval   string
	retryIntervals string
	key            string
	rateLimit      int
}

type serverFlags struct {
	endpoint        string
	storeInterval   string
	fileStoragePath string
	restore         string
	databaseDSN     string
	retryIntervals  string
	key             string
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
			Key:                  "",
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

func (p *ConfigParser) ParseAgentFlags(args []string) (*agentFlags, error) {
	var flags agentFlags
	fs := flag.NewFlagSet("agent", flag.PanicOnError)
	fs.StringVar(&flags.endpoint, "a", p.defaults.Endpoint, "endpoint address")
	fs.StringVar(&flags.reportInterval, "r", p.defaults.ReportInterval, "report interval (seconds)")
	fs.StringVar(&flags.pollInterval, "p", p.defaults.PollInterval, "poll interval (seconds)")
	fs.StringVar(&flags.retryIntervals, "retry-intervals", p.defaults.AgentRetryIntervals, "retry intervals in seconds, comma-separated (e.g., '1,3,5')")
	fs.StringVar(&flags.key, "k", p.defaults.Key, "key for the hmac signature")
	fs.IntVar(&flags.rateLimit, "l", 3, "rate-limit value")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &flags, nil

}

func (p *ConfigParser) ParseServerFlags(args []string) (*serverFlags, error) {
	var flags serverFlags
	fs := flag.NewFlagSet("server", flag.PanicOnError)
	fs.StringVar(&flags.endpoint, "a", p.defaults.Endpoint, "endpoint address")
	fs.StringVar(&flags.storeInterval, "i", p.defaults.StoreInterval, "store interval (seconds)")
	fs.StringVar(&flags.fileStoragePath, "f", p.defaults.FileStoragePath, "file storage path")
	fs.StringVar(&flags.restore, "r", p.defaults.Restore, "restore from file (true/false)")
	fs.StringVar(&flags.databaseDSN, "d", "", "database DSN")
	fs.StringVar(&flags.retryIntervals, "retry-intervals", "", "retry intervals in seconds, comma-separated (e.g., '2,4,8')")
	fs.StringVar(&flags.key, "k", p.defaults.Key, "key for the hmac signature")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &flags, nil
}

func (p *ConfigParser) MergeServerFlags(cfg *ServerConfig, flags *serverFlags) (*ServerConfig, error) {
	var serverConfig ServerConfig

	if flags.endpoint != "" {
		serverConfig.Endpoint = flags.endpoint
	}
	if flags.storeInterval != "" {
		serverConfig.StoreInterval = flags.storeInterval
		serverConfig.StoreIntervalDuration = p.parseDurationWithDefault(flags.storeInterval, 300)
	}
	if flags.restore != "" {
		serverConfig.Restore = p.parseBoolWithDefault(flags.restore, true)

	}
	if flags.fileStoragePath != "" {
		serverConfig.FileStoragePath = flags.fileStoragePath
	}
	if flags.databaseDSN != "" {
		serverConfig.DatabaseDSN = flags.databaseDSN
		serverConfig.DBConfig.DSN = flags.databaseDSN
	}
	if flags.retryIntervals != "" {
		intervals := p.parseRetryIntervals(flags.retryIntervals)
		if len(intervals) > 0 {
			serverConfig.RetryConfig = p.createRetryConfig(intervals)
		}
	}
	if flags.key != "" {
		serverConfig.Key = flags.key
	}
	return &serverConfig, nil
}

func (p *ConfigParser) MergeAgentFlags(cfg *AgentConfig, flags *agentFlags) (*AgentConfig, error) {
	var agentConfig AgentConfig
	if flags.endpoint != "" {
		agentConfig.Endpoint = flags.endpoint
	}
	if flags.reportInterval != "" {
		agentConfig.ReportInterval = p.parseDurationWithDefault(flags.reportInterval, 10)
	}
	if flags.pollInterval != "" {
		agentConfig.PollInterval = p.parseDurationWithDefault(flags.pollInterval, 2)
	}
	if flags.retryIntervals != "" {
		intervals := p.parseRetryIntervals(flags.retryIntervals)
		if len(intervals) > 0 {
			agentConfig.RetryConfig = p.createRetryConfig(intervals)
		}
	}
	if flags.key != "" {
		agentConfig.Key = flags.key
	}
	if flags.rateLimit > 0 {
		agentConfig.Ratelimit = flags.rateLimit
	}
	return &agentConfig, nil
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
		Key:            "",
		Ratelimit:      3,
	}

	flags, err := p.ParseAgentFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("err: mistake while parsing flag:%v", err)
	}
	agentConfig, err := p.MergeAgentFlags(appConfig, flags)
	if err != nil {
		log.Fatalf("err: problem merging flag and config: %v", err)
	}

	if err := env.Parse(agentConfig); err != nil {
		log.Printf("err: parsing env variables: %v", err)
	}

	return agentConfig, nil
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
		Key:                   "",
	}

	flags, err := p.ParseServerFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("err: mistake while parsing flag:%v", err)
	}

	serverConf, err := p.MergeServerFlags(config, flags)
	if err != nil {
		log.Fatalf("err: problem merging server flags:%v", err)
	}

	if err := env.Parse(serverConf); err != nil {
		return serverConf, err
	}

	if serverConf.DatabaseDSN != "" {
		serverConf.DBConfig.DSN = serverConf.DatabaseDSN
	}
	if serverConf.StoreInterval != "" {
		serverConf.StoreIntervalDuration = p.parseDurationWithDefault(serverConf.StoreInterval, 300)
	}

	return serverConf, nil
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
