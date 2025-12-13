package models

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL       string
	OpenAIAPIKey      string
	AnthropicAPIKey   string
	EmbeddingProvider string // "openai" or "ollama"
	OllamaURL         string // Ollama API URL (default: http://localhost:11434)
	OllamaModel       string // Ollama embedding model (default: nomic-embed-text)
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

	// Set defaults
	if cfg.EmbeddingProvider == "" {
		cfg.EmbeddingProvider = "openai" // Default to OpenAI
	}
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = "http://localhost:11434"
	}
	if cfg.OllamaModel == "" {
		cfg.OllamaModel = "nomic-embed-text"
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("NEON_DATABASE_URL is required")
	}
	if cfg.EmbeddingProvider == "openai" && cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required when using OpenAI embeddings")
	}
	if cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}

	return cfg, nil
}

// WithEnvDefaults loads configuration from environment variables
// It automatically loads .env file if it exists (errors are ignored if file doesn't exist)
func WithEnvDefaults() ConfigOption {
	return func(cfg *Config) error {
		// Try to load .env file (ignore error if file doesn't exist)
		_ = godotenv.Load()

		cfg.DatabaseURL = os.Getenv("NEON_DATABASE_URL")
		cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
		cfg.AnthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")
		cfg.EmbeddingProvider = os.Getenv("EMBEDDING_PROVIDER")
		cfg.OllamaURL = os.Getenv("OLLAMA_URL")
		cfg.OllamaModel = os.Getenv("OLLAMA_MODEL")
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
