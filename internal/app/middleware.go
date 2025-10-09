package app

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// setupMiddleware configures all application middleware
func (a *App) setupMiddleware() {
	// Core middleware
	a.Echo.Use(middleware.Logger())
	a.Echo.Use(middleware.Recover())
	a.Echo.Use(middleware.BodyLimit("2M"))

	// Database middleware - inject database client into context
	a.Echo.Use(database.Middleware(a.DB))

	// Security middleware
	a.Echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// Custom headers middleware
	a.Echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			return next(c)
		}
	})
}
