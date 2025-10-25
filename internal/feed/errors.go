package feed

import (
	"net/http"

	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
)

// NewFeedAlreadyExistsError creates a ServiceError when attempting to add a duplicate feed URL
// Returns 409 Conflict with field error
func NewFeedAlreadyExistsError() *sharederrors.ServiceError {
	return sharederrors.NewServiceErrorWithFields(
		http.StatusConflict,
		"",
		map[string]string{
			"URL": "You have already added this feed",
		},
	)
}

// NewFeedNotFoundError creates a ServiceError when a feed is not found or doesn't belong to the user
// Returns 404 Not Found
func NewFeedNotFoundError() *sharederrors.ServiceError {
	return sharederrors.NewServiceError(
		http.StatusNotFound,
		"Feed not found",
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
