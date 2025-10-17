package fetcher

import (
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
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
	// Exponential backoff: 5min, 10min, 20min, 40min, 80min, max 2 hours
	backoffMinutes := 5 * (1 << min(retryCount, 5))
	if backoffMinutes > 120 {
		backoffMinutes = 120
	}

	return now.Add(time.Duration(backoffMinutes) * time.Minute)
}

// transformFeedItems transforms gofeed items to our Article format
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
