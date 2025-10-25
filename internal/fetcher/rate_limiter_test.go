package fetcher

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(delay)

	assert.NotNil(t, rl)
	assert.NotNil(t, rl.domainLastRequest)
	assert.Equal(t, delay, rl.domainDelay)
	assert.Empty(t, rl.domainLastRequest)
}

func TestRateLimiter_WaitIfNeeded_FirstRequest(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	start := time.Now()
	err := rl.WaitIfNeeded(ctx, "https://example.com/feed.xml")
	elapsed := time.Since(start)

	require.NoError(t, err)
	// First request should not wait
	assert.Less(t, elapsed, 10*time.Millisecond)
	// Domain should be tracked
	assert.Contains(t, rl.domainLastRequest, "example.com")
}

func TestRateLimiter_WaitIfNeeded_RateLimiting(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(delay)
	ctx := context.Background()
	feedURL := "https://example.com/feed.xml"

	// First request
	err := rl.WaitIfNeeded(ctx, feedURL)
	require.NoError(t, err)

	// Second request immediately after - should wait
	start := time.Now()
	err = rl.WaitIfNeeded(ctx, feedURL)
	elapsed := time.Since(start)

	require.NoError(t, err)
	// Should wait approximately the delay time (with some tolerance)
	assert.GreaterOrEqual(t, elapsed, delay-10*time.Millisecond)
	assert.LessOrEqual(t, elapsed, delay+50*time.Millisecond)
}

func TestRateLimiter_WaitIfNeeded_DifferentDomains(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	// Request to first domain
	err := rl.WaitIfNeeded(ctx, "https://example.com/feed.xml")
	require.NoError(t, err)

	// Request to different domain - should not wait
	start := time.Now()
	err = rl.WaitIfNeeded(ctx, "https://other.com/feed.xml")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 10*time.Millisecond)
	// Both domains should be tracked
	assert.Contains(t, rl.domainLastRequest, "example.com")
	assert.Contains(t, rl.domainLastRequest, "other.com")
}

func TestRateLimiter_WaitIfNeeded_SameDomainDifferentPaths(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(delay)
	ctx := context.Background()

	// First request
	err := rl.WaitIfNeeded(ctx, "https://example.com/feed1.xml")
	require.NoError(t, err)

	// Second request to same domain, different path - should wait
	start := time.Now()
	err = rl.WaitIfNeeded(ctx, "https://example.com/feed2.xml")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, delay-10*time.Millisecond)
}

func TestRateLimiter_WaitIfNeeded_AfterDelayElapsed(t *testing.T) {
	delay := 50 * time.Millisecond
	rl := NewRateLimiter(delay)
	ctx := context.Background()
	feedURL := "https://example.com/feed.xml"

	// First request
	err := rl.WaitIfNeeded(ctx, feedURL)
	require.NoError(t, err)

	// Wait for delay to pass
	time.Sleep(delay + 10*time.Millisecond)

	// Second request after delay - should not wait
	start := time.Now()
	err = rl.WaitIfNeeded(ctx, feedURL)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 10*time.Millisecond)
}

func TestRateLimiter_WaitIfNeeded_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter(200 * time.Millisecond)
	feedURL := "https://example.com/feed.xml"

	// First request to set up rate limiting
	err := rl.WaitIfNeeded(context.Background(), feedURL)
	require.NoError(t, err)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Second request should be cancelled while waiting
	start := time.Now()
	err = rl.WaitIfNeeded(ctx, feedURL)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	// Should cancel before full delay elapsed
	assert.Less(t, elapsed, 150*time.Millisecond)
}

func TestRateLimiter_WaitIfNeeded_ContextTimeout(t *testing.T) {
	rl := NewRateLimiter(200 * time.Millisecond)
	feedURL := "https://example.com/feed.xml"

	// First request
	err := rl.WaitIfNeeded(context.Background(), feedURL)
	require.NoError(t, err)

	// Create context with timeout shorter than delay
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Second request should timeout
	err = rl.WaitIfNeeded(ctx, feedURL)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRateLimiter_WaitIfNeeded_InvalidURL(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	// Invalid URL should return error
	err := rl.WaitIfNeeded(ctx, "://invalid-url")

	require.Error(t, err)
}

func TestRateLimiter_WaitIfNeeded_ConcurrentSameDomain(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(delay)
	feedURL := "https://example.com/feed.xml"
	goroutines := 5

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	start := time.Now()

	// Launch concurrent requests to same domain
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rl.WaitIfNeeded(context.Background(), feedURL)
			errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	totalElapsed := time.Since(start)

	// Check all requests succeeded
	for err := range errors {
		require.NoError(t, err)
	}

	// With 5 goroutines and 100ms delay, we expect at least 4 delays to occur
	// (first request is immediate, then 4 more requests each wait 100ms)
	// We use generous tolerance because goroutine scheduling can vary
	minExpected := delay * time.Duration(goroutines-1)
	maxExpected := delay * time.Duration(goroutines+1)

	assert.GreaterOrEqual(t, totalElapsed, minExpected-50*time.Millisecond,
		"Expected at least %v but got %v", minExpected, totalElapsed)
	assert.LessOrEqual(t, totalElapsed, maxExpected,
		"Expected at most %v but got %v", maxExpected, totalElapsed)
}

func TestRateLimiter_WaitIfNeeded_ConcurrentDifferentDomains(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(delay)
	goroutines := 5

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	start := time.Now()

	// Launch concurrent requests to different domains
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			feedURL := "https://example" + string(rune('a'+idx)) + ".com/feed.xml"
			err := rl.WaitIfNeeded(context.Background(), feedURL)
			errors <- err
		}(i)
	}

	wg.Wait()
	close(errors)

	totalElapsed := time.Since(start)

	// Check all requests succeeded
	for err := range errors {
		require.NoError(t, err)
	}

	// Different domains should not wait for each other
	// Total time should be much less than delay * goroutines
	assert.Less(t, totalElapsed, delay*time.Duration(goroutines)/2)
}

func TestRateLimiter_WaitIfNeeded_URLVariants(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	tests := []struct {
		name       string
		url1       string
		url2       string
		sameDomain bool
	}{
		{
			name:       "http vs https same domain",
			url1:       "http://example.com/feed.xml",
			url2:       "https://example.com/feed.xml",
			sameDomain: true,
		},
		{
			name:       "different subdomains",
			url1:       "https://www.example.com/feed.xml",
			url2:       "https://api.example.com/feed.xml",
			sameDomain: false,
		},
		{
			name:       "with and without port",
			url1:       "https://example.com/feed.xml",
			url2:       "https://example.com:8080/feed.xml",
			sameDomain: true,
		},
		{
			name:       "different ports",
			url1:       "https://example.com:8080/feed.xml",
			url2:       "https://example.com:9090/feed.xml",
			sameDomain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset rate limiter for each test
			rl = NewRateLimiter(100 * time.Millisecond)

			// First request
			err := rl.WaitIfNeeded(ctx, tt.url1)
			require.NoError(t, err)

			// Second request
			start := time.Now()
			err = rl.WaitIfNeeded(ctx, tt.url2)
			elapsed := time.Since(start)

			require.NoError(t, err)
			if tt.sameDomain {
				// Should wait
				assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond)
			} else {
				// Should not wait
				assert.Less(t, elapsed, 10*time.Millisecond)
			}
		})
	}
}

func TestRateLimiter_CleanOldEntries(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	// Add some entries
	err := rl.WaitIfNeeded(ctx, "https://example1.com/feed.xml")
	require.NoError(t, err)

	err = rl.WaitIfNeeded(ctx, "https://example2.com/feed.xml")
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Add another entry
	err = rl.WaitIfNeeded(ctx, "https://example3.com/feed.xml")
	require.NoError(t, err)

	// Verify all entries exist
	assert.Len(t, rl.domainLastRequest, 3)

	// Clean entries older than 40ms (should remove first two)
	cutoff := time.Now().Add(-40 * time.Millisecond)
	rl.CleanOldEntries(cutoff)

	// Only the newest entry should remain
	assert.Len(t, rl.domainLastRequest, 1)
	assert.Contains(t, rl.domainLastRequest, "example3.com")
	assert.NotContains(t, rl.domainLastRequest, "example1.com")
	assert.NotContains(t, rl.domainLastRequest, "example2.com")
}

func TestRateLimiter_CleanOldEntries_NoEntries(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)

	// Clean on empty map should not panic
	assert.NotPanics(t, func() {
		rl.CleanOldEntries(time.Now())
	})

	assert.Empty(t, rl.domainLastRequest)
}

func TestRateLimiter_CleanOldEntries_AllOld(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	// Add entries
	err := rl.WaitIfNeeded(ctx, "https://example1.com/feed.xml")
	require.NoError(t, err)

	err = rl.WaitIfNeeded(ctx, "https://example2.com/feed.xml")
	require.NoError(t, err)

	// Wait for all entries to become old
	time.Sleep(50 * time.Millisecond)

	// Clean with cutoff in the future (all entries should be removed)
	cutoff := time.Now().Add(100 * time.Millisecond)
	rl.CleanOldEntries(cutoff)

	assert.Empty(t, rl.domainLastRequest)
}

func TestRateLimiter_CleanOldEntries_NoneOld(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	ctx := context.Background()

	// Add entries
	err := rl.WaitIfNeeded(ctx, "https://example1.com/feed.xml")
	require.NoError(t, err)

	err = rl.WaitIfNeeded(ctx, "https://example2.com/feed.xml")
	require.NoError(t, err)

	// Clean with cutoff in the past (no entries should be removed)
	cutoff := time.Now().Add(-1 * time.Hour)
	rl.CleanOldEntries(cutoff)

	assert.Len(t, rl.domainLastRequest, 2)
}

func TestRateLimiter_CleanOldEntries_Concurrent(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)
	ctx := context.Background()

	var wg sync.WaitGroup

	// Concurrent adds and cleans
	for i := 0; i < 10; i++ {
		wg.Add(2)

		// Add entry
		go func(idx int) {
			defer wg.Done()
			feedURL := "https://example" + string(rune('a'+idx)) + ".com/feed.xml"
			_ = rl.WaitIfNeeded(ctx, feedURL)
		}(i)

		// Clean old entries
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			cutoff := time.Now().Add(-25 * time.Millisecond)
			rl.CleanOldEntries(cutoff)
		}()
	}

	wg.Wait()

	// Should not panic and should have some entries
	assert.NotPanics(t, func() {
		rl.mu.Lock()
		_ = len(rl.domainLastRequest)
		rl.mu.Unlock()
	})
}
