// Package config предоставляет функционал для загрузки и валидации конфигурации приложения.
package config

import "fmt"

// Config объединяет парсер и валидатор конфигурации.
type Config struct {
	parser    *ConfigParser
	validator *ConfigValidator
}

// NewConfig создает новый экземпляр Config с парсером и валидатором.
func NewConfig() *Config {
	return &Config{
		parser:    NewConfigParser(),
		validator: NewConfigValidator(),
	}
}

// LoadAgentConfig загружает и валидирует конфигурацию агента.
// Возвращает ошибку, если конфигурация невалидна или не может быть загружена.
func (c *Config) LoadAgentConfig() (*AgentConfig, error) {
	config, err := c.parser.ParseAgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}
	if err := c.validator.ValidateAgentConfig(config); err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}
	return config, nil
}

// LoadServerConfig загружает и валидирует конфигурацию сервера.
// Возвращает ошибку, если конфигурация невалидна или не может быть загружена.
func (c *Config) LoadServerConfig() (*ServerConfig, error) {
	config, err := c.parser.ParseServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}
	if err := c.validator.ValidateServerConfig(config); err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}
	return config, nil
}
