package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shuklarituparn/go-metric-tracker/internal/config/db"
)

func TestLoadServerJSONConfig(t *testing.T) {
	content := `{
		"address": "localhost:9090",
		"restore": true,
		"store_interval": "10s",
		"store_file": "/tmp/test.db",
		"database_dsn": "postgres://user:pass@localhost/db",
		"crypto_key": "/path/to/key.pem"
	}`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "server_config.json")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	loader := NewJSONConfigLoader()
	cfg, err := loader.LoadServerJSONConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load server JSON config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Errorf("Expected address 'localhost:9090', got '%s'", cfg.Address)
	}
	if cfg.Restore == nil || *cfg.Restore != true {
		t.Errorf("Expected restore true, got %v", cfg.Restore)
	}
	if cfg.StoreInterval != "10s" {
		t.Errorf("Expected store_interval '10s', got '%s'", cfg.StoreInterval)
	}
	if cfg.StoreFile != "/tmp/test.db" {
		t.Errorf("Expected store_file '/tmp/test.db', got '%s'", cfg.StoreFile)
	}
	if cfg.DatabaseDSN != "postgres://user:pass@localhost/db" {
		t.Errorf("Expected database_dsn 'postgres://user:pass@localhost/db', got '%s'", cfg.DatabaseDSN)
	}
	if cfg.CryptoKey != "/path/to/key.pem" {
		t.Errorf("Expected crypto_key '/path/to/key.pem', got '%s'", cfg.CryptoKey)
	}
}

func TestLoadAgentJSONConfig(t *testing.T) {
	content := `{
		"address": "localhost:9090",
		"report_interval": "5s",
		"poll_interval": "2s",
		"crypto_key": "/path/to/public.pem"
	}`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agent_config.json")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	loader := NewJSONConfigLoader()
	cfg, err := loader.LoadAgentJSONConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load agent JSON config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Errorf("Expected address 'localhost:9090', got '%s'", cfg.Address)
	}
	if cfg.ReportInterval != "5s" {
		t.Errorf("Expected report_interval '5s', got '%s'", cfg.ReportInterval)
	}
	if cfg.PollInterval != "2s" {
		t.Errorf("Expected poll_interval '2s', got '%s'", cfg.PollInterval)
	}
	if cfg.CryptoKey != "/path/to/public.pem" {
		t.Errorf("Expected crypto_key '/path/to/public.pem', got '%s'", cfg.CryptoKey)
	}
}

func TestApplyServerJSONConfig(t *testing.T) {
	loader := NewJSONConfigLoader()

	cfg := &ServerConfig{
		Endpoint:              "localhost:8080",
		Restore:               false,
		StoreInterval:         "300",
		StoreIntervalDuration: 300 * time.Second,
		FileStoragePath:       "/tmp/default.json",
		DatabaseDSN:           "",
		CryptoKey:             "",
		DBConfig:              db.NewDefaultConfig(),
	}

	restore := true
	jsonCfg := &ServerJSONConfig{
		Address:       "localhost:9090",
		Restore:       &restore,
		StoreInterval: "10s",
		StoreFile:     "/tmp/test.db",
		DatabaseDSN:   "postgres://localhost/db",
		CryptoKey:     "/path/to/key.pem",
	}

	loader.ApplyServerJSONConfig(cfg, jsonCfg)

	if cfg.Endpoint != "localhost:9090" {
		t.Errorf("Expected endpoint 'localhost:9090', got '%s'", cfg.Endpoint)
	}
	if cfg.Restore != true {
		t.Errorf("Expected restore true, got %v", cfg.Restore)
	}
	if cfg.StoreIntervalDuration != 10*time.Second {
		t.Errorf("Expected store_interval 10s, got %v", cfg.StoreIntervalDuration)
	}
	if cfg.FileStoragePath != "/tmp/test.db" {
		t.Errorf("Expected store_file '/tmp/test.db', got '%s'", cfg.FileStoragePath)
	}
	if cfg.DatabaseDSN != "postgres://localhost/db" {
		t.Errorf("Expected database_dsn 'postgres://localhost/db', got '%s'", cfg.DatabaseDSN)
	}
	if cfg.CryptoKey != "/path/to/key.pem" {
		t.Errorf("Expected crypto_key '/path/to/key.pem', got '%s'", cfg.CryptoKey)
	}
}

func TestApplyAgentJSONConfig(t *testing.T) {
	loader := NewJSONConfigLoader()

	cfg := &AgentConfig{
		Endpoint:       "localhost:8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
		CryptoKey:      "",
	}

	jsonCfg := &AgentJSONConfig{
		Address:        "localhost:9090",
		ReportInterval: "5s",
		PollInterval:   "1s",
		CryptoKey:      "/path/to/public.pem",
	}

	loader.ApplyAgentJSONConfig(cfg, jsonCfg)

	if cfg.Endpoint != "localhost:9090" {
		t.Errorf("Expected endpoint 'localhost:9090', got '%s'", cfg.Endpoint)
	}
	if cfg.ReportInterval != 5*time.Second {
		t.Errorf("Expected report_interval 5s, got %v", cfg.ReportInterval)
	}
	if cfg.PollInterval != 1*time.Second {
		t.Errorf("Expected poll_interval 1s, got %v", cfg.PollInterval)
	}
	if cfg.CryptoKey != "/path/to/public.pem" {
		t.Errorf("Expected crypto_key '/path/to/public.pem', got '%s'", cfg.CryptoKey)
	}
}

func TestParseDuration(t *testing.T) {
	loader := NewJSONConfigLoader()

	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1s", 1 * time.Second},
		{"10s", 10 * time.Second},
		{"1m", 1 * time.Minute},
		{"5m", 5 * time.Minute},
		{"1h", 1 * time.Hour},
		{"10", 10 * time.Second}, 
		{"", 0},
	}

	for _, tt := range tests {
		result := loader.parseDuration(tt.input)
		if result != tt.expected {
			t.Errorf("parseDuration(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestLoadEmptyConfigPath(t *testing.T) {
	loader := NewJSONConfigLoader()

	serverCfg, err := loader.LoadServerJSONConfig("")
	if err != nil {
		t.Errorf("Expected no error for empty path, got %v", err)
	}
	if serverCfg != nil {
		t.Errorf("Expected nil config for empty path")
	}

	agentCfg, err := loader.LoadAgentJSONConfig("")
	if err != nil {
		t.Errorf("Expected no error for empty path, got %v", err)
	}
	if agentCfg != nil {
		t.Errorf("Expected nil config for empty path")
	}
}

func TestApplyNilJSONConfig(t *testing.T) {
	loader := NewJSONConfigLoader()

	serverCfg := &ServerConfig{Endpoint: "original"}
	loader.ApplyServerJSONConfig(serverCfg, nil)
	if serverCfg.Endpoint != "original" {
		t.Errorf("Expected endpoint to remain 'original', got '%s'", serverCfg.Endpoint)
	}

	agentCfg := &AgentConfig{Endpoint: "original"}
	loader.ApplyAgentJSONConfig(agentCfg, nil)
	if agentCfg.Endpoint != "original" {
		t.Errorf("Expected endpoint to remain 'original', got '%s'", agentCfg.Endpoint)
	}
}

func TestPartialJSONConfig(t *testing.T) {
	content := `{
		"address": "localhost:9090"
	}`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial_config.json")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	loader := NewJSONConfigLoader()

	serverCfg := &ServerConfig{
		Endpoint:        "localhost:8080",
		FileStoragePath: "/tmp/original.json",
		Restore:         true,
		DBConfig:        db.NewDefaultConfig(),
	}

	jsonCfg, err := loader.LoadServerJSONConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load JSON config: %v", err)
	}
	loader.ApplyServerJSONConfig(serverCfg, jsonCfg)

	if serverCfg.Endpoint != "localhost:9090" {
		t.Errorf("Expected endpoint 'localhost:9090', got '%s'", serverCfg.Endpoint)
	}
	if serverCfg.FileStoragePath != "/tmp/original.json" {
		t.Errorf("Expected file_storage_path to remain '/tmp/original.json', got '%s'", serverCfg.FileStoragePath)
	}
	if serverCfg.Restore != true {
		t.Errorf("Expected restore to remain true, got %v", serverCfg.Restore)
	}
}

func TestRestoreFalseInJSON(t *testing.T) {
	content := `{
		"restore": false
	}`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "restore_false.json")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	loader := NewJSONConfigLoader()
	jsonCfg, err := loader.LoadServerJSONConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load JSON config: %v", err)
	}

	if jsonCfg.Restore == nil {
		t.Fatal("Expected restore to be non-nil")
	}
	if *jsonCfg.Restore != false {
		t.Errorf("Expected restore to be false, got %v", *jsonCfg.Restore)
	}

	serverCfg := &ServerConfig{
		Restore:  true,
		DBConfig: db.NewDefaultConfig(),
	}
	loader.ApplyServerJSONConfig(serverCfg, jsonCfg)

	if serverCfg.Restore != false {
		t.Errorf("Expected restore to be changed to false, got %v", serverCfg.Restore)
	}
}