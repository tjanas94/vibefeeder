package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// MockHTTPClient implements HTTPClientInterface for testing
type MockHTTPClient struct {
	ExecuteRequestFunc func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error)
}

func (m *MockHTTPClient) ExecuteRequest(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
	if m.ExecuteRequestFunc != nil {
		return m.ExecuteRequestFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

// TestNewFeedFetcher tests FeedFetcher constructor
func TestNewFeedFetcher(t *testing.T) {
	tests := []struct {
		name             string
		httpClient       HTTPClientInterface
		responseHandler  *HTTPResponseHandler
		logger           *slog.Logger
		maxBodySize      int64
		maxArticlesCount int
		expectNilLogger  bool
		description      string
	}{
		{
			name:             "creates fetcher with all parameters",
			httpClient:       &MockHTTPClient{},
			responseHandler:  NewHTTPResponseHandler(slog.Default(), 15*time.Minute),
			logger:           slog.Default(),
			maxBodySize:      10 * 1024 * 1024,
			maxArticlesCount: 100,
			expectNilLogger:  false,
			description:      "Should create FeedFetcher with provided logger",
		},
		{
			name:             "uses default logger when nil",
			httpClient:       &MockHTTPClient{},
			responseHandler:  NewHTTPResponseHandler(nil, 15*time.Minute),
			logger:           nil,
			maxBodySize:      10 * 1024 * 1024,
			maxArticlesCount: 100,
			expectNilLogger:  false,
			description:      "Should use slog.Default() when logger is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := NewFeedFetcher(tt.httpClient, tt.responseHandler, tt.logger, tt.maxBodySize, tt.maxArticlesCount)

			assert.NotNil(t, ff, tt.description)
			assert.NotNil(t, ff.logger, "logger should never be nil")
			assert.Equal(t, tt.maxBodySize, ff.maxBodySize)
			assert.Equal(t, tt.maxArticlesCount, ff.maxArticlesCount)
		})
	}
}

// TestFetchInvalidURL tests Fetch with invalid feed URL
func TestFetchInvalidURL(t *testing.T) {
	ff := NewFeedFetcher(
		&MockHTTPClient{},
		NewHTTPResponseHandler(slog.Default(), 15*time.Minute),
		slog.Default(),
		10*1024*1024,
		100,
	)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "not a valid url://",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "permanent_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
	assert.Contains(t, *decision.ErrorMessage, "Invalid feed URL")
}

// TestFetchSuccess tests successful fetch with valid RSS feed
func TestFetchSuccess(t *testing.T) {
	// Create a minimal valid RSS feed
	validFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>Test Description</description>
    <item>
      <title>Test Article</title>
      <link>https://example.com/article</link>
      <description>Article description</description>
      <pubDate>Mon, 15 Jan 2024 10:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`

	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"application/rss+xml"},
				},
				Body: io.NopCloser(strings.NewReader(validFeed)),
			}, nil
		},
	}

	handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)
	ff := NewFeedFetcher(mockClient, handler, slog.Default(), 10*1024*1024, 100)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "success", decision.Status)
	assert.Nil(t, decision.ErrorMessage)
	assert.Greater(t, len(decision.Articles), 0)
}

// TestFetchSSRFError tests Fetch with SSRF validation error
func TestFetchSSRFError(t *testing.T) {
	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			return nil, errors.New("security validation failed: private IP detected")
		},
	}

	ff := NewFeedFetcher(
		mockClient,
		NewHTTPResponseHandler(slog.Default(), 15*time.Minute),
		slog.Default(),
		10*1024*1024,
		100,
	)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "http://localhost:8080/feed.xml",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "permanent_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
	assert.Contains(t, *decision.ErrorMessage, "security validation failed")
}

// TestFetchNetworkError tests Fetch with temporary network error
func TestFetchNetworkError(t *testing.T) {
	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			return nil, errors.New("connection timeout")
		},
	}

	ff := NewFeedFetcher(
		mockClient,
		NewHTTPResponseHandler(slog.Default(), 15*time.Minute),
		slog.Default(),
		10*1024*1024,
		100,
	)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "temporary_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
	assert.False(t, decision.ShouldRetry)
}

// TestFetchMaxRetryCount tests reaching max retry limit
func TestFetchMaxRetryCount(t *testing.T) {
	callCount := 0
	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			callCount++
			return nil, errors.New("connection timeout")
		},
	}

	ff := NewFeedFetcher(
		mockClient,
		NewHTTPResponseHandler(slog.Default(), 15*time.Minute),
		slog.Default(),
		10*1024*1024,
		100,
	)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	// retryCount=9 means we've already failed 9 times, this is the 10th attempt
	decision := ff.Fetch(context.Background(), feed, 9)

	assert.Equal(t, "permanent_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
}

// TestFetchURLTooManyRedirects tests redirect limit (10)
func TestFetchURLTooManyRedirects(t *testing.T) {
	callCount := 0
	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			callCount++
			// Return unique URLs to avoid loop detection, eventually hitting limit
			nextNum := callCount + 1
			location := fmt.Sprintf("https://example.com/feed-%d", nextNum)
			return &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header:     http.Header{"Location": []string{location}},
				Body:       http.NoBody,
			}, nil
		},
	}

	handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)
	ff := NewFeedFetcher(mockClient, handler, slog.Default(), 10*1024*1024, 100)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "permanent_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
	assert.Contains(t, *decision.ErrorMessage, "Too many redirects")
}

// TestFetchRedirectLoop tests detection of redirect loops
func TestFetchRedirectLoop(t *testing.T) {
	redirects := map[string]string{
		"https://example.com/feed1": "https://example.com/feed2",
		"https://example.com/feed2": "https://example.com/feed1",
	}

	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			location := redirects[params.URL]
			return &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header:     http.Header{"Location": []string{location}},
				Body:       http.NoBody,
			}, nil
		},
	}

	handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)
	ff := NewFeedFetcher(mockClient, handler, slog.Default(), 10*1024*1024, 100)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed1",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "permanent_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
	assert.Contains(t, *decision.ErrorMessage, "Redirect loop detected")
}

// TestFetchPermanentRedirectTracking tests that permanent redirect URL is tracked through chain
func TestFetchPermanentRedirectTracking(t *testing.T) {
	// Simulate: 302 (temp) -> 301 (perm) -> 200 (success)
	validFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>Test Description</description>
  </channel>
</rss>`

	requestCount := 0
	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			requestCount++
			switch requestCount {
			case 1: // Initial request
				return &http.Response{
					StatusCode: http.StatusFound, // 302
					Header:     http.Header{"Location": []string{"https://example.com/feed-v2"}},
					Body:       http.NoBody,
				}, nil
			case 2: // After 302
				return &http.Response{
					StatusCode: http.StatusMovedPermanently, // 301
					Header:     http.Header{"Location": []string{"https://example.com/feed-final"}},
					Body:       http.NoBody,
				}, nil
			case 3: // Final successful request
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/rss+xml"}},
					Body:       io.NopCloser(strings.NewReader(validFeed)),
				}, nil
			}
			return nil, errors.New("unexpected request")
		},
	}

	handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)
	ff := NewFeedFetcher(mockClient, handler, slog.Default(), 10*1024*1024, 100)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "success", decision.Status)
	assert.NotNil(t, decision.NewURL)
	assert.Equal(t, "https://example.com/feed-final", *decision.NewURL)
}

// TestFetchInvalidRedirectURL tests handling of invalid redirect URLs
func TestFetchInvalidRedirectURL(t *testing.T) {
	mockClient := &MockHTTPClient{
		ExecuteRequestFunc: func(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
			invalidLocation := "ht!tp://invalid url"
			return &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header:     http.Header{"Location": []string{invalidLocation}},
				Body:       http.NoBody,
			}, nil
		},
	}

	handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)
	ff := NewFeedFetcher(mockClient, handler, slog.Default(), 10*1024*1024, 100)

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed",
	}

	decision := ff.Fetch(context.Background(), feed, 0)

	assert.Equal(t, "permanent_error", decision.Status)
	assert.NotNil(t, decision.ErrorMessage)
	assert.Contains(t, *decision.ErrorMessage, "Invalid redirect URL")
}

// TestIsSSRFError tests SSRF error detection
func TestIsSSRFError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectSSRF  bool
		description string
	}{
		{
			name:        "error with SSRF message",
			err:         errors.New("security validation failed: private IP detected"),
			expectSSRF:  true,
			description: "Should detect SSRF error from message",
		},
		{
			name:        "generic network error",
			err:         errors.New("connection timeout"),
			expectSSRF:  false,
			description: "Should not treat generic error as SSRF",
		},
		{
			name:        "nil error",
			err:         nil,
			expectSSRF:  false,
			description: "Should handle nil error safely",
		},
		{
			name:        "wrapped SSRF error",
			err:         errors.New("wrapped: security validation failed: localhost"),
			expectSSRF:  true,
			description: "Should detect SSRF in wrapped errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSSRFError(tt.err)
			assert.Equal(t, tt.expectSSRF, result, tt.description)
		})
	}
}

// TestTransformFeedItems tests feed item transformation to Article format
func TestTransformFeedItems(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	publishedTime := time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 14, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		items       []*gofeed.Item
		wantCount   int
		validate    func(t *testing.T, articles []Article)
		description string
	}{
		{
			name: "valid items with all fields",
			items: []*gofeed.Item{
				{
					Title:           "Article 1",
					Link:            "https://example.com/article1",
					Description:     "Description 1",
					PublishedParsed: &publishedTime,
				},
				{
					Title:           "Article 2",
					Link:            "https://example.com/article2",
					Content:         "Content 2",
					PublishedParsed: &publishedTime,
				},
			},
			wantCount: 2,
			validate: func(t *testing.T, articles []Article) {
				assert.Equal(t, "Article 1", articles[0].Title)
				assert.Equal(t, "https://example.com/article1", articles[0].URL)
				require.NotNil(t, articles[0].Content)
				assert.Equal(t, "Description 1", *articles[0].Content)
				assert.Equal(t, publishedTime, articles[0].PublishedAt)

				assert.Equal(t, "Article 2", articles[1].Title)
				require.NotNil(t, articles[1].Content)
				assert.Equal(t, "Content 2", *articles[1].Content)
			},
			description: "Should transform all valid items correctly",
		},
		{
			name: "item without published date uses updated date",
			items: []*gofeed.Item{
				{
					Title:         "Article",
					Link:          "https://example.com/article",
					UpdatedParsed: &updatedTime,
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, articles []Article) {
				assert.Equal(t, updatedTime, articles[0].PublishedAt)
			},
			description: "Should fallback to updated date when published is missing",
		},
		{
			name: "item without any date uses current time",
			items: []*gofeed.Item{
				{
					Title: "Article",
					Link:  "https://example.com/article",
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, articles []Article) {
				assert.Equal(t, now, articles[0].PublishedAt)
			},
			description: "Should use current time when no dates available",
		},
		{
			name: "item with both description and content prefers description",
			items: []*gofeed.Item{
				{
					Title:           "Article",
					Link:            "https://example.com/article",
					Description:     "Description text",
					Content:         "Content text",
					PublishedParsed: &publishedTime,
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, articles []Article) {
				require.NotNil(t, articles[0].Content)
				assert.Equal(t, "Description text", *articles[0].Content)
			},
			description: "Should prefer description over content",
		},
		{
			name: "item without title is skipped",
			items: []*gofeed.Item{
				{
					Link:            "https://example.com/article",
					Description:     "Description",
					PublishedParsed: &publishedTime,
				},
			},
			wantCount:   0,
			validate:    func(t *testing.T, articles []Article) {},
			description: "Should skip items without title",
		},
		{
			name: "item without link is skipped",
			items: []*gofeed.Item{
				{
					Title:           "Article",
					Description:     "Description",
					PublishedParsed: &publishedTime,
				},
			},
			wantCount:   0,
			validate:    func(t *testing.T, articles []Article) {},
			description: "Should skip items without link",
		},
		{
			name: "mixed valid and invalid items",
			items: []*gofeed.Item{
				{
					Title:           "Valid Article",
					Link:            "https://example.com/valid",
					PublishedParsed: &publishedTime,
				},
				{
					Title:       "No Link",
					Description: "Description",
				},
				{
					Link:        "https://example.com/no-title",
					Description: "Description",
				},
				{
					Title:           "Another Valid",
					Link:            "https://example.com/valid2",
					PublishedParsed: &publishedTime,
				},
			},
			wantCount: 2,
			validate: func(t *testing.T, articles []Article) {
				assert.Equal(t, "Valid Article", articles[0].Title)
				assert.Equal(t, "Another Valid", articles[1].Title)
			},
			description: "Should only transform valid items",
		},
		{
			name:        "empty items list",
			items:       []*gofeed.Item{},
			wantCount:   0,
			validate:    func(t *testing.T, articles []Article) {},
			description: "Should handle empty list",
		},
		{
			name:        "nil items list",
			items:       nil,
			wantCount:   0,
			validate:    func(t *testing.T, articles []Article) {},
			description: "Should handle nil list",
		},
		{
			name: "item with empty description and content",
			items: []*gofeed.Item{
				{
					Title:           "Article",
					Link:            "https://example.com/article",
					Description:     "",
					Content:         "",
					PublishedParsed: &publishedTime,
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, articles []Article) {
				assert.Nil(t, articles[0].Content)
			},
			description: "Should set content to nil when both description and content are empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles := transformFeedItems(tt.items, now)

			assert.Len(t, articles, tt.wantCount, tt.description)
			if tt.wantCount > 0 {
				tt.validate(t, articles)
			}
		})
	}
}

// BenchmarkTransformFeedItems benchmarks feed item transformation
func BenchmarkTransformFeedItems(b *testing.B) {
	now := time.Now()
	publishedTime := now.Add(-24 * time.Hour)

	items := make([]*gofeed.Item, 100)
	for i := 0; i < 100; i++ {
		items[i] = &gofeed.Item{
			Title:           "Article Title",
			Link:            "https://example.com/article",
			Description:     "Article description",
			PublishedParsed: &publishedTime,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transformFeedItems(items, now)
	}
}
