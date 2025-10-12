package database

import "strings"

// IsUniqueViolationError checks if the error is due to unique constraint violation.
// PostgREST returns 409 status code wrapped in error message for unique violations.
// This helper can be used across all features to detect duplicate key errors.
func IsUniqueViolationError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// PostgREST returns errors with "409" or "duplicate key" in the message
	return strings.Contains(errMsg, "409") ||
		strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint")
}

// IsNotFoundError checks if the error indicates that a resource was not found.
// PostgREST returns 404 or "no rows" in error message when resource doesn't exist.
// This helper can be used across all features to detect missing resources.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "no rows") ||
		strings.Contains(errMsg, "404")
}
