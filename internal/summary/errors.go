package summary

import (
	"net/http"

	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
)

// NewNoArticlesFoundError creates a ServiceError when no articles are found for summarization
// Returns 404 Not Found
func NewNoArticlesFoundError() *sharederrors.ServiceError {
	return sharederrors.NewServiceError(
		http.StatusNotFound,
		"No articles found in the last 24 hours",
	)
}

// NewAIServiceUnavailableError creates a ServiceError when the AI service fails
// Returns 503 Service Unavailable
func NewAIServiceUnavailableError() *sharederrors.ServiceError {
	return sharederrors.NewServiceError(
		http.StatusServiceUnavailable,
		"AI service is currently unavailable",
	)
}

// NewDatabaseError creates a ServiceError for database operation failures
// Returns 500 Internal Server Error
func NewDatabaseError(err error) *sharederrors.ServiceError {
	return sharederrors.NewServiceErrorWithCause(
		http.StatusInternalServerError,
		"Database operation failed",
		err,
	)
}
