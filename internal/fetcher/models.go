package fetcher

import "time"

// Article represents a parsed feed article
type Article struct {
	Title       string
	URL         string
	Content     *string
	PublishedAt time.Time
}

// FetchDecision represents the decision to make after handling an HTTP response
type FetchDecision struct {
	ShouldRetry   bool
	NextFetchTime time.Time
	Status        string
	ErrorMessage  *string
	ETag          *string
	LastModified  *string
	NewURL        *string
	Articles      []Article
}
