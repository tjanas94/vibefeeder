package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/fetcher"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/errors"
	"github.com/tjanas94/vibefeeder/internal/shared/logger"
	"github.com/tjanas94/vibefeeder/internal/shared/validator"
	"github.com/tjanas94/vibefeeder/internal/shared/view"
)

// App holds the application dependencies
type App struct {
	Echo         *echo.Echo
	Config       *config.Config
	DB           *database.Client
	Logger       *slog.Logger
	FeedFetcher  *fetcher.FeedFetcherService
	AppCtx       context.Context
	CancelAppCtx context.CancelFunc
}

// New creates and configures a new application instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log := logger.New(cfg)

	// Initialize database client
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Check database health
	if err := db.Health(); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	log.Info("Database connection established")

	// Create application context for graceful shutdown
	appCtx, cancelAppCtx := context.WithCancel(context.Background())

	// Initialize feed fetcher service
	fetcherRepo := fetcher.NewRepository(db)
	httpClient := fetcher.NewHTTPClient(fetcher.HTTPClientConfig{
		Timeout:         cfg.Fetcher.RequestTimeout,
		FollowRedirects: false,
		Logger:          log,
	})
	feedFetcher := fetcher.NewFeedFetcherService(fetcherRepo, httpClient, log, cfg.Fetcher, appCtx)

	// Create Echo instance
	e := echo.New()

	// Disable Echo's default logger banner
	e.HideBanner = true
	e.HidePort = true

	// Register Templ renderer
	e.Renderer = view.NewTemplRenderer()

	// Register validator
	e.Validator = validator.New()

	// Create app instance
	app := &App{
		Echo:         e,
		Config:       cfg,
		DB:           db,
		Logger:       log,
		FeedFetcher:  feedFetcher,
		AppCtx:       appCtx,
		CancelAppCtx: cancelAppCtx,
	}

	// Register custom error handler
	e.HTTPErrorHandler = errors.NewHTTPErrorHandler(log)

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
	a.Logger.Info("Starting server", "address", a.Config.Server.Address)
	return a.Echo.Start(a.Config.Server.Address)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("Shutting down application...")

	// Cancel application context to signal feed fetcher to stop
	a.CancelAppCtx()

	// Shutdown Echo server
	if err := a.Echo.Shutdown(ctx); err != nil {
		a.Logger.Error("Error shutting down server", "error", err)
		return err
	}

	a.Logger.Info("Application shut down successfully")
	return nil
}

// StartFeedFetcher starts the feed fetcher service in a goroutine
func (a *App) StartFeedFetcher() {
	go a.FeedFetcher.Start()
	a.Logger.Info("Feed fetcher service started")
}
