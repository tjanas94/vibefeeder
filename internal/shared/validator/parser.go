package validator

import (
	"github.com/labstack/echo/v4"
)

// ParseFieldErrors extracts field-specific errors from validation error.
// Returns a map of field names to their error messages.
// Expected format from validator: map[string]string{"FieldName": "error message"}
func ParseFieldErrors(err error) map[string]string {
	// Check if it's an Echo HTTP error (from validator)
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		return nil
	}

	// Check if message is already a map (from our validator)
	if errorMap, ok := httpErr.Message.(map[string]string); ok {
		return errorMap
	}

	return nil
}
