package fetcher

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHTTPResponseHandler tests HTTPResponseHandler constructor
func TestNewHTTPResponseHandler(t *testing.T) {
	tests := []struct {
		name            string
		logger          *slog.Logger
		successInterval time.Duration
		expectNilLogger bool
		description     string
	}{
		{
			name:            "creates handler with provided logger",
			logger:          slog.Default(),
			successInterval: 15 * time.Minute,
			expectNilLogger: false,
			description:     "Should create handler with provided logger",
		},
		{
			name:            "creates handler with nil logger",
			logger:          nil,
			successInterval: 15 * time.Minute,
			expectNilLogger: false,
			description:     "Should use slog.Default() when logger is nil",
		},
		{
			name:            "stores success interval",
			logger:          slog.Default(),
			successInterval: 30 * time.Minute,
			expectNilLogger: false,
			description:     "Should store the provided success interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(tt.logger, tt.successInterval)

			assert.NotNil(t, handler, tt.description)
			assert.NotNil(t, handler.logger, "logger should never be nil")
			assert.Equal(t, tt.successInterval, handler.successInterval)
		})
	}
}

// TestHandleResponse routes to correct handler based on status code
func TestHandleResponse(t *testing.T) {
	validFeed := `<?xml version="1.0"?><rss version="2.0"><channel><title>Test</title></channel></rss>`

	tests := []struct {
		name           string
		statusCode     int
		body           string
		headers        http.Header
		expectedStatus string
		description    string
	}{
		{
			name:           "routes 200 to success handler",
			statusCode:     http.StatusOK,
			body:           validFeed,
			headers:        http.Header{},
			expectedStatus: "success",
			description:    "Should route HTTP 200 to success handler",
		},
		{
			name:           "routes 304 to not modified handler",
			statusCode:     http.StatusNotModified,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "success",
			description:    "Should route HTTP 304 to not modified handler",
		},
		{
			name:           "routes 301 to permanent redirect handler",
			statusCode:     http.StatusMovedPermanently,
			body:           "",
			headers:        http.Header{"Location": []string{"https://example.com/feed-v2"}},
			expectedStatus: "redirect",
			description:    "Should route HTTP 301 to permanent redirect handler",
		},
		{
			name:           "routes 308 to permanent redirect handler",
			statusCode:     http.StatusPermanentRedirect,
			body:           "",
			headers:        http.Header{"Location": []string{"https://example.com/feed-v2"}},
			expectedStatus: "redirect",
			description:    "Should route HTTP 308 to permanent redirect handler",
		},
		{
			name:           "routes 302 to temporary redirect handler",
			statusCode:     http.StatusFound,
			body:           "",
			headers:        http.Header{"Location": []string{"https://example.com/feed-v2"}},
			expectedStatus: "redirect",
			description:    "Should route HTTP 302 to temporary redirect handler",
		},
		{
			name:           "routes 303 to temporary redirect handler",
			statusCode:     http.StatusSeeOther,
			body:           "",
			headers:        http.Header{"Location": []string{"https://example.com/feed-v2"}},
			expectedStatus: "redirect",
			description:    "Should route HTTP 303 to temporary redirect handler",
		},
		{
			name:           "routes 307 to temporary redirect handler",
			statusCode:     http.StatusTemporaryRedirect,
			body:           "",
			headers:        http.Header{"Location": []string{"https://example.com/feed-v2"}},
			expectedStatus: "redirect",
			description:    "Should route HTTP 307 to temporary redirect handler",
		},
		{
			name:           "routes 401 to unauthorized handler",
			statusCode:     http.StatusUnauthorized,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "unauthorized",
			description:    "Should route HTTP 401 to unauthorized handler",
		},
		{
			name:           "routes 403 to unauthorized handler",
			statusCode:     http.StatusForbidden,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "unauthorized",
			description:    "Should route HTTP 403 to unauthorized handler",
		},
		{
			name:           "routes 429 to rate limit handler",
			statusCode:     http.StatusTooManyRequests,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "temporary_error",
			description:    "Should route HTTP 429 to rate limit handler",
		},
		{
			name:           "routes 404 to not found handler",
			statusCode:     http.StatusNotFound,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "permanent_error",
			description:    "Should route HTTP 404 to not found handler",
		},
		{
			name:           "routes 410 to not found handler",
			statusCode:     http.StatusGone,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "permanent_error",
			description:    "Should route HTTP 410 to not found handler",
		},
		{
			name:           "routes 400 to client error handler",
			statusCode:     http.StatusBadRequest,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "permanent_error",
			description:    "Should route HTTP 400 to client error handler",
		},
		{
			name:           "routes 500 to server error handler",
			statusCode:     http.StatusInternalServerError,
			body:           "",
			headers:        http.Header{},
			expectedStatus: "temporary_error",
			description:    "Should route HTTP 500 to server error handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     tt.headers,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			assert.Equal(t, tt.expectedStatus, decision.Status, tt.description)
		})
	}
}

// TestHandleSuccess tests successful 200 OK responses
func TestHandleSuccess(t *testing.T) {
	successInterval := 15 * time.Minute

	validFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>Test Description</description>
    <item>
      <title>Article 1</title>
      <link>https://example.com/article1</link>
      <description>Description 1</description>
      <pubDate>Mon, 15 Jan 2024 10:00:00 GMT</pubDate>
    </item>
    <item>
      <title>Article 2</title>
      <link>https://example.com/article2</link>
      <description>Description 2</description>
      <pubDate>Mon, 15 Jan 2024 11:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`

	tests := []struct {
		name               string
		feedContent        string
		headers            http.Header
		maxArticlesPerFeed int
		validate           func(t *testing.T, decision FetchDecision)
		description        string
	}{
		{
			name:               "successfully parses valid RSS feed",
			feedContent:        validFeed,
			headers:            http.Header{},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				assert.Nil(t, decision.ErrorMessage)
				assert.False(t, decision.ShouldRetry)
				assert.Len(t, decision.Articles, 2)
				assert.Equal(t, "Article 1", decision.Articles[0].Title)
				assert.Equal(t, "Article 2", decision.Articles[1].Title)
			},
			description: "Should successfully parse valid RSS feed",
		},
		{
			name:               "extracts ETag header when present",
			feedContent:        validFeed,
			headers:            http.Header{"Etag": []string{"W/\"abc123\""}},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				require.NotNil(t, decision.ETag)
				assert.Equal(t, "W/\"abc123\"", *decision.ETag)
			},
			description: "Should extract ETag header when present",
		},
		{
			name:               "extracts Last-Modified header",
			feedContent:        validFeed,
			headers:            http.Header{"Last-Modified": []string{"Mon, 15 Jan 2024 10:00:00 GMT"}},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				require.NotNil(t, decision.LastModified)
				assert.Equal(t, "Mon, 15 Jan 2024 10:00:00 GMT", *decision.LastModified)
			},
			description: "Should extract and store Last-Modified header",
		},
		{
			name:               "limits number of articles",
			feedContent:        validFeed,
			headers:            http.Header{},
			maxArticlesPerFeed: 1,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				assert.Len(t, decision.Articles, 1)
				assert.Equal(t, "Article 1", decision.Articles[0].Title)
			},
			description: "Should limit articles when exceeding max count",
		},
		{
			name:               "respects cache control max-age",
			feedContent:        validFeed,
			headers:            http.Header{"Cache-Control": []string{"max-age=7200"}},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				// Verify that next fetch time is approximately 120 minutes in the future
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 119*time.Minute)
				assert.Less(t, timeDiff, 121*time.Minute)
			},
			description: "Should respect Cache-Control max-age header",
		},
		{
			name:               "uses success interval when no cache control",
			feedContent:        validFeed,
			headers:            http.Header{},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				// Verify that next fetch time is approximately 15 minutes in the future
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 14*time.Minute)
				assert.Less(t, timeDiff, 16*time.Minute)
			},
			description: "Should use success interval when no cache control",
		},
		{
			name:               "handles invalid feed content",
			feedContent:        "not a valid feed",
			headers:            http.Header{},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "Failed to parse feed")
			},
			description: "Should return error for invalid feed",
		},
		{
			name:               "handles empty feed",
			feedContent:        `<?xml version="1.0"?><rss version="2.0"><channel><title>Empty</title></channel></rss>`,
			headers:            http.Header{},
			maxArticlesPerFeed: 100,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				assert.Len(t, decision.Articles, 0)
			},
			description: "Should handle empty feed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), successInterval)

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     tt.headers,
				Body:       io.NopCloser(strings.NewReader(tt.feedContent)),
			}

			decision := handler.HandleResponse(
				resp,
				"https://example.com/feed",
				10*1024*1024,
				tt.maxArticlesPerFeed,
				0,
				nil,
				nil,
			)

			tt.validate(t, decision)
		})
	}
}

// TestHandleNotModified tests 304 Not Modified responses
func TestHandleNotModified(t *testing.T) {
	successInterval := 15 * time.Minute

	tests := []struct {
		name         string
		cacheControl string
		validate     func(t *testing.T, decision FetchDecision)
		description  string
	}{
		{
			name:         "returns success status",
			cacheControl: "",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				assert.False(t, decision.ShouldRetry)
				assert.Len(t, decision.Articles, 0)
			},
			description: "Should return success with empty articles",
		},
		{
			name:         "uses success interval when no cache control",
			cacheControl: "",
			validate: func(t *testing.T, decision FetchDecision) {
				// Verify that next fetch time is approximately 15 minutes in the future
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 14*time.Minute)
				assert.Less(t, timeDiff, 16*time.Minute)
			},
			description: "Should use success interval when no cache control",
		},
		{
			name:         "respects cache control max-age",
			cacheControl: "max-age=3600",
			validate: func(t *testing.T, decision FetchDecision) {
				// Verify that next fetch time is approximately 60 minutes in the future
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 59*time.Minute)
				assert.Less(t, timeDiff, 61*time.Minute)
			},
			description: "Should respect Cache-Control max-age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), successInterval)

			headers := http.Header{}
			if tt.cacheControl != "" {
				headers.Set("Cache-Control", tt.cacheControl)
			}

			resp := &http.Response{
				StatusCode: http.StatusNotModified,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandlePermanentRedirect tests 301/308 permanent redirects
func TestHandlePermanentRedirect(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		location    string
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "301 redirect with location",
			statusCode: http.StatusMovedPermanently,
			location:   "https://example.com/feed-v2",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "redirect", decision.Status)
				assert.True(t, decision.ShouldRetry)
				require.NotNil(t, decision.NewURL)
				assert.Equal(t, "https://example.com/feed-v2", *decision.NewURL)
			},
			description: "Should return redirect decision with new URL",
		},
		{
			name:       "308 redirect with location",
			statusCode: http.StatusPermanentRedirect,
			location:   "https://example.com/feed-final",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "redirect", decision.Status)
				assert.True(t, decision.ShouldRetry)
				require.NotNil(t, decision.NewURL)
				assert.Equal(t, "https://example.com/feed-final", *decision.NewURL)
			},
			description: "Should handle 308 permanent redirect",
		},
		{
			name:       "missing location header",
			statusCode: http.StatusMovedPermanently,
			location:   "",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "Redirect without Location header")
			},
			description: "Should return error when Location header missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			headers := http.Header{}
			if tt.location != "" {
				headers.Set("Location", tt.location)
			}

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleTemporaryRedirect tests 302/303/307 temporary redirects
func TestHandleTemporaryRedirect(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		location    string
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "302 redirect with location",
			statusCode: http.StatusFound,
			location:   "https://example.com/feed-v2",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "redirect", decision.Status)
				assert.True(t, decision.ShouldRetry)
				require.NotNil(t, decision.NewURL)
				assert.Equal(t, "https://example.com/feed-v2", *decision.NewURL)
			},
			description: "Should return redirect decision for 302",
		},
		{
			name:       "303 redirect with location",
			statusCode: http.StatusSeeOther,
			location:   "https://example.com/feed-v3",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "redirect", decision.Status)
				assert.True(t, decision.ShouldRetry)
			},
			description: "Should handle 303 redirect",
		},
		{
			name:       "307 redirect with location",
			statusCode: http.StatusTemporaryRedirect,
			location:   "https://example.com/feed-v4",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "redirect", decision.Status)
				assert.True(t, decision.ShouldRetry)
			},
			description: "Should handle 307 redirect",
		},
		{
			name:       "missing location header",
			statusCode: http.StatusFound,
			location:   "",
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "Redirect without Location header")
			},
			description: "Should return error when Location header missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			headers := http.Header{}
			if tt.location != "" {
				headers.Set("Location", tt.location)
			}

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleUnauthorized tests 401/403 authorization errors
func TestHandleUnauthorized(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "unauthorized", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "401")
			},
			description: "Should return unauthorized status for 401",
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "unauthorized", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "403")
			},
			description: "Should return unauthorized status for 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleTooManyRequests tests 429 Too Many Requests
func TestHandleTooManyRequests(t *testing.T) {
	tests := []struct {
		name        string
		retryAfter  string
		retryCount  int
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "rate limit with Retry-After seconds",
			retryAfter: "300",
			retryCount: 0,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "temporary_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "429")
				assert.False(t, decision.ShouldRetry)
				// Verify that next fetch time is approximately 300 seconds (5 minutes) in the future
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 299*time.Second)
				assert.Less(t, timeDiff, 301*time.Second)
			},
			description: "Should use Retry-After seconds",
		},
		{
			name:       "rate limit without Retry-After uses backoff",
			retryAfter: "",
			retryCount: 2,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "temporary_error", decision.Status)
				// Verify that next fetch time is approximately 60 minutes in the future (backoff for retryCount=2)
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 59*time.Minute)
				assert.Less(t, timeDiff, 61*time.Minute)
			},
			description: "Should fallback to exponential backoff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			headers := http.Header{}
			if tt.retryAfter != "" {
				headers.Set("Retry-After", tt.retryAfter)
			}

			resp := &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, tt.retryCount, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleNotFound tests 404/410 not found errors
func TestHandleNotFound(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "404")
			},
			description: "Should return permanent error for 404",
		},
		{
			name:       "410 gone",
			statusCode: http.StatusGone,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "410")
			},
			description: "Should return permanent error for 410",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleClientError tests other 4xx client errors
func TestHandleClientError(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "400 bad request",
			statusCode: http.StatusBadRequest,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "400")
			},
			description: "Should return permanent error for 400",
		},
		{
			name:       "418 teapot",
			statusCode: http.StatusTeapot,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "418")
			},
			description: "Should return permanent error for other 4xx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleServerError tests 5xx server errors
func TestHandleServerError(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		retryAfter  string
		retryCount  int
		validate    func(t *testing.T, decision FetchDecision)
		description string
	}{
		{
			name:       "500 internal server error",
			statusCode: http.StatusInternalServerError,
			retryAfter: "",
			retryCount: 0,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "temporary_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
				assert.Contains(t, *decision.ErrorMessage, "500")
				assert.False(t, decision.ShouldRetry)
			},
			description: "Should return temporary error for 500",
		},
		{
			name:       "503 service unavailable with Retry-After",
			statusCode: http.StatusServiceUnavailable,
			retryAfter: "600",
			retryCount: 1,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "temporary_error", decision.Status)
				assert.Contains(t, *decision.ErrorMessage, "503")
				// Verify that next fetch time is approximately 600 seconds (10 minutes) in the future
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 599*time.Second)
				assert.Less(t, timeDiff, 601*time.Second)
			},
			description: "Should respect Retry-After header for 503",
		},
		{
			name:       "502 bad gateway without Retry-After",
			statusCode: http.StatusBadGateway,
			retryAfter: "",
			retryCount: 2,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "temporary_error", decision.Status)
				// Verify that next fetch time is approximately 60 minutes in the future (backoff for retryCount=2)
				timeDiff := time.Until(decision.NextFetchTime)
				assert.Greater(t, timeDiff, 59*time.Minute)
				assert.Less(t, timeDiff, 61*time.Minute)
			},
			description: "Should use backoff without Retry-After",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			headers := http.Header{}
			if tt.retryAfter != "" {
				headers.Set("Retry-After", tt.retryAfter)
			}

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, tt.retryCount, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// TestHandleUnexpectedStatus tests unexpected HTTP status codes
func TestHandleUnexpectedStatus(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		retryAfter  string
		retryCount  int
		description string
	}{
		{
			name:        "199 info status",
			statusCode:  199,
			retryAfter:  "",
			retryCount:  0,
			description: "Should treat unexpected status as temporary error",
		},
		{
			name:        "299 success variant",
			statusCode:  299,
			retryAfter:  "",
			retryCount:  1,
			description: "Should treat unexpected success code as temporary error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			headers := http.Header{}
			if tt.retryAfter != "" {
				headers.Set("Retry-After", tt.retryAfter)
			}

			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, tt.retryCount, nil, nil)

			assert.Equal(t, "temporary_error", decision.Status, tt.description)
			assert.NotNil(t, decision.ErrorMessage)
			assert.Contains(t, *decision.ErrorMessage, fmt.Sprintf("%d", tt.statusCode))
		})
	}
}

// TestValidateURL tests URL validation and resolution
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		newLocation string
		expectError bool
		expectedURL string
		description string
	}{
		{
			name:        "valid absolute URL",
			baseURL:     "https://example.com/feed",
			newLocation: "https://cdn.example.com/feed-v2",
			expectError: false,
			expectedURL: "https://cdn.example.com/feed-v2",
			description: "Should resolve absolute URL",
		},
		{
			name:        "valid relative URL",
			baseURL:     "https://example.com/feed",
			newLocation: "/feed-v2",
			expectError: false,
			expectedURL: "https://example.com/feed-v2",
			description: "Should resolve relative URL",
		},
		{
			name:        "path-only relative URL",
			baseURL:     "https://example.com/feeds/main",
			newLocation: "main-v2",
			expectError: false,
			expectedURL: "https://example.com/feeds/main-v2",
			description: "Should resolve path-relative URL",
		},
		{
			name:        "protocol-relative URL",
			baseURL:     "https://example.com/feed",
			newLocation: "//cdn.example.com/feed",
			expectError: false,
			expectedURL: "https://cdn.example.com/feed",
			description: "Should resolve protocol-relative URL",
		},
		{
			name:        "invalid URL characters",
			baseURL:     "https://example.com/feed",
			newLocation: "ht!tp://invalid",
			expectError: true,
			description: "Should reject invalid URL",
		},
		{
			name:        "empty location resolves to base",
			baseURL:     "https://example.com/feed",
			newLocation: "",
			expectError: false,
			expectedURL: "https://example.com/feed",
			description: "Should resolve empty location to base URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse(tt.baseURL)
			require.NoError(t, err)

			result, err := ValidateURL(baseURL, tt.newLocation)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				require.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedURL, result.String())
			}
		})
	}
}

// TestResponseBodySize tests limiting response body to prevent zip bomb attacks
func TestResponseBodySize(t *testing.T) {
	// Create a feed that exceeds max response body size
	largeContent := strings.Repeat("x", 11*1024*1024) // 11 MB

	tests := []struct {
		name                string
		feedContent         string
		maxResponseBodySize int64
		validate            func(t *testing.T, decision FetchDecision)
		description         string
	}{
		{
			name:                "small feed within limit",
			feedContent:         `<?xml version="1.0"?><rss version="2.0"><channel><title>Test</title></channel></rss>`,
			maxResponseBodySize: 10 * 1024 * 1024,
			validate: func(t *testing.T, decision FetchDecision) {
				assert.Equal(t, "success", decision.Status)
				assert.Nil(t, decision.ErrorMessage)
			},
			description: "Should accept feed within size limit",
		},
		{
			name:                "large feed exceeding limit truncates",
			feedContent:         largeContent,
			maxResponseBodySize: 100, // Very small limit
			validate: func(t *testing.T, decision FetchDecision) {
				// Feed parsing should fail due to truncated content
				assert.Equal(t, "permanent_error", decision.Status)
				assert.NotNil(t, decision.ErrorMessage)
			},
			description: "Should limit response body size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(tt.feedContent)),
			}

			decision := handler.HandleResponse(resp, "https://example.com/feed", tt.maxResponseBodySize, 100, 0, nil, nil)

			tt.validate(t, decision)
		})
	}
}

// BenchmarkHandleSuccess benchmarks successful response handling
func BenchmarkHandleSuccess(b *testing.B) {
	validFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Article 1</title>
      <link>https://example.com/article1</link>
      <description>Description 1</description>
    </item>
  </channel>
</rss>`

	handler := NewHTTPResponseHandler(slog.Default(), 15*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(validFeed)),
		}

		_ = handler.HandleResponse(resp, "https://example.com/feed", 10*1024*1024, 100, 0, nil, nil)
	}
}

// BenchmarkValidateURL benchmarks URL validation
func BenchmarkValidateURL(b *testing.B) {
	baseURL, _ := url.Parse("https://example.com/feed")
	newLocation := "https://cdn.example.com/feed-v2"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateURL(baseURL, newLocation)
	}
}
