package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Supabase SupabaseConfig
}

// SupabaseConfig contains Supabase-related configuration
type SupabaseConfig struct {
	URL string
	Key string
}

// Load reads configuration from environment variables
// It automatically loads .env file if present
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Supabase: SupabaseConfig{
			URL: os.Getenv("SUPABASE_URL"),
			Key: os.Getenv("SUPABASE_KEY"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
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
