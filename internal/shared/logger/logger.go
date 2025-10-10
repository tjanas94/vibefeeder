package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/tjanas94/vibefeeder/internal/shared/config"
)

// New creates a new slog logger based on configuration
func New(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	// Parse log level
	level := parseLevel(cfg.Log.Level)

	// Create handler based on format
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(cfg.Log.Format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
