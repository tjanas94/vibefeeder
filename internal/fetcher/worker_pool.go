package fetcher

import (
	"context"
	"log/slog"
	"sync"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// WorkerPool manages concurrent processing with a limited number of workers
// Using generic-like approach with type assertion for feeds
type WorkerPool struct {
	workerCount int
	semaphore   chan struct{}
	logger      *slog.Logger
}

// NewWorkerPool creates a new worker pool with specified worker count
func NewWorkerPool(workerCount int, logger *slog.Logger) *WorkerPool {
	if logger == nil {
		logger = slog.Default()
	}

	return &WorkerPool{
		workerCount: workerCount,
		semaphore:   make(chan struct{}, workerCount),
		logger:      logger,
	}
}

// ProcessFeeds processes a list of feeds with a processing function
// Processing respects the worker limit and context cancellation
// Each worker is protected against panics
func (wp *WorkerPool) ProcessFeeds(ctx context.Context, feeds []database.PublicFeedsSelect, processFn func(database.PublicFeedsSelect)) {
	var wg sync.WaitGroup

	for _, feed := range feeds {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			wp.logger.Debug("Worker pool processing interrupted by context cancellation")
			wg.Wait() // Wait for currently running workers
			return
		default:
		}

		// Acquire semaphore slot
		wp.semaphore <- struct{}{}
		wg.Add(1)

		// Launch worker goroutine
		go func(f database.PublicFeedsSelect) {
			defer wg.Done()
			defer func() { <-wp.semaphore }() // Release semaphore slot

			// Recover from panic to prevent one feed from crashing the pool
			defer func() {
				if r := recover(); r != nil {
					wp.logger.Error("Panic in worker pool", "feed_id", f.Id, "panic", r)
				}
			}()

			processFn(f)
		}(feed)
	}

	// Wait for all workers to complete
	wg.Wait()
}
