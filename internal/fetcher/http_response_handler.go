package fetcher

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
)

// HTTPResponseHandler handles different HTTP response status codes and determines next actions
type HTTPResponseHandler struct {
	logger          *slog.Logger
	successInterval time.Duration
}

// NewHTTPResponseHandler creates a new HTTP response handler
func NewHTTPResponseHandler(logger *slog.Logger, successInterval time.Duration) *HTTPResponseHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &HTTPResponseHandler{
		logger:          logger,
		successInterval: successInterval,
	}
}

// HandleResponse processes an HTTP response and returns a FetchDecision
func (h *HTTPResponseHandler) HandleResponse(
	resp *http.Response,
	feedURL string,
	maxResponseBodySize int64,
	maxArticlesPerFeed int,
	retryCount int,
	currentETag *string,
	currentLastModified *string,
) FetchDecision {
	statusCode := resp.StatusCode
	h.logger.Debug("Received HTTP response", "status", statusCode)

	switch {
	case statusCode == http.StatusOK: // 200
		return h.handleSuccess(resp, feedURL, maxResponseBodySize, maxArticlesPerFeed)

	case statusCode == http.StatusNotModified: // 304
		return h.handleNotModified(resp)

	case statusCode == http.StatusMovedPermanently || statusCode == http.StatusPermanentRedirect: // 301, 308
		return h.handlePermanentRedirect(resp)

	case statusCode == http.StatusFound || statusCode == http.StatusSeeOther || statusCode == http.StatusTemporaryRedirect: // 302, 303, 307
		return h.handleTemporaryRedirect(resp)

	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden: // 401, 403
		return h.handleUnauthorized(statusCode)

	case statusCode == http.StatusTooManyRequests: // 429
		return h.handleTooManyRequests(resp, retryCount)

	case statusCode == http.StatusNotFound || statusCode == http.StatusGone: // 404, 410
		return h.handleNotFound(statusCode)

	case statusCode >= 400 && statusCode < 500: // Other 4xx
		return h.handleClientError(statusCode)

	case statusCode >= 500: // 5xx
		return h.handleServerError(resp, retryCount)

	default:
		return h.handleUnexpectedStatus(statusCode, resp, retryCount)
	}
}

// handleSuccess processes a successful 200 OK response
func (h *HTTPResponseHandler) handleSuccess(
	resp *http.Response,
	feedURL string,
	maxResponseBodySize int64,
	maxArticlesPerFeed int,
) FetchDecision {
	// Extract conditional headers
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
	limitedReader := io.LimitReader(resp.Body, maxResponseBodySize)

	// Parse feed
	parser := gofeed.NewParser()
	parsedFeed, err := parser.Parse(limitedReader)
	if err != nil {
		h.logger.Error("Failed to parse feed", "error", err)
		errorMsg := fmt.Sprintf("Failed to parse feed: %v", err)
		return FetchDecision{
			Status:       "permanent_error",
			ErrorMessage: &errorMsg,
		}
	}

	// Transform feed items to articles
	articles := transformFeedItems(parsedFeed.Items, time.Now())

	// Limit number of articles to prevent database spam
	if len(articles) > maxArticlesPerFeed {
		h.logger.Warn("Feed has too many articles, limiting", "total", len(articles), "limit", maxArticlesPerFeed)
		articles = articles[:maxArticlesPerFeed]
	}

	// Calculate next fetch time
	cacheControl := resp.Header.Get(HeaderCacheControl)
	nextFetch := calculateNextFetch(cacheControl, h.successInterval, time.Now())

	h.logger.Info("Feed fetched successfully", "articles", len(articles), "next_fetch", nextFetch)

	return FetchDecision{
		ShouldRetry:   false,
		NextFetchTime: nextFetch,
		Status:        "success",
		Articles:      articles,
		ETag:          etagPtr,
		LastModified:  lastModifiedPtr,
	}
}

// handleNotModified processes a 304 Not Modified response
func (h *HTTPResponseHandler) handleNotModified(resp *http.Response) FetchDecision {
	h.logger.Debug("Feed not modified")

	// Calculate next fetch time based on Cache-Control header
	cacheControl := resp.Header.Get(HeaderCacheControl)
	nextFetch := calculateNextFetch(cacheControl, h.successInterval, time.Now())

	return FetchDecision{
		ShouldRetry:   false,
		NextFetchTime: nextFetch,
		Status:        "success",
		Articles:      []Article{},
	}
}

// handlePermanentRedirect processes 301/308 permanent redirects
func (h *HTTPResponseHandler) handlePermanentRedirect(resp *http.Response) FetchDecision {
	location := resp.Header.Get("Location")
	if location == "" {
		errorMsg := "Redirect without Location header"
		h.logger.Error("Redirect without Location header", "status", resp.StatusCode)
		return FetchDecision{
			Status:       "permanent_error",
			ErrorMessage: &errorMsg,
		}
	}

	h.logger.Info("Following permanent redirect", "location", location)

	return FetchDecision{
		ShouldRetry:   true,
		NextFetchTime: time.Now(),
		Status:        "redirect",
		NewURL:        &location,
	}
}

// handleTemporaryRedirect processes 302/303/307 temporary redirects
func (h *HTTPResponseHandler) handleTemporaryRedirect(resp *http.Response) FetchDecision {
	location := resp.Header.Get("Location")
	if location == "" {
		errorMsg := "Redirect without Location header"
		h.logger.Error("Redirect without Location header", "status", resp.StatusCode)
		return FetchDecision{
			Status:       "permanent_error",
			ErrorMessage: &errorMsg,
		}
	}

	h.logger.Debug("Following temporary redirect", "location", location)

	return FetchDecision{
		ShouldRetry:   true,
		NextFetchTime: time.Now(),
		Status:        "redirect",
		NewURL:        &location,
	}
}

// handleUnauthorized processes 401/403 authorization errors
func (h *HTTPResponseHandler) handleUnauthorized(statusCode int) FetchDecision {
	h.logger.Warn("Feed requires authorization", "status", statusCode)

	errorMsg := fmt.Sprintf("Authorization required (HTTP %d)", statusCode)
	return FetchDecision{
		Status:       "unauthorized",
		ErrorMessage: &errorMsg,
	}
}

// handleTooManyRequests processes 429 Too Many Requests
func (h *HTTPResponseHandler) handleTooManyRequests(resp *http.Response, retryCount int) FetchDecision {
	h.logger.Warn("Rate limited by server")

	// Parse Retry-After header to calculate next fetch time
	retryAfter := resp.Header.Get(HeaderRetryAfter)
	nextFetch := parseRetryAfter(retryAfter, retryCount, time.Now())

	errorMsg := "Rate limited by server (HTTP 429)"
	return FetchDecision{
		ShouldRetry:   false,
		NextFetchTime: nextFetch,
		Status:        "temporary_error",
		ErrorMessage:  &errorMsg,
	}
}

// handleNotFound processes 404/410 not found errors
func (h *HTTPResponseHandler) handleNotFound(statusCode int) FetchDecision {
	h.logger.Warn("Feed not found", "status", statusCode)

	errorMsg := fmt.Sprintf("Feed not found (HTTP %d)", statusCode)
	return FetchDecision{
		Status:       "permanent_error",
		ErrorMessage: &errorMsg,
	}
}

// handleClientError processes other 4xx client errors
func (h *HTTPResponseHandler) handleClientError(statusCode int) FetchDecision {
	h.logger.Warn("Client error", "status", statusCode)

	errorMsg := fmt.Sprintf("Client error (HTTP %d)", statusCode)
	return FetchDecision{
		Status:       "permanent_error",
		ErrorMessage: &errorMsg,
	}
}

// handleServerError processes 5xx server errors
func (h *HTTPResponseHandler) handleServerError(resp *http.Response, retryCount int) FetchDecision {
	h.logger.Warn("Server error", "status", resp.StatusCode)

	retryAfter := resp.Header.Get(HeaderRetryAfter)
	nextFetch := parseRetryAfter(retryAfter, retryCount, time.Now())

	errorMsg := fmt.Sprintf("Server error (HTTP %d)", resp.StatusCode)
	return FetchDecision{
		ShouldRetry:   false,
		NextFetchTime: nextFetch,
		Status:        "temporary_error",
		ErrorMessage:  &errorMsg,
	}
}

// handleUnexpectedStatus processes unexpected HTTP status codes
func (h *HTTPResponseHandler) handleUnexpectedStatus(statusCode int, resp *http.Response, retryCount int) FetchDecision {
	h.logger.Warn("Unexpected HTTP status", "status", statusCode)

	retryAfter := resp.Header.Get(HeaderRetryAfter)
	nextFetch := parseRetryAfter(retryAfter, retryCount, time.Now())

	errorMsg := fmt.Sprintf("Unexpected HTTP status: %d", statusCode)
	return FetchDecision{
		ShouldRetry:   false,
		NextFetchTime: nextFetch,
		Status:        "temporary_error",
		ErrorMessage:  &errorMsg,
	}
}

// ValidateURL validates and resolves redirect URLs
func ValidateURL(baseURL *url.URL, newLocation string) (*url.URL, error) {
	redirectURL, err := baseURL.Parse(newLocation)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URL: %w", err)
	}
	return redirectURL, nil
}
