package fetcher

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// FeedFetcherService handles automatic feed fetching in the background
type FeedFetcherService struct {
	repo       *Repository
	logger     *slog.Logger
	httpClient *HTTPClient
	config     config.FetcherConfig
	appCtx     context.Context // Main application context for graceful shutdown

	// Rate limiting: tracks last request time per domain
	domainLastRequest map[string]time.Time
	// Mutex for thread-safe access to domainLastRequest map
	mu sync.Mutex
	// Semaphore for limiting concurrent immediate fetch requests
	immediateFetchSem chan struct{}
}

// NewFeedFetcherService creates a new feed fetcher service instance
func NewFeedFetcherService(
	dbClient *database.Client,
	logger *slog.Logger,
	cfg config.FetcherConfig,
	appCtx context.Context,
) *FeedFetcherService {
	httpClient := NewHTTPClient(HTTPClientConfig{
		Timeout:         cfg.RequestTimeout,
		FollowRedirects: false,
		Logger:          logger,
	})

	return &FeedFetcherService{
		repo:              NewRepository(dbClient),
		logger:            logger,
		httpClient:        httpClient,
		config:            cfg,
		appCtx:            appCtx,
		domainLastRequest: make(map[string]time.Time),
		immediateFetchSem: make(chan struct{}, cfg.WorkerCount),
	}
}

// Start begins the main service loop that periodically fetches feeds
func (s *FeedFetcherService) Start() {
	s.logger.Info("Starting feed fetcher service",
		"fetch_interval", s.config.FetchInterval,
		"worker_count", s.config.WorkerCount,
		"domain_delay", s.config.DomainDelay,
	)

	ticker := time.NewTicker(s.config.FetchInterval)
	defer ticker.Stop()

	// Run immediately on startup
	s.ProcessFeeds()

	for {
		select {
		case <-ticker.C:
			s.ProcessFeeds()
		case <-s.appCtx.Done():
			s.logger.Info("Feed fetcher service shutting down gracefully")
			return
		}
	}
}

// ProcessFeeds orchestrates the fetching of all due feeds
func (s *FeedFetcherService) ProcessFeeds() {
	s.logger.Info("Starting feed processing batch")

	// Clean old rate limiting entries to prevent memory leak
	// Only remove entries older than 2x DomainDelay to avoid race with FetchFeedNow
	s.mu.Lock()
	cutoff := time.Now().Add(-2 * s.config.DomainDelay)
	for domain, lastRequest := range s.domainLastRequest {
		if lastRequest.Before(cutoff) {
			delete(s.domainLastRequest, domain)
		}
	}
	s.mu.Unlock()

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

	// Create semaphore channel for worker pool
	// Buffer size = number of workers
	semaphore := make(chan struct{}, s.config.WorkerCount)

	// WaitGroup to wait for all workers to complete
	var wg sync.WaitGroup

	// Process each feed in a separate goroutine with worker limit
	for _, feed := range feeds {
		// Check if context is cancelled (graceful shutdown)
		select {
		case <-s.appCtx.Done():
			s.logger.Info("Batch processing interrupted by shutdown")
			wg.Wait() // Wait for currently running workers
			return
		default:
		}

		// Acquire semaphore slot (blocks if pool is full)
		semaphore <- struct{}{}
		wg.Add(1)

		// Launch worker goroutine
		go func(f database.PublicFeedsSelect) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore slot

			// Recover from panic to prevent one feed from crashing the service
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("Panic in feed fetcher", "feed_id", f.Id, "panic", r)
				}
			}()

			s.fetchSingleFeed(f)
		}(feed)
	}

	// Wait for all workers to complete
	wg.Wait()

	s.logger.Info("Feed processing batch completed", "processed", len(feeds))
}

// FetchFeedNow immediately fetches a specific feed asynchronously
// Uses the main application context for cancellation
func (s *FeedFetcherService) FetchFeedNow(feedID string) {
	go func() {
		// Acquire semaphore slot (blocks if pool is full)
		s.immediateFetchSem <- struct{}{}
		defer func() { <-s.immediateFetchSem }()

		s.logger.Info("Immediate fetch requested", "feed_id", feedID)

		// Recover from panic to prevent crash
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic in immediate feed fetch", "feed_id", feedID, "panic", r)
			}
		}()

		// Query database for specific feed
		feed, err := s.repo.FindFeedByID(s.appCtx, feedID)
		if err != nil {
			s.logger.Error("Failed to find feed for immediate fetch", "feed_id", feedID, "error", err)
			return
		}

		// Fetch the feed
		s.fetchSingleFeed(*feed)
	}()
}

// fetchSingleFeed processes a single feed (fetch, parse, save)
func (s *FeedFetcherService) fetchSingleFeed(feed database.PublicFeedsSelect) {
	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(s.appCtx, s.config.JobTimeout)
	defer cancel()

	s.logger.Debug("Processing feed", "feed_id", feed.Id, "url", feed.Url)

	// Parse URL and extract domain for rate limiting
	parsedURL, err := url.Parse(feed.Url)
	if err != nil {
		s.logger.Error("Invalid feed URL", "feed_id", feed.Id, "url", feed.Url, "error", err)
		errorMsg := "Invalid feed URL"
		_ = s.updateFeedStatus(jobCtx, UpdateFeedStatusParams{
			FeedID:   feed.Id,
			Status:   "permanent_error",
			ErrorMsg: &errorMsg,
		})
		return
	}

	domain := parsedURL.Hostname()

	// Apply rate limiting for this domain
	if err := s.applyRateLimit(jobCtx, domain); err != nil {
		s.logger.Debug("Rate limiting cancelled", "feed_id", feed.Id, "domain", domain, "error", err)
		return
	}

	// Execute HTTP request with conditional headers
	resp, err := s.httpClient.ExecuteRequest(jobCtx, ExecuteRequestParams{
		URL:          feed.Url,
		ETag:         feed.Etag,
		LastModified: feed.LastModified,
	})
	if err != nil {
		s.handleRequestError(jobCtx, HandleRequestErrorParams{
			FeedID:     feed.Id,
			Err:        err,
			Operation:  "HTTP request",
			RetryCount: feed.RetryCount,
		})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("failed to close response body", "error", err)
		}
	}()

	// Handle different HTTP status codes (starting with 0 redirects)
	s.handleHTTPResponse(jobCtx, HandleHTTPResponseParams{
		Feed:          feed,
		Resp:          resp,
		OriginalURL:   parsedURL,
		RedirectCount: 0,
		VisitedURLs:   nil, // Will be initialized on first redirect if needed
	})
}

// HandleRequestErrorParams contains parameters for handling HTTP request errors
type HandleRequestErrorParams struct {
	FeedID     string
	Err        error
	Operation  string
	RetryCount int
}

// handleRequestError handles HTTP request errors, distinguishing between permanent (SSRF) and temporary errors
func (s *FeedFetcherService) handleRequestError(ctx context.Context, params HandleRequestErrorParams) {
	// Check if error is SSRF validation failure
	if strings.Contains(params.Err.Error(), "security validation failed") {
		// SSRF errors are permanent - URL is malicious/blocked
		s.logger.Error(params.Operation+" failed (SSRF)", "feed_id", params.FeedID, "error", params.Err)
		errorMsg := params.Err.Error()
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:   params.FeedID,
			Status:   "permanent_error",
			ErrorMsg: &errorMsg,
		})
	} else {
		// Other errors are temporary - network issues, timeouts, etc.
		s.logger.Error(params.Operation+" failed", "feed_id", params.FeedID, "error", params.Err)
		errorMsg := fmt.Sprintf("%s failed: %v", params.Operation, params.Err)
		nextFetch := calculateBackoff(params.RetryCount, time.Now())
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:     params.FeedID,
			Status:     "temporary_error",
			ErrorMsg:   &errorMsg,
			FetchAfter: &nextFetch,
		})
	}
}

// applyRateLimit ensures minimum delay between requests to same domain
// Returns error if context is cancelled during rate limiting
func (s *FeedFetcherService) applyRateLimit(ctx context.Context, domain string) error {
	for {
		s.mu.Lock()
		lastRequest, exists := s.domainLastRequest[domain]

		if !exists {
			// No previous request, safe to proceed
			s.domainLastRequest[domain] = time.Now()
			s.mu.Unlock()
			return nil
		}

		elapsed := time.Since(lastRequest)
		if elapsed >= s.config.DomainDelay {
			// Enough time has passed, safe to proceed
			s.domainLastRequest[domain] = time.Now()
			s.mu.Unlock()
			return nil
		}

		// Calculate wait time while holding lock
		waitTime := s.config.DomainDelay - elapsed
		s.mu.Unlock()

		s.logger.Debug("Rate limiting domain", "domain", domain, "wait_time", waitTime)

		// Wait with context cancellation support
		select {
		case <-time.After(waitTime):
			// Wait time elapsed, loop back to check again atomically
			// (another goroutine might have taken the slot)
		case <-ctx.Done():
			// Context cancelled during rate limiting
			return ctx.Err()
		}
	}
}

// UpdateFeedStatusParams contains parameters for updateFeedStatus
type UpdateFeedStatusParams struct {
	FeedID       string
	Status       string
	ErrorMsg     *string
	FetchAfter   *time.Time
	Etag         *string
	LastModified *string
	NewURL       *string
}

// updateFeedStatus updates feed status in database after fetch attempt
func (s *FeedFetcherService) updateFeedStatus(ctx context.Context, params UpdateFeedStatusParams) error {
	s.logger.Debug("Updating feed status",
		"feed_id", params.FeedID,
		"status", params.Status,
		"has_error", params.ErrorMsg != nil,
		"url_changed", params.NewURL != nil,
	)

	now := time.Now().UTC().Format(time.RFC3339)

	// Build update struct
	update := database.PublicFeedsUpdate{
		LastFetchStatus: &params.Status,
		LastFetchError:  params.ErrorMsg,
		LastFetchedAt:   &now,
	}

	// Reset retry count on success
	if params.Status == "success" {
		retryCount := 0
		update.RetryCount = &retryCount
	}

	// Set fetch_after if provided (for scheduling next fetch)
	if params.FetchAfter != nil {
		fetchAfterStr := params.FetchAfter.UTC().Format(time.RFC3339)
		update.FetchAfter = &fetchAfterStr
	}

	// Set etag if provided (for conditional requests)
	if params.Etag != nil {
		update.Etag = params.Etag
	}

	// Set last_modified if provided (for conditional requests)
	if params.LastModified != nil {
		update.LastModified = params.LastModified
	}

	// Set new URL if provided (for permanent redirects)
	if params.NewURL != nil {
		update.Url = params.NewURL
	}

	// Update feed in database
	if err := s.repo.UpdateFeedAfterFetch(ctx, params.FeedID, update); err != nil {
		return fmt.Errorf("failed to update feed status: %w", err)
	}

	return nil
}

// saveNewArticles saves parsed articles to database, avoiding duplicates
func (s *FeedFetcherService) saveNewArticles(
	ctx context.Context,
	feedID string,
	articles []Article,
) error {
	if len(articles) == 0 {
		s.logger.Debug("No articles to save", "feed_id", feedID)
		return nil
	}

	s.logger.Info("Saving articles", "feed_id", feedID, "count", len(articles))

	// Transform articles to database insert format
	dbArticles := make([]database.PublicArticlesInsert, 0, len(articles))
	for _, article := range articles {
		dbArticle := database.PublicArticlesInsert{
			FeedId:      feedID,
			Title:       article.Title,
			Url:         article.URL,
			Content:     article.Content,
			PublishedAt: article.PublishedAt.UTC().Format(time.RFC3339),
		}
		dbArticles = append(dbArticles, dbArticle)
	}

	// Insert articles (duplicates are ignored by UNIQUE constraint on feed_id, url)
	if err := s.repo.InsertArticles(ctx, dbArticles); err != nil {
		return fmt.Errorf("failed to insert articles: %w", err)
	}

	s.logger.Info("Articles saved successfully", "feed_id", feedID, "count", len(articles))
	return nil
}

// Article represents a parsed feed article
type Article struct {
	Title       string
	URL         string
	Content     *string
	PublishedAt time.Time
}

// HandleHTTPResponseParams contains parameters for handleHTTPResponse
type HandleHTTPResponseParams struct {
	Feed          database.PublicFeedsSelect
	Resp          *http.Response
	OriginalURL   *url.URL
	RedirectCount int
	VisitedURLs   map[string]bool // Track visited URLs for redirect loop detection
}

// handleHTTPResponse processes the HTTP response based on status code
func (s *FeedFetcherService) handleHTTPResponse(ctx context.Context, params HandleHTTPResponseParams) {
	statusCode := params.Resp.StatusCode
	s.logger.Debug("Received HTTP response", "feed_id", params.Feed.Id, "status", statusCode, "redirects", params.RedirectCount)

	switch {
	case statusCode == http.StatusOK: // 200
		s.handleSuccess(ctx, params.Feed, params.Resp)

	case statusCode == http.StatusNotModified: // 304
		s.handleNotModified(ctx, params.Feed, params.Resp)

	case statusCode == http.StatusMovedPermanently || statusCode == http.StatusPermanentRedirect: // 301, 308
		s.handleRedirect(ctx, HandleRedirectParams{
			Feed:          params.Feed,
			Resp:          params.Resp,
			OriginalURL:   params.OriginalURL,
			RedirectCount: params.RedirectCount,
			IsPermanent:   true,
			VisitedURLs:   params.VisitedURLs,
		})

	case statusCode == http.StatusFound || statusCode == http.StatusSeeOther || statusCode == http.StatusTemporaryRedirect: // 302, 303, 307
		s.handleRedirect(ctx, HandleRedirectParams{
			Feed:          params.Feed,
			Resp:          params.Resp,
			OriginalURL:   params.OriginalURL,
			RedirectCount: params.RedirectCount,
			IsPermanent:   false,
			VisitedURLs:   params.VisitedURLs,
		})

	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden: // 401, 403
		s.handleUnauthorized(ctx, params.Feed, statusCode)

	case statusCode == http.StatusTooManyRequests: // 429
		s.handleTooManyRequests(ctx, params.Feed, params.Resp)

	case statusCode == http.StatusNotFound || statusCode == http.StatusGone: // 404, 410
		s.handleNotFound(ctx, params.Feed, statusCode)

	case statusCode >= 400 && statusCode < 500: // Other 4xx
		s.handleClientError(ctx, params.Feed, statusCode)

	case statusCode >= 500: // 5xx
		s.handleServerError(ctx, params.Feed, statusCode)

	default:
		s.logger.Warn("Unexpected HTTP status", "feed_id", params.Feed.Id, "status", statusCode)
		errorMsg := fmt.Sprintf("Unexpected HTTP status: %d", statusCode)
		nextFetch := calculateBackoff(params.Feed.RetryCount, time.Now())
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:     params.Feed.Id,
			Status:     "temporary_error",
			ErrorMsg:   &errorMsg,
			FetchAfter: &nextFetch,
		})
	}
}

// handleSuccess processes a successful 200 OK response
func (s *FeedFetcherService) handleSuccess(ctx context.Context, feed database.PublicFeedsSelect, resp *http.Response) {
	// Extract ETag and Last-Modified headers
	etag := resp.Header.Get(HeaderETag)
	lastModified := resp.Header.Get(HeaderLastModified)

	var etagPtr, lastModifiedPtr *string
	if etag != "" {
		etagPtr = &etag
	}
	if lastModified != "" {
		lastModifiedPtr = &lastModified
	}

	// Limit response body size to prevent zip bomb attacks
	limitedReader := io.LimitReader(resp.Body, s.config.MaxResponseBodySize)

	// Parse feed
	parser := gofeed.NewParser()
	parsedFeed, err := parser.Parse(limitedReader)
	if err != nil {
		s.logger.Error("Failed to parse feed", "feed_id", feed.Id, "error", err)
		errorMsg := fmt.Sprintf("Failed to parse feed: %v", err)
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:   feed.Id,
			Status:   "permanent_error",
			ErrorMsg: &errorMsg,
		})
		return
	}

	// Transform feed items to articles
	articles := transformFeedItems(parsedFeed.Items, time.Now())

	// Limit number of articles to prevent database spam
	if len(articles) > s.config.MaxArticlesPerFeed {
		s.logger.Warn("Feed has too many articles, limiting", "feed_id", feed.Id, "total", len(articles), "limit", s.config.MaxArticlesPerFeed)
		articles = articles[:s.config.MaxArticlesPerFeed]
	}

	// Save articles to database
	if err := s.saveNewArticles(ctx, feed.Id, articles); err != nil {
		s.logger.Error("Failed to save articles", "feed_id", feed.Id, "error", err)
		errorMsg := fmt.Sprintf("Failed to save articles: %v", err)
		nextFetch := calculateBackoff(feed.RetryCount, time.Now())
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:     feed.Id,
			Status:     "temporary_error",
			ErrorMsg:   &errorMsg,
			FetchAfter: &nextFetch,
		})
		return
	}

	// Calculate next fetch time
	cacheControl := resp.Header.Get(HeaderCacheControl)
	nextFetch := calculateNextFetch(cacheControl, s.config.SuccessInterval, time.Now())

	// Update feed status to success
	status := "success"
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:       feed.Id,
		Status:       status,
		FetchAfter:   &nextFetch,
		Etag:         etagPtr,
		LastModified: lastModifiedPtr,
	})

	s.logger.Info("Feed fetched successfully", "feed_id", feed.Id, "articles", len(articles), "next_fetch", nextFetch)
}

// handleNotModified processes a 304 Not Modified response
func (s *FeedFetcherService) handleNotModified(ctx context.Context, feed database.PublicFeedsSelect, resp *http.Response) {
	s.logger.Debug("Feed not modified", "feed_id", feed.Id)

	// Calculate next fetch time based on Cache-Control header
	cacheControl := resp.Header.Get(HeaderCacheControl)
	nextFetch := calculateNextFetch(cacheControl, s.config.SuccessInterval, time.Now())

	status := "success"
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:     feed.Id,
		Status:     status,
		FetchAfter: &nextFetch,
	})
}

// HandleRedirectParams contains parameters for handleRedirect
type HandleRedirectParams struct {
	Feed          database.PublicFeedsSelect
	Resp          *http.Response
	OriginalURL   *url.URL
	RedirectCount int
	IsPermanent   bool
	VisitedURLs   map[string]bool // Track visited URLs to detect redirect loops
}

// handleRedirect processes all redirect types (301/302/303/307/308) by following them
func (s *FeedFetcherService) handleRedirect(ctx context.Context, params HandleRedirectParams) {
	// Check redirect limit (max 10 redirects, same as Go's default)
	if params.RedirectCount >= 10 {
		s.logger.Error("Too many redirects", "feed_id", params.Feed.Id, "count", params.RedirectCount)
		errorMsg := "Too many redirects (stopped after 10)"
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:   params.Feed.Id,
			Status:   "permanent_error",
			ErrorMsg: &errorMsg,
		})
		return
	}

	newLocation := params.Resp.Header.Get("Location")
	if newLocation == "" {
		s.logger.Error("Redirect without Location header", "feed_id", params.Feed.Id, "permanent", params.IsPermanent)
		errorMsg := "Redirect without Location header"
		statusType := "temporary_error"
		if params.IsPermanent {
			statusType = "permanent_error"
		}
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:   params.Feed.Id,
			Status:   statusType,
			ErrorMsg: &errorMsg,
		})
		return
	}

	// Parse redirect URL (might be relative)
	redirectURL, err := params.OriginalURL.Parse(newLocation)
	if err != nil {
		s.logger.Error("Invalid redirect URL", "feed_id", params.Feed.Id, "location", newLocation, "error", err)
		errorMsg := fmt.Sprintf("Invalid redirect URL: %v", err)
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:   params.Feed.Id,
			Status:   "temporary_error",
			ErrorMsg: &errorMsg,
		})
		return
	}

	// Initialize visited URLs map on first redirect
	if params.VisitedURLs == nil {
		params.VisitedURLs = make(map[string]bool)
	}

	// Check for redirect loop
	redirectURLStr := redirectURL.String()
	if params.VisitedURLs[redirectURLStr] {
		s.logger.Error("Redirect loop detected",
			"feed_id", params.Feed.Id,
			"url", redirectURLStr,
			"redirect_count", params.RedirectCount,
		)
		errorMsg := "Redirect loop detected"
		_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
			FeedID:   params.Feed.Id,
			Status:   "permanent_error",
			ErrorMsg: &errorMsg,
		})
		return
	}

	// Mark current URL as visited
	params.VisitedURLs[params.OriginalURL.String()] = true

	// Log redirect with appropriate level
	if params.IsPermanent {
		s.logger.Info("Following permanent redirect", "feed_id", params.Feed.Id, "from", params.OriginalURL.String(), "to", redirectURL.String(), "count", params.RedirectCount+1)

		// Update feed URL immediately for permanent redirects
		// This ensures the new URL is saved even if subsequent requests fail
		newURLStr := redirectURL.String()
		if err := s.repo.UpdateFeedAfterFetch(ctx, params.Feed.Id, database.PublicFeedsUpdate{
			Url: &newURLStr,
		}); err != nil {
			s.logger.Error("Failed to update feed URL after permanent redirect", "feed_id", params.Feed.Id, "new_url", newURLStr, "error", err)
			// Continue anyway - we still want to fetch from the new URL
		}
	} else {
		s.logger.Debug("Following temporary redirect", "feed_id", params.Feed.Id, "from", params.OriginalURL.String(), "to", redirectURL.String(), "count", params.RedirectCount+1)
	}

	// Follow redirect
	redirectResp, err := s.httpClient.ExecuteRequest(ctx, ExecuteRequestParams{
		URL: redirectURL.String(),
		// Don't use conditional headers on redirects - fetch full content from new URL
	})
	if err != nil {
		s.handleRequestError(ctx, HandleRequestErrorParams{
			FeedID:     params.Feed.Id,
			Err:        err,
			Operation:  "Redirect request",
			RetryCount: params.Feed.RetryCount,
		})
		return
	}
	defer func() {
		if err := redirectResp.Body.Close(); err != nil {
			s.logger.Error("failed to close redirect response body", "error", err)
		}
	}()

	// Handle redirect response with incremented counter and visited URLs
	s.handleHTTPResponse(ctx, HandleHTTPResponseParams{
		Feed:          params.Feed,
		Resp:          redirectResp,
		OriginalURL:   redirectURL,
		RedirectCount: params.RedirectCount + 1,
		VisitedURLs:   params.VisitedURLs,
	})
}

// handleUnauthorized processes 401/403 authorization errors
func (s *FeedFetcherService) handleUnauthorized(ctx context.Context, feed database.PublicFeedsSelect, statusCode int) {
	s.logger.Warn("Feed requires authorization", "feed_id", feed.Id, "status", statusCode)

	errorMsg := fmt.Sprintf("Authorization required (HTTP %d)", statusCode)
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:   feed.Id,
		Status:   "unauthorized",
		ErrorMsg: &errorMsg,
	})
}

// handleTooManyRequests processes 429 Too Many Requests
func (s *FeedFetcherService) handleTooManyRequests(ctx context.Context, feed database.PublicFeedsSelect, resp *http.Response) {
	s.logger.Warn("Rate limited by server", "feed_id", feed.Id)

	// Parse Retry-After header to calculate next fetch time
	retryAfter := resp.Header.Get(HeaderRetryAfter)
	nextFetch := parseRetryAfter(retryAfter, feed.RetryCount, time.Now())

	errorMsg := "Rate limited by server (HTTP 429)"
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:     feed.Id,
		Status:     "temporary_error",
		ErrorMsg:   &errorMsg,
		FetchAfter: &nextFetch,
	})
}

// handleNotFound processes 404/410 not found errors
func (s *FeedFetcherService) handleNotFound(ctx context.Context, feed database.PublicFeedsSelect, statusCode int) {
	s.logger.Warn("Feed not found", "feed_id", feed.Id, "status", statusCode)

	errorMsg := fmt.Sprintf("Feed not found (HTTP %d)", statusCode)
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:   feed.Id,
		Status:   "permanent_error",
		ErrorMsg: &errorMsg,
	})
}

// handleClientError processes other 4xx client errors
func (s *FeedFetcherService) handleClientError(ctx context.Context, feed database.PublicFeedsSelect, statusCode int) {
	s.logger.Warn("Client error", "feed_id", feed.Id, "status", statusCode)

	errorMsg := fmt.Sprintf("Client error (HTTP %d)", statusCode)
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:   feed.Id,
		Status:   "permanent_error",
		ErrorMsg: &errorMsg,
	})
}

// handleServerError processes 5xx server errors
func (s *FeedFetcherService) handleServerError(ctx context.Context, feed database.PublicFeedsSelect, statusCode int) {
	s.logger.Warn("Server error", "feed_id", feed.Id, "status", statusCode)

	errorMsg := fmt.Sprintf("Server error (HTTP %d)", statusCode)
	nextFetch := calculateBackoff(feed.RetryCount, time.Now())
	_ = s.updateFeedStatus(ctx, UpdateFeedStatusParams{
		FeedID:     feed.Id,
		Status:     "temporary_error",
		ErrorMsg:   &errorMsg,
		FetchAfter: &nextFetch,
	})
}
