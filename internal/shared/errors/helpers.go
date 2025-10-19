package errors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents an error response with title and message
type ErrorResponse struct {
	Code    int
	Title   string
	Message string
}

// getErrorResponse maps HTTP status code and error to an ErrorResponse.
// This is a pure function that centralizes error message mapping logic.
func getErrorResponse(code int, err error) ErrorResponse {
	// Check if it's an Echo HTTPError
	he, ok := err.(*echo.HTTPError)
	if !ok {
		// Not an HTTPError - return default 500 response
		return ErrorResponse{
			Code:    code,
			Title:   "Internal Server Error",
			Message: "An unexpected error occurred. Please try again later.",
		}
	}

	// Map status code to response
	switch he.Code {
	case http.StatusNotFound:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Not Found",
			Message: "The page you're looking for doesn't exist.",
		}
	case http.StatusMethodNotAllowed:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Method Not Allowed",
			Message: "The requested method is not allowed for this resource.",
		}
	case http.StatusBadRequest:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Bad Request",
			Message: "The request could not be understood or was missing required parameters.",
		}
	case http.StatusUnauthorized:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Unauthorized",
			Message: "You need to be authenticated to access this resource.",
		}
	case http.StatusForbidden:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Forbidden",
			Message: "You don't have permission to access this resource.",
		}
	case http.StatusTooManyRequests:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Too Many Requests",
			Message: "You've made too many requests. Please try again later.",
		}
	case http.StatusServiceUnavailable:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Service Unavailable",
			Message: "The service is temporarily unavailable. Please try again later.",
		}
	case http.StatusInternalServerError:
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Internal Server Error",
			Message: "An unexpected error occurred. Please try again later.",
		}
	default:
		// For other error codes, use generic "Error" title with custom message if available
		message := "An error occurred. Please try again later."
		if msg, ok := he.Message.(string); ok && msg != "" {
			message = msg
		}
		return ErrorResponse{
			Code:    he.Code,
			Title:   "Error",
			Message: message,
		}
	}
}
