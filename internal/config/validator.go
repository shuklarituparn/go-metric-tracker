package config

import (
	"errors"
	"strings"
	"time"
)

type ConfigValidator struct{}

func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

func (v *ConfigValidator) ValidateAgentConfig(config *AgentConfig) error {
	var errs []string

	if config.Endpoint == "" {
		errs = append(errs, "endpoint cannot be empty")
	}

	if config.PollInterval <= 0 {
		errs = append(errs, "poll interval must be positive")
	}

	if config.ReportInterval <= 0 {
		errs = append(errs, "report interval must be positive")
	}

	if config.PollInterval > config.ReportInterval {
		errs = append(errs, "poll interval should not exceed report interval")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func (v *ConfigValidator) ValidateServerConfig(config *ServerConfig) error {
	var errs []string

	if config.Endpoint == "" {
		errs = append(errs, "endpoint cannot be empty")
	}

	if config.StoreIntervalDuration < 0 {
		errs = append(errs, "store interval must be non-negative")
	}

	if config.FileStoragePath == "" && config.DatabaseDSN == "" {
		errs = append(errs, "either file storage path or database DSN must be provided")
	}

	if config.StoreIntervalDuration > 24*time.Hour {
		errs = append(errs, "store interval exceeds 24 hours, which may be too long")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
