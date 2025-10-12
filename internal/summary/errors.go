package summary

import "errors"

var (
	// ErrNoArticlesFound indicates that no articles were found in the last 24 hours
	ErrNoArticlesFound = errors.New("no articles found in last 24 hours")

	// ErrAIServiceUnavailable indicates that the AI service is not responding or returned an error
	ErrAIServiceUnavailable = errors.New("AI service is currently unavailable")

	// ErrDatabase indicates a database operation failure
	ErrDatabase = errors.New("database operation failed")
)
