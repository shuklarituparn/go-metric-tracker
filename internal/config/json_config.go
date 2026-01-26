package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
)

type ServerJSONConfig struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

type AgentJSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

type JSONConfigLoader struct{}

func NewJSONConfigLoader() *JSONConfigLoader {
	return &JSONConfigLoader{}
}

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
	if jsonCfg.TrustedSubnet != "" {
		cfg.TrustedSubnet = jsonCfg.TrustedSubnet
	}
}


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