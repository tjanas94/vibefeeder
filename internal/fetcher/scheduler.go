package fetcher

import (
	"context"
	"log/slog"
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// Scheduler orchestrates batch feed processing with timing and worker coordination
type Scheduler struct {
	repo            FetcherRepository
	workerPool      *WorkerPool
	rateLimiter     *RateLimiter
	feedFetcher     *FeedFetcher
	statusManager   *FeedStatusManager
	logger          *slog.Logger
	config          config.FetcherConfig
	appCtx          context.Context
	cleanupInterval time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler(
	repo FetcherRepository,
	workerPool *WorkerPool,
	rateLimiter *RateLimiter,
	feedFetcher *FeedFetcher,
	statusManager *FeedStatusManager,
	logger *slog.Logger,
	cfg config.FetcherConfig,
	appCtx context.Context,
) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Scheduler{
		repo:            repo,
		workerPool:      workerPool,
		rateLimiter:     rateLimiter,
		feedFetcher:     feedFetcher,
		statusManager:   statusManager,
		logger:          logger,
		config:          cfg,
		appCtx:          appCtx,
		cleanupInterval: 2 * cfg.DomainDelay, // Clean old entries 2x DomainDelay
	}
}

// Start begins the main scheduling loop
func (s *Scheduler) Start() {
	s.logger.Info("Starting feed fetcher scheduler",
		"fetch_interval", s.config.FetchInterval,
		"worker_count", s.config.WorkerCount,
		"domain_delay", s.config.DomainDelay,
	)

	ticker := time.NewTicker(s.config.FetchInterval)
	defer ticker.Stop()

	// Run immediately on startup
	s.ProcessBatch()

	for {
		select {
		case <-ticker.C:
			s.ProcessBatch()
		case <-s.appCtx.Done():
			s.logger.Info("Feed fetcher scheduler shutting down gracefully")
			return
		}
	}
}

// ProcessBatch fetches and processes all feeds due for fetching
func (s *Scheduler) ProcessBatch() {
	s.logger.Info("Starting feed processing batch")

	// Clean old rate limiting entries to prevent memory leak
	s.rateLimiter.CleanOldEntries(time.Now().Add(-s.cleanupInterval))

	// Query database for feeds ready to be fetched
	feeds, err := s.repo.FindFeedsDueForFetch(s.appCtx, s.config.BatchSize)
	if err != nil {
		s.logger.Error("Failed to find feeds due for fetch", "error", err)
		return
	}

	if len(feeds) == 0 {
		s.logger.Info("No feeds due for fetch")
		return
	}

	s.logger.Info("Processing feeds", "count", len(feeds))

	// Process feeds with worker pool
	s.workerPool.ProcessFeeds(s.appCtx, feeds, func(feed database.PublicFeedsSelect) {
		s.processSingleFeed(feed)
	})

	s.logger.Info("Feed processing batch completed", "processed", len(feeds))
}

// FetchSingleFeedByID fetches a specific feed by ID immediately
// Uses the same WorkerPool and processing pipeline as batch processing
// This ensures consistency: rate limiting, timeouts, and concurrency control
// The immediate fetch respects the worker pool limit and may wait if all workers are busy
func (s *Scheduler) FetchSingleFeedByID(feedID string) {
	s.logger.Info("Immediate fetch requested", "feed_id", feedID)

	// Query database for specific feed
	feed, err := s.repo.FindFeedByID(s.appCtx, feedID)
	if err != nil {
		s.logger.Error("Failed to find feed for immediate fetch", "feed_id", feedID, "error", err)
		return
	}

	// Use WorkerPool to respect concurrency limits
	// This ensures immediate fetch doesn't overwhelm the system
	s.workerPool.ProcessFeeds(s.appCtx, []database.PublicFeedsSelect{*feed},
		func(f database.PublicFeedsSelect) {
			s.processSingleFeed(f) // Reuse existing processing logic
		})
}

// processSingleFeed processes a single feed
// Shared by both batch processing and immediate fetch
// Implements the complete processing pipeline: rate limiting → fetching → decision application
func (s *Scheduler) processSingleFeed(feed database.PublicFeedsSelect) {
	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(s.appCtx, s.config.JobTimeout)
	defer cancel()

	s.logger.Debug("Processing feed", "feed_id", feed.Id, "url", feed.Url)

	// Apply rate limiting for this domain
	if err := s.rateLimiter.WaitIfNeeded(jobCtx, feed.Url); err != nil {
		s.logger.Debug("Rate limiting cancelled", "feed_id", feed.Id, "error", err)
		return
	}

	// Fetch the feed
	decision := s.feedFetcher.Fetch(jobCtx, feed, feed.RetryCount)

	// Apply decision to database
	if err := s.statusManager.ApplyDecision(jobCtx, feed, decision); err != nil {
		s.logger.Error("Failed to apply fetch decision", "feed_id", feed.Id, "error", err)
	}
}
