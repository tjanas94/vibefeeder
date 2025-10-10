package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	Supabase   SupabaseConfig
	Log        LogConfig
	OpenRouter OpenRouterConfig
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Address string
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// SupabaseConfig contains Supabase-related configuration
type SupabaseConfig struct {
	URL string
	Key string
}

// OpenRouterConfig contains OpenRouter AI service configuration
type OpenRouterConfig struct {
	APIKey string
}

// Load reads configuration from environment variables
// It automatically loads .env file if present
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Address: getEnvOrDefault("SERVER_ADDRESS", "localhost:8080"),
		},
		Supabase: SupabaseConfig{
			URL: os.Getenv("SUPABASE_URL"),
			Key: os.Getenv("SUPABASE_KEY"),
		},
		Log: LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		OpenRouter: OpenRouterConfig{
			APIKey: getEnvOrDefault("OPENROUTER_API_KEY", "mock-api-key"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validate checks if all required configuration values are set
func (c *Config) validate() error {
	if c.Supabase.URL == "" {
		return fmt.Errorf("SUPABASE_URL is required")
	}

	if c.Supabase.Key == "" {
		return fmt.Errorf("SUPABASE_KEY is required")
	}

	return nil
}
