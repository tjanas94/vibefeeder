package errors

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	sharedview "github.com/tjanas94/vibefeeder/internal/shared/view"
)

// NewHTTPErrorHandler creates a custom error handler for Echo
func NewHTTPErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// Log the error
		logger.Error("Request error",
			"error", err,
			"method", c.Request().Method,
			"path", c.Request().URL.Path,
			"remote_ip", c.RealIP(),
		)

		// If response has already been committed, we can't send a new response
		if c.Response().Committed {
			return
		}

		// Default to 500 Internal Server Error
		code := http.StatusInternalServerError
		title := "Internal Server Error"
		message := "An unexpected error occurred. Please try again later."

		// Check if it's an Echo HTTPError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code

			// Set title based on status code
			switch code {
			case http.StatusNotFound:
				title = "Not Found"
				message = "The page you're looking for doesn't exist."
			case http.StatusMethodNotAllowed:
				title = "Method Not Allowed"
				message = "The requested method is not allowed for this resource."
			case http.StatusBadRequest:
				title = "Bad Request"
				message = "The request could not be understood or was missing required parameters."
			case http.StatusUnauthorized:
				title = "Unauthorized"
				message = "You need to be authenticated to access this resource."
			case http.StatusForbidden:
				title = "Forbidden"
				message = "You don't have permission to access this resource."
			case http.StatusTooManyRequests:
				title = "Too Many Requests"
				message = "You've made too many requests. Please try again later."
			case http.StatusServiceUnavailable:
				title = "Service Unavailable"
				message = "The service is temporarily unavailable. Please try again later."
			default:
				title = "Error"
				// Try to use custom message if available and it's a string
				if msg, ok := he.Message.(string); ok && msg != "" {
					message = msg
				}
			}
		}

		// Check if this is an HTMX request
		isHTMX := c.Request().Header.Get("HX-Request") == "true"

		var component templ.Component
		if isHTMX {
			// For HTMX requests, use error fragment with hx-swap-oob
			component = sharedview.ErrorFragment(sharedview.ErrorFragmentProps{
				Message: message,
			})
		} else {
			// For regular requests, use the full error page
			component = sharedview.ErrorPage(sharedview.ErrorPageProps{
				Code:    code,
				Title:   title,
				Message: message,
			})
		}

		// Render the component
		if err := c.Render(code, "", component); err != nil {
			logger.Error("Failed to render error response", "error", err)
		}
	}
}
