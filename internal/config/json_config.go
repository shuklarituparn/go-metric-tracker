package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
)

// ServerJSONConfig represents server configuration from JSON file.
type ServerJSONConfig struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

// AgentJSONConfig represents agent configuration from JSON file.
type AgentJSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// JSONConfigLoader handles loading configuration from JSON files.
type JSONConfigLoader struct{}

// NewJSONConfigLoader creates a new JSONConfigLoader instance.
func NewJSONConfigLoader() *JSONConfigLoader {
	return &JSONConfigLoader{}
}

// LoadServerJSONConfig loads server configuration from a JSON file.
func (l *JSONConfigLoader) LoadServerJSONConfig(filePath string) (*ServerJSONConfig, error) {
	if filePath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg ServerJSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadAgentJSONConfig loads agent configuration from a JSON file.
func (l *JSONConfigLoader) LoadAgentJSONConfig(filePath string) (*AgentJSONConfig, error) {
	if filePath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg AgentJSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ApplyServerJSONConfig applies JSON configuration to ServerConfig.
// Only non-empty values from JSON config are applied.
func (l *JSONConfigLoader) ApplyServerJSONConfig(cfg *ServerConfig, jsonCfg *ServerJSONConfig) {
	if jsonCfg == nil {
		return
	}

	if jsonCfg.Address != "" {
		cfg.Endpoint = jsonCfg.Address
	}
	if jsonCfg.Restore != nil {
		cfg.Restore = *jsonCfg.Restore
	}
	if jsonCfg.StoreInterval != "" {
		cfg.StoreInterval = l.parseIntervalToSeconds(jsonCfg.StoreInterval)
		cfg.StoreIntervalDuration = l.parseDuration(jsonCfg.StoreInterval)
	}
	if jsonCfg.StoreFile != "" {
		cfg.FileStoragePath = jsonCfg.StoreFile
	}
	if jsonCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonCfg.DatabaseDSN
		if cfg.DBConfig == nil {
			cfg.DBConfig = db.NewDefaultConfig()
		}
		cfg.DBConfig.DSN = jsonCfg.DatabaseDSN
	}
	if jsonCfg.CryptoKey != "" {
		cfg.CryptoKey = jsonCfg.CryptoKey
	}
}

// ApplyAgentJSONConfig applies JSON configuration to AgentConfig.
// Only non-empty values from JSON config are applied.
func (l *JSONConfigLoader) ApplyAgentJSONConfig(cfg *AgentConfig, jsonCfg *AgentJSONConfig) {
	if jsonCfg == nil {
		return
	}

	if jsonCfg.Address != "" {
		cfg.Endpoint = jsonCfg.Address
	}
	if jsonCfg.ReportInterval != "" {
		cfg.ReportInterval = l.parseDuration(jsonCfg.ReportInterval)
	}
	if jsonCfg.PollInterval != "" {
		cfg.PollInterval = l.parseDuration(jsonCfg.PollInterval)
	}
	if jsonCfg.CryptoKey != "" {
		cfg.CryptoKey = jsonCfg.CryptoKey
	}
}

// parseDuration parses duration string in format "1s", "10s", "1m", etc.
// Falls back to parsing as seconds if no unit is specified.
func (l *JSONConfigLoader) parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}

	if d, err := time.ParseDuration(s); err == nil {
		return d
	}

	if seconds, err := strconv.Atoi(s); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return 0
}

// parseIntervalToSeconds converts duration string to seconds string for StoreInterval field.
func (l *JSONConfigLoader) parseIntervalToSeconds(s string) string {
	if s == "" {
		return ""
	}

	if d, err := time.ParseDuration(s); err == nil {
		return strconv.Itoa(int(d.Seconds()))
	}

	if _, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return strings.TrimSpace(s)
	}

	return s
}