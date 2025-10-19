package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tjanas94/vibefeeder/internal/app"
	"github.com/tjanas94/vibefeeder/internal/container"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg)

	// Initialize database client
	db, err := database.New(cfg)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Check database health
	if err := db.Health(); err != nil {
		log.Error("Database health check failed", "error", err)
		os.Exit(1)
	}

	log.Info("Database connection established")

	// Create application context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create dependency injection container
	c, err := container.New(cfg, db, log, ctx)
	if err != nil {
		log.Error("Failed to initialize container", "error", err)
		os.Exit(1)
	}

	// Create application instance
	application, err := app.New(c)
	if err != nil {
		log.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Start feed fetcher service in background
	go c.FeedFetcher.Start()
	log.Info("Feed fetcher service started")

	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start HTTP server in a goroutine
	go func() {
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error("Server failed to start", "error", err)
		// Cancel context before exiting to signal background services
		cancel()
		os.Exit(1)
	case sig := <-quit:
		log.Info("Received shutdown signal", "signal", sig)
	}

	// Cancel context to signal background services (fetcher) to stop
	cancel()
	log.Info("Shutdown signal sent to background services")

	// Create shutdown context with timeout for HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Gracefully shutdown the HTTP server
	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to shutdown gracefully", "error", err)
		os.Exit(1)
	}

	log.Info("Application exited cleanly")
}
