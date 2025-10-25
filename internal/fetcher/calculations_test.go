package fetcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCalculateNextFetch tests the next fetch time calculation logic
func TestCalculateNextFetch(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	successInterval := 15 * time.Minute

	tests := []struct {
		name         string
		cacheControl string
		wantDuration time.Duration
		description  string
	}{
		{
			name:         "empty cache control uses success interval",
			cacheControl: "",
			wantDuration: 15 * time.Minute,
			description:  "Should use default success interval when no cache control",
		},
		{
			name:         "server interval smaller than success interval",
			cacheControl: "max-age=600", // 10 minutes
			wantDuration: 15 * time.Minute,
			description:  "Should use success interval when server suggests shorter time",
		},
		{
			name:         "server interval larger than success interval",
			cacheControl: "max-age=3600", // 60 minutes
			wantDuration: 60 * time.Minute,
			description:  "Should respect server's longer max-age",
		},
		{
			name:         "cache control with multiple directives",
			cacheControl: "public, max-age=7200, must-revalidate",
			wantDuration: 120 * time.Minute,
			description:  "Should extract max-age from multiple directives",
		},
		{
			name:         "cache control with no-cache",
			cacheControl: "no-cache, no-store",
			wantDuration: 15 * time.Minute,
			description:  "Should use default when no max-age present",
		},
		{
			name:         "invalid max-age value",
			cacheControl: "max-age=invalid",
			wantDuration: 15 * time.Minute,
			description:  "Should fallback to default on parse error",
		},
		{
			name:         "max-age with spaces",
			cacheControl: "max-age = 1800",
			wantDuration: 15 * time.Minute,
			description:  "Should handle malformed max-age (spaces around =)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNextFetch(tt.cacheControl, successInterval, now)
			expected := now.Add(tt.wantDuration)

			assert.Equal(t, expected, result, tt.description)
		})
	}
}

// TestParseCacheControlMaxAge tests Cache-Control header parsing
func TestParseCacheControlMaxAge(t *testing.T) {
	tests := []struct {
		name         string
		cacheControl string
		wantDuration time.Duration
		wantFound    bool
	}{
		{
			name:         "valid max-age",
			cacheControl: "max-age=3600",
			wantDuration: 3600 * time.Second,
			wantFound:    true,
		},
		{
			name:         "max-age with other directives before",
			cacheControl: "public, max-age=1800",
			wantDuration: 1800 * time.Second,
			wantFound:    true,
		},
		{
			name:         "max-age with other directives after",
			cacheControl: "max-age=7200, must-revalidate",
			wantDuration: 7200 * time.Second,
			wantFound:    true,
		},
		{
			name:         "max-age with spaces in directives",
			cacheControl: "public , max-age=600 , must-revalidate",
			wantDuration: 600 * time.Second,
			wantFound:    true,
		},
		{
			name:         "empty cache control",
			cacheControl: "",
			wantDuration: 0,
			wantFound:    false,
		},
		{
			name:         "no max-age directive",
			cacheControl: "no-cache, no-store, must-revalidate",
			wantDuration: 0,
			wantFound:    false,
		},
		{
			name:         "invalid max-age value",
			cacheControl: "max-age=not-a-number",
			wantDuration: 0,
			wantFound:    false,
		},
		{
			name:         "negative max-age",
			cacheControl: "max-age=-100",
			wantDuration: -100 * time.Second,
			wantFound:    true,
		},
		{
			name:         "max-age zero",
			cacheControl: "max-age=0",
			wantDuration: 0,
			wantFound:    true,
		},
		{
			name:         "very large max-age",
			cacheControl: "max-age=31536000", // 1 year in seconds
			wantDuration: 31536000 * time.Second,
			wantFound:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, found := parseCacheControlMaxAge(tt.cacheControl)

			assert.Equal(t, tt.wantFound, found, "found flag mismatch")
			if tt.wantFound {
				assert.Equal(t, tt.wantDuration, duration, "duration mismatch")
			}
		})
	}
}

// TestCalculateBackoff tests exponential backoff calculation
func TestCalculateBackoff(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		retryCount   int
		wantDuration time.Duration
		description  string
	}{
		{
			name:         "first retry",
			retryCount:   0,
			wantDuration: 15 * time.Minute,
			description:  "First retry should wait 15 minutes",
		},
		{
			name:         "second retry",
			retryCount:   1,
			wantDuration: 30 * time.Minute,
			description:  "Second retry should wait 30 minutes",
		},
		{
			name:         "third retry",
			retryCount:   2,
			wantDuration: 60 * time.Minute,
			description:  "Third retry should wait 60 minutes",
		},
		{
			name:         "fourth retry",
			retryCount:   3,
			wantDuration: 120 * time.Minute,
			description:  "Fourth retry should wait 120 minutes",
		},
		{
			name:         "fifth retry",
			retryCount:   4,
			wantDuration: 240 * time.Minute,
			description:  "Fifth retry should wait 240 minutes",
		},
		{
			name:         "sixth retry caps at 6 hours",
			retryCount:   5,
			wantDuration: 360 * time.Minute,
			description:  "Sixth retry should cap at 360 minutes",
		},
		{
			name:         "many retries still cap at 6 hours",
			retryCount:   10,
			wantDuration: 360 * time.Minute,
			description:  "High retry counts should still cap at 360 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBackoff(tt.retryCount, now)
			expected := now.Add(tt.wantDuration)

			assert.Equal(t, expected, result, tt.description)
		})
	}
}

// TestParseRetryAfter tests Retry-After header parsing
func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	httpDate := "Mon, 15 Jan 2024 13:30:00 GMT"
	httpDateTime, _ := time.Parse(time.RFC1123, httpDate)

	tests := []struct {
		name        string
		retryAfter  string
		retryCount  int
		wantTime    time.Time
		description string
	}{
		{
			name:        "retry after in seconds",
			retryAfter:  "120",
			retryCount:  0,
			wantTime:    now.Add(120 * time.Second),
			description: "Should parse seconds correctly",
		},
		{
			name:        "retry after as http date",
			retryAfter:  httpDate,
			retryCount:  0,
			wantTime:    httpDateTime,
			description: "Should parse HTTP date format",
		},
		{
			name:        "empty retry after uses backoff",
			retryAfter:  "",
			retryCount:  2,
			wantTime:    now.Add(60 * time.Minute), // retryCount 2 = 60 min backoff
			description: "Should fallback to exponential backoff when empty",
		},
		{
			name:        "invalid retry after uses backoff",
			retryAfter:  "invalid-value",
			retryCount:  1,
			wantTime:    now.Add(30 * time.Minute), // retryCount 1 = 30 min backoff
			description: "Should fallback to exponential backoff when invalid",
		},
		{
			name:        "zero seconds",
			retryAfter:  "0",
			retryCount:  0,
			wantTime:    now,
			description: "Should handle zero seconds",
		},
		{
			name:        "large seconds value",
			retryAfter:  "86400", // 1 day
			retryCount:  0,
			wantTime:    now.Add(86400 * time.Second),
			description: "Should handle large seconds value",
		},
		{
			name:        "negative seconds uses that value",
			retryAfter:  "-60",
			retryCount:  3,
			wantTime:    now.Add(-60 * time.Second),
			description: "Should parse negative seconds (time in the past)",
		},
		{
			name:        "malformed http date uses backoff",
			retryAfter:  "Mon, 15 Jan 2024 99:99:99 GMT",
			retryCount:  0,
			wantTime:    now.Add(15 * time.Minute),
			description: "Should fallback to backoff for malformed dates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.retryAfter, tt.retryCount, now)

			assert.Equal(t, tt.wantTime, result, tt.description)
		})
	}
}

// BenchmarkParseCacheControlMaxAge benchmarks cache control parsing
func BenchmarkParseCacheControlMaxAge(b *testing.B) {
	cacheControl := "public, max-age=3600, must-revalidate"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseCacheControlMaxAge(cacheControl)
	}
}

// BenchmarkCalculateBackoff benchmarks backoff calculation
func BenchmarkCalculateBackoff(b *testing.B) {
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateBackoff(3, now)
	}
}

// BenchmarkParseRetryAfter benchmarks retry after parsing
func BenchmarkParseRetryAfter(b *testing.B) {
	now := time.Now()

	b.Run("seconds", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parseRetryAfter("120", 0, now)
		}
	})

	b.Run("http_date", func(b *testing.B) {
		httpDate := "Mon, 15 Jan 2024 13:30:00 GMT"
		for i := 0; i < b.N; i++ {
			_ = parseRetryAfter(httpDate, 0, now)
		}
	})

	b.Run("fallback", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parseRetryAfter("", 2, now)
		}
	})
}
