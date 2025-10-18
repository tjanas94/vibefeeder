package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tjanas94/vibefeeder/internal/shared/csrf"
	"github.com/tjanas94/vibefeeder/internal/shared/logger"
)

// setupMiddleware configures all application middleware
func (a *App) setupMiddleware() {
	// Core middleware
	a.Echo.Use(logger.RequestLoggerConfig(a.Logger))
	a.Echo.Use(middleware.Recover())
	a.Echo.Use(middleware.BodyLimit("2M"))

	// Security middleware
	a.Echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:",
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

// csrfMiddleware returns configured CSRF protection middleware
func (a *App) csrfMiddleware() echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:csrf_token,header:X-CSRF-Token",
		ContextKey:     csrf.EchoContextKey,
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieSecure:   a.Config.Auth.CookieSecure,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		TokenLength:    32,
		// Don't use Skipper - we need tokens to be generated for GET requests
		// The middleware will only validate tokens for unsafe methods (POST, PUT, PATCH, DELETE)
	})
}
