package models

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL     string
	OpenAIAPIKey    string
	AnthropicAPIKey string
}

// ConfigOption is a functional option for Config
type ConfigOption func(*Config) error

// NewConfig creates a new Config with the given options
func NewConfig(opts ...ConfigOption) (*Config, error) {
	cfg := &Config{}
	
	// Apply default options (load from environment)
	if err := WithEnvDefaults()(cfg); err != nil {
		return nil, err
	}
	
	// Apply user-provided options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	
	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("NEON_DATABASE_URL is required")
	}
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}
	if cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	
	return cfg, nil
}

// WithEnvDefaults loads configuration from environment variables
func WithEnvDefaults() ConfigOption {
	return func(cfg *Config) error {
		cfg.DatabaseURL = os.Getenv("NEON_DATABASE_URL")
		cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
		cfg.AnthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")
		return nil
	}
}

// WithDatabaseURL sets the database URL
func WithDatabaseURL(url string) ConfigOption {
	return func(cfg *Config) error {
		cfg.DatabaseURL = url
		return nil
	}
}

// WithOpenAIAPIKey sets the OpenAI API key
func WithOpenAIAPIKey(key string) ConfigOption {
	return func(cfg *Config) error {
		cfg.OpenAIAPIKey = key
		return nil
	}
}

// WithAnthropicAPIKey sets the Anthropic API key
func WithAnthropicAPIKey(key string) ConfigOption {
	return func(cfg *Config) error {
		cfg.AnthropicAPIKey = key
		return nil
	}
}
