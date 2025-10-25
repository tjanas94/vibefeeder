package fetcher

import (
	"context"
	"net/url"
	"sync"
	"time"
)

// RateLimiter ensures minimum delay between requests to same domain
type RateLimiter struct {
	domainLastRequest map[string]time.Time
	mu                sync.Mutex
	domainDelay       time.Duration
}

// NewRateLimiter creates a new rate limiter with specified domain delay
func NewRateLimiter(domainDelay time.Duration) *RateLimiter {
	return &RateLimiter{
		domainLastRequest: make(map[string]time.Time),
		domainDelay:       domainDelay,
	}
}

// WaitIfNeeded blocks until enough time has passed since the last request to this domain
// Returns error if context is cancelled during rate limiting
// Extracts domain from feed URL automatically
func (rl *RateLimiter) WaitIfNeeded(ctx context.Context, feedURL string) error {
	// Parse URL and extract domain
	parsedURL, err := url.Parse(feedURL)
	if err != nil {
		return err
	}

	domain := parsedURL.Hostname()
	return rl.waitForDomain(ctx, domain)
}

// waitForDomain blocks until enough time has passed for a specific domain
func (rl *RateLimiter) waitForDomain(ctx context.Context, domain string) error {
	for {
		rl.mu.Lock()
		lastRequest, exists := rl.domainLastRequest[domain]

		if !exists {
			// No previous request, safe to proceed
			rl.domainLastRequest[domain] = time.Now()
			rl.mu.Unlock()
			return nil
		}

		elapsed := time.Since(lastRequest)
		if elapsed >= rl.domainDelay {
			// Enough time has passed, safe to proceed
			rl.domainLastRequest[domain] = time.Now()
			rl.mu.Unlock()
			return nil
		}

		// Calculate wait time while holding lock
		waitTime := rl.domainDelay - elapsed
		rl.mu.Unlock()

		// Wait with context cancellation support
		select {
		case <-time.After(waitTime):
			// Wait time elapsed, loop back to check again atomically
		case <-ctx.Done():
			// Context cancelled during rate limiting
			return ctx.Err()
		}
	}
}

// CleanOldEntries removes entries older than the specified cutoff time
// This prevents memory leaks from accumulating domain entries
func (rl *RateLimiter) CleanOldEntries(cutoff time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for domain, lastRequest := range rl.domainLastRequest {
		if lastRequest.Before(cutoff) {
			delete(rl.domainLastRequest, domain)
		}
	}
}
