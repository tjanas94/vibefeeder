package logger

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// RequestLoggerConfig returns Echo's RequestLogger middleware configured for slog
func RequestLoggerConfig(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogHost:     true,
		HandleError: true, // forwards error to the global error handler
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Determine log level based on status code
			level := slog.LevelInfo
			if v.Status >= 500 {
				level = slog.LevelError
			} else if v.Status >= 400 {
				level = slog.LevelWarn
			}

			// Build log attributes
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("host", v.Host),
			}

			// Add error if present
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
			}

			logger.LogAttrs(context.Background(), level, "HTTP request", attrs...)
			return nil
		},
	})
}
