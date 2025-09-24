package config

import "fmt"

type Config struct {
	parser    *ConfigParser
	validator *ConfigValidator
}

func NewConfig() *Config {
	return &Config{
		parser:    NewConfigParser(),
		validator: NewConfigValidator(),
	}
}

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
