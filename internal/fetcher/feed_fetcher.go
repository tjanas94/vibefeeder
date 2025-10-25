package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// FeedFetcher handles HTTP requests and feed parsing
type FeedFetcher struct {
	httpClient       HTTPClientInterface
	responseHandler  *HTTPResponseHandler
	logger           *slog.Logger
	maxBodySize      int64
	maxArticlesCount int
}

// NewFeedFetcher creates a new feed fetcher
func NewFeedFetcher(
	httpClient HTTPClientInterface,
	responseHandler *HTTPResponseHandler,
	logger *slog.Logger,
	maxBodySize int64,
	maxArticlesCount int,
) *FeedFetcher {
	if logger == nil {
		logger = slog.Default()
	}

	return &FeedFetcher{
		httpClient:       httpClient,
		responseHandler:  responseHandler,
		logger:           logger,
		maxBodySize:      maxBodySize,
		maxArticlesCount: maxArticlesCount,
	}
}

// Fetch executes an HTTP request for a feed and returns a FetchDecision
// Handles redirects internally with loop detection
func (ff *FeedFetcher) Fetch(
	ctx context.Context,
	feed database.PublicFeedsSelect,
	retryCount int,
) FetchDecision {
	parsedURL, err := url.Parse(feed.Url)
	if err != nil {
		ff.logger.Error("Invalid feed URL", "feed_id", feed.Id, "url", feed.Url, "error", err)
		errorMsg := "Invalid feed URL"
		return FetchDecision{
			Status:       "permanent_error",
			ErrorMessage: &errorMsg,
		}
	}

	// Start redirect chain
	visitedURLs := make(map[string]bool)
	return ff.fetchURL(ctx, feed, parsedURL, feed.Url, retryCount, 0, visitedURLs, nil)
}

// fetchURL recursively follows redirects
func (ff *FeedFetcher) fetchURL(
	ctx context.Context,
	feed database.PublicFeedsSelect,
	baseURL *url.URL,
	currentURL string,
	retryCount int,
	redirectCount int,
	visitedURLs map[string]bool,
	permanentRedirectURL *string,
) FetchDecision {
	// Check redirect limit
	if redirectCount >= 10 {
		ff.logger.Error("Too many redirects", "feed_id", feed.Id, "count", redirectCount)
		errorMsg := "Too many redirects (stopped after 10)"
		return FetchDecision{
			Status:       "permanent_error",
			ErrorMessage: &errorMsg,
		}
	}

	// Check for redirect loop
	if visitedURLs[currentURL] {
		ff.logger.Error("Redirect loop detected", "feed_id", feed.Id, "url", currentURL)
		errorMsg := "Redirect loop detected"
		return FetchDecision{
			Status:       "permanent_error",
			ErrorMessage: &errorMsg,
		}
	}

	visitedURLs[currentURL] = true

	// Execute HTTP request
	resp, err := ff.httpClient.ExecuteRequest(ctx, ExecuteRequestParams{
		URL:          currentURL,
		ETag:         feed.Etag,
		LastModified: feed.LastModified,
	})
	if err != nil {
		ff.logger.Error("HTTP request failed", "feed_id", feed.Id, "url", currentURL, "error", err)

		// Check if error is SSRF validation failure
		if isSSRFError(err) {
			errorMsg := err.Error()
			return FetchDecision{
				Status:       "permanent_error",
				ErrorMessage: &errorMsg,
			}
		}

		// Other errors are temporary
		newRetryCount := retryCount + 1
		status := "temporary_error"
		if newRetryCount >= 10 {
			status = "permanent_error"
			ff.logger.Warn("Feed reached max retry count", "feed_id", feed.Id, "retry_count", newRetryCount)
		}

		nextFetch := calculateBackoff(newRetryCount, time.Now())

		errMsg := err.Error()
		return FetchDecision{
			ShouldRetry:   false,
			NextFetchTime: nextFetch,
			Status:        status,
			ErrorMessage:  &errMsg,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			ff.logger.Error("failed to close response body", "error", err)
		}
	}()

	// Handle HTTP response
	decision := ff.responseHandler.HandleResponse(
		resp,
		currentURL,
		ff.maxBodySize,
		ff.maxArticlesCount,
		retryCount,
		feed.Etag,
		feed.LastModified,
	)

	// Handle redirects
	if decision.Status == "redirect" && decision.NewURL != nil {
		redirectURL, err := ValidateURL(baseURL, *decision.NewURL)
		if err != nil {
			ff.logger.Error("Invalid redirect URL", "feed_id", feed.Id, "location", *decision.NewURL, "error", err)
			errorMsg := fmt.Sprintf("Invalid redirect URL: %v", err)
			return FetchDecision{
				Status:       "permanent_error",
				ErrorMessage: &errorMsg,
			}
		}

		ff.logger.Debug("Following redirect", "from", currentURL, "to", redirectURL.String(), "count", redirectCount+1)

		// Track permanent redirects (301/308) - always update to get the final destination URL
		newPermanentRedirectURL := permanentRedirectURL
		if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusPermanentRedirect {
			newPermanentRedirectURL = decision.NewURL
		}

		// Recursively follow redirect
		return ff.fetchURL(ctx, feed, redirectURL, redirectURL.String(), retryCount, redirectCount+1, visitedURLs, newPermanentRedirectURL)
	}

	// Add permanent redirect URL to final decision if one was found
	if permanentRedirectURL != nil {
		decision.NewURL = permanentRedirectURL
	}

	return decision
}

// isSSRFError checks if an error is from SSRF validation
func isSSRFError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "security validation failed")
}

// transformFeedItems transforms gofeed items to our Article format
// Pure function kept at package level for easy testing
func transformFeedItems(items []*gofeed.Item, now time.Time) []Article {
	articles := make([]Article, 0, len(items))

	for _, item := range items {
		// Skip items without required fields
		if item.Title == "" || item.Link == "" {
			continue
		}

		// Parse published date
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = *item.UpdatedParsed
		} else {
			publishedAt = now
		}

		// Use description or content
		var content *string
		if item.Description != "" {
			content = &item.Description
		} else if item.Content != "" {
			content = &item.Content
		}

		article := Article{
			Title:       item.Title,
			URL:         item.Link,
			Content:     content,
			PublishedAt: publishedAt,
		}

		articles = append(articles, article)
	}

	return articles
}
