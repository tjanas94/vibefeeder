//go:build prod

package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	static "github.com/tjanas94/vibefeeder"
)

// setupStatic configures static file serving from embedded files
func (a *App) setupStatic() error {
	// Production: serve embedded files with cache headers middleware
	staticFS, err := static.FS()
	if err != nil {
		return err
	}

	// StaticFS automatically handles the path prefix and serves from fs.FS
	staticGroup := a.Echo.Group("/static")
	staticGroup.Use(cacheControlMiddleware)
	staticGroup.StaticFS("/", staticFS)

	// Serve all icons from embedded bytes (no cache busting)
	a.Echo.GET("/favicon.ico", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=604800")
		return c.Blob(http.StatusOK, "image/x-icon", static.FaviconData)
	})

	a.Echo.GET("/icon.svg", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=604800")
		return c.Blob(http.StatusOK, "image/svg+xml", static.IconSVGData)
	})

	a.Echo.GET("/apple-touch-icon.png", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=604800")
		return c.Blob(http.StatusOK, "image/png", static.AppleTouchIconData)
	})

	return nil
}

// cacheControlMiddleware adds cache control headers for static assets
func cacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if request has cache busting query parameter
		if c.QueryParam("v") != "" {
			// Cache versioned assets for 1 year
			c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// No cache busting - cache for shorter period and revalidate
			c.Response().Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
		}
		return next(c)
	}
}
