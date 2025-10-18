package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	Supabase   SupabaseConfig
	Auth       AuthConfig
	Log        LogConfig
	OpenRouter OpenRouterConfig
	Fetcher    FetcherConfig
	RateLimit  RateLimitConfig
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

// AuthConfig contains authentication configuration
type AuthConfig struct {
	RedirectURL        string        // Base URL for auth redirects (e.g., http://localhost:8080 or https://yourdomain.com)
	CookieSecure       bool          // Whether to set Secure flag on cookies (true for HTTPS)
	AccessTokenMaxAge  time.Duration // Max age for access token cookie
	RefreshTokenMaxAge time.Duration // Max age for refresh token cookie
	SessionCookieName  string        // Name for session cookie
	RefreshCookieName  string        // Name for refresh token cookie
	RegistrationCode   string        // Optional code required for new user registration (if set, registration requires this code)
}

// OpenRouterConfig contains OpenRouter AI service configuration
type OpenRouterConfig struct {
	APIKey string
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	SummaryGenerationInterval time.Duration // Minimum interval between summary generation requests per user
}

// FetcherConfig holds configuration for the feed fetcher service
type FetcherConfig struct {
	FetchInterval       time.Duration // How often to check for feeds to fetch (in seconds)
	SuccessInterval     time.Duration // Minimum interval after successful fetch (in seconds)
	WorkerCount         int           // Number of concurrent workers
	BatchSize           int           // Maximum number of feeds to process per batch
	DomainDelay         time.Duration // Delay between requests to same domain (in seconds)
	JobTimeout          time.Duration // Timeout for entire job (in seconds)
	RequestTimeout      time.Duration // Timeout for HTTP request (in seconds)
	MaxArticlesPerFeed  int           // Maximum number of articles to save per feed
	MaxResponseBodySize int64         // Maximum response body size in bytes
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
		Auth: AuthConfig{
			RedirectURL:        getEnvOrDefault("AUTH_REDIRECT_URL", "http://localhost:8080"),
			CookieSecure:       getEnvOrDefault("AUTH_COOKIE_SECURE", "false") == "true",
			AccessTokenMaxAge:  getDurationSeconds("AUTH_ACCESS_TOKEN_MAX_AGE", 3600),    // 1 hour
			RefreshTokenMaxAge: getDurationSeconds("AUTH_REFRESH_TOKEN_MAX_AGE", 604800), // 7 days
			SessionCookieName:  getEnvOrDefault("AUTH_SESSION_COOKIE_NAME", "vibefeeder_session"),
			RefreshCookieName:  getEnvOrDefault("AUTH_REFRESH_COOKIE_NAME", "vibefeeder_refresh"),
			RegistrationCode:   os.Getenv("AUTH_REGISTRATION_CODE"), // Optional - empty means open registration
		},
		Log: LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		OpenRouter: OpenRouterConfig{
			APIKey: os.Getenv("OPENROUTER_API_KEY"),
		},
		Fetcher: FetcherConfig{
			FetchInterval:       getDurationSeconds("FETCHER_INTERVAL", 300),          // 5 minutes
			SuccessInterval:     getDurationSeconds("FETCHER_SUCCESS_INTERVAL", 3600), // 1 hour
			WorkerCount:         getEnvInt("FETCHER_WORKERS", 10),
			BatchSize:           getEnvInt("FETCHER_BATCH_SIZE", 1000),
			DomainDelay:         getDurationSeconds("FETCHER_DOMAIN_DELAY", 3),
			RequestTimeout:      getDurationSeconds("FETCHER_REQUEST_TIMEOUT", 30),
			JobTimeout:          getDurationSeconds("FETCHER_JOB_TIMEOUT", 45),
			MaxArticlesPerFeed:  getEnvInt("FETCHER_MAX_ARTICLES", 100),
			MaxResponseBodySize: int64(getEnvInt("FETCHER_MAX_BODY_SIZE_MB", 2) * 1024 * 1024),
		},
		RateLimit: RateLimitConfig{
			SummaryGenerationInterval: getDurationSeconds("RATE_LIMIT_SUMMARY_INTERVAL", 30), // 30 seconds (for testing, use 300 for production)
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

// getEnvInt returns environment variable as int or default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDurationSeconds returns environment variable as duration in seconds or default
func getDurationSeconds(key string, defaultSeconds int) time.Duration {
	seconds := getEnvInt(key, defaultSeconds)
	return time.Duration(seconds) * time.Second
}

// validate checks if all required configuration values are set
func (c *Config) validate() error {
	if c.Supabase.URL == "" {
		return fmt.Errorf("SUPABASE_URL is required")
	}

	if c.Supabase.Key == "" {
		return fmt.Errorf("SUPABASE_KEY is required")
	}

	if c.OpenRouter.APIKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY is required")
	}

	return nil
}
