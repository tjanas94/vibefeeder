package app

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/container"
	"github.com/tjanas94/vibefeeder/internal/shared/errors"
	"github.com/tjanas94/vibefeeder/internal/shared/validator"
	"github.com/tjanas94/vibefeeder/internal/shared/view"
)

// App holds the Echo server and container
type App struct {
	Echo      *echo.Echo
	Container *container.Container
}

// New creates and configures a new application instance
func New(c *container.Container) (*App, error) {
	// Create Echo instance
	e := echo.New()

	// Disable Echo's default logger banner
	e.HideBanner = true
	e.HidePort = true

	// Register Templ renderer
	e.Renderer = view.NewTemplRenderer()

	// Register validator
	e.Validator = validator.New()

	// Register custom error handler
	e.HTTPErrorHandler = errors.NewHTTPErrorHandler(c.Logger)

	// Create app instance
	app := &App{
		Echo:      e,
		Container: c,
	}

	// Setup middleware
	app.setupMiddleware()

	// Setup static files
	if err := app.setupStatic(); err != nil {
		return nil, fmt.Errorf("failed to setup static files: %w", err)
	}

	// Setup routes
	app.setupRoutes()

	return app, nil
}

// Start starts the HTTP server
func (a *App) Start() error {
	a.Container.Logger.Info("Starting server", "address", a.Container.Config.Server.Address)
	return a.Echo.Start(a.Container.Config.Server.Address)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.Container.Logger.Info("Shutting down HTTP server...")

	// Shutdown Echo server
	if err := a.Echo.Shutdown(ctx); err != nil {
		a.Container.Logger.Error("Error shutting down server", "error", err)
		return err
	}

	a.Container.Logger.Info("HTTP server shut down successfully")
	return nil
}
