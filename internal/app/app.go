package app

import (
	"context"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// App holds the application dependencies
type App struct {
	Echo   *echo.Echo
	Config *config.Config
	DB     *database.Client
}

// New creates and configures a new application instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database client
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Check database health
	if err := db.Health(); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Create Echo instance
	e := echo.New()

	// Create app instance
	app := &App{
		Echo:   e,
		Config: cfg,
		DB:     db,
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
func (a *App) Start(address string) error {
	log.Printf("Starting server on %s", address)
	return a.Echo.Start(address)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	// Shutdown Echo server
	if err := a.Echo.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
		return err
	}

	log.Println("Application shut down successfully")
	return nil
}
