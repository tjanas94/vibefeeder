//go:build !prod

package app

import (
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// setupStatic configures static file serving from filesystem
func (a *App) setupStatic() error {
	// Development: serve files from disk relative to binary location
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	staticDir := filepath.Join(filepath.Dir(exe), "static")
	a.Echo.Static("/static", staticDir)

	// Serve all icons from root (no cache in dev)
	rootDir := filepath.Dir(exe)

	a.Echo.GET("/favicon.ico", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "image/x-icon")
		return c.File(filepath.Join(rootDir, "favicon.ico"))
	})

	a.Echo.GET("/icon.svg", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "image/svg+xml")
		return c.File(filepath.Join(rootDir, "icon.svg"))
	})

	a.Echo.GET("/apple-touch-icon.png", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "image/png")
		return c.File(filepath.Join(rootDir, "apple-touch-icon.png"))
	})

	return nil
}
