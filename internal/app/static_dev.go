//go:build !prod

package app

import (
	"os"
	"path/filepath"
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
	return nil
}
