package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tjanas94/vibefeeder/internal/app"
)

func main() {
	// Initialize application
	application, err := app.New()
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	logger := application.Logger

	// Start feed fetcher service
	application.StartFeedFetcher()

	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := application.Start(); err != nil {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	case <-quit:
		logger.Info("Received shutdown signal")
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the application
	if err := application.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown gracefully", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")
}
