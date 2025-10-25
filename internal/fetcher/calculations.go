package fetcher

import (
	"strconv"
	"strings"
	"time"
)

// calculateNextFetch calculates the next fetch time based on Cache-Control header
func calculateNextFetch(cacheControl string, successInterval time.Duration, now time.Time) time.Time {
	serverInterval, found := parseCacheControlMaxAge(cacheControl)

	// Use the larger of server interval or our minimum interval
	if found && serverInterval > successInterval {
		return now.Add(serverInterval)
	}

	return now.Add(successInterval)
}

// parseCacheControlMaxAge extracts max-age directive from Cache-Control header
func parseCacheControlMaxAge(cacheControl string) (time.Duration, bool) {
	if cacheControl == "" {
		return 0, false
	}

	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if !strings.HasPrefix(part, "max-age=") {
			continue
		}

		maxAgeStr := strings.TrimPrefix(part, "max-age=")
		maxAge, err := strconv.Atoi(maxAgeStr)
		if err != nil {
			continue
		}

		return time.Duration(maxAge) * time.Second, true
	}

	return 0, false
}

// calculateBackoff calculates exponential backoff for retry attempts
func calculateBackoff(retryCount int, now time.Time) time.Time {
	// Exponential backoff: 15min, 30min, 60min, 120min, 240min, max 6 hours
	backoffMinutes := 15 * (1 << min(retryCount, 5))
	if backoffMinutes > 360 { // 6 hours
		backoffMinutes = 360
	}

	return now.Add(time.Duration(backoffMinutes) * time.Minute)
}

// parseRetryAfter parses the Retry-After header and returns the next fetch time
// Returns zero time if parsing fails or header is empty
func parseRetryAfter(retryAfter string, retryCount int, now time.Time) time.Time {
	if retryAfter == "" {
		return calculateBackoff(retryCount, now)
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return now.Add(time.Duration(seconds) * time.Second)
	}

	// Try parsing as HTTP date
	if parsedTime, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return parsedTime
	}

	// Fallback to exponential backoff
	return calculateBackoff(retryCount, now)
}
