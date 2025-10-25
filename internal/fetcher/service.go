package fetcher

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// FetcherRepository defines the interface for fetcher data access
type FetcherRepository interface {
	FindFeedsDueForFetch(ctx context.Context, limit int) ([]database.PublicFeedsSelect, error)
	FindFeedByID(ctx context.Context, feedID string) (*database.PublicFeedsSelect, error)
	UpdateFeedAfterFetch(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error
	InsertArticles(ctx context.Context, articles []database.PublicArticlesInsert) error
}

// HTTPClientInterface defines the interface for HTTP client operations
type HTTPClientInterface interface {
	ExecuteRequest(ctx context.Context, params ExecuteRequestParams) (*http.Response, error)
}

// FeedFetcherService is the main orchestrator for feed fetching operations
// Delegates all processing to Scheduler to ensure consistency and reuse of processing pipeline
type FeedFetcherService struct {
	scheduler *Scheduler
	logger    *slog.Logger
	appCtx    context.Context
}

// NewFeedFetcherService creates a new feed fetcher service instance
// Initializes all sub-components and sets up dependencies
func NewFeedFetcherService(
	repo FetcherRepository,
	httpClient HTTPClientInterface,
	logger *slog.Logger,
	cfg config.FetcherConfig,
	appCtx context.Context,
) *FeedFetcherService {
	if logger == nil {
		logger = slog.Default()
	}

	// Create response handler
	responseHandler := NewHTTPResponseHandler(logger, cfg.SuccessInterval)

	// Create feed fetcher
	feedFetcher := NewFeedFetcher(
		httpClient,
		responseHandler,
		logger,
		cfg.MaxResponseBodySize,
		cfg.MaxArticlesPerFeed,
	)

	// Create status manager
	statusManager := NewFeedStatusManager(repo, logger)

	// Create rate limiter
	rateLimiter := NewRateLimiter(cfg.DomainDelay)

	// Create worker pool
	workerPool := NewWorkerPool(cfg.WorkerCount, logger)

	// Create scheduler
	scheduler := NewScheduler(
		repo,
		workerPool,
		rateLimiter,
		feedFetcher,
		statusManager,
		logger,
		cfg,
		appCtx,
	)

	return &FeedFetcherService{
		scheduler: scheduler,
		logger:    logger,
		appCtx:    appCtx,
	}
}

// Start begins the main service loop that periodically fetches feeds
// Delegates to the scheduler for batch processing
func (s *FeedFetcherService) Start() {
	s.scheduler.Start()
}

// FetchFeedNow immediately fetches a specific feed asynchronously
// Delegates to scheduler to reuse the same WorkerPool and processing pipeline
// Panic recovery is handled by WorkerPool, so this is just async delegation
func (s *FeedFetcherService) FetchFeedNow(feedID string) {
	go s.scheduler.FetchSingleFeedByID(feedID)
}
