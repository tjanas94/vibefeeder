package summary

import "errors"

// Business logic errors - map to specific HTTP status codes
var (
	// ErrNoFeeds indicates the user has no RSS feeds configured (400 Bad Request)
	ErrNoFeeds = errors.New("you must add at least one RSS feed before generating a summary")

	// ErrNoArticlesFound indicates no articles were found in the last 24 hours (404 Not Found)
	ErrNoArticlesFound = errors.New("no articles found from the last 24 hours")

	// ErrAIServiceUnavailable indicates the AI service is not responding (503 Service Unavailable)
	ErrAIServiceUnavailable = errors.New("AI service is temporarily unavailable")

	// ErrDatabase indicates a database operation failed (500 Internal Server Error)
	ErrDatabase = errors.New("failed to generate summary. Please try again later")
)
