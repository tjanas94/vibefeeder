//go:build prod

package app

import (
	static "github.com/tjanas94/vibefeeder"
)

// setupStatic configures static file serving from embedded files
func (a *App) setupStatic() error {
	// Production: serve embedded files
	staticFS, err := static.FS()
	if err != nil {
		return err
	}
	a.Echo.StaticFS("/static", staticFS)
	return nil
}
