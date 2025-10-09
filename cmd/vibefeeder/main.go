package main

import (
	"context"
	"log"
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
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start server in a goroutine
	go func() {
		if err := application.Start(":8080"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the application
	if err := application.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}

	log.Println("Server exited")
}
