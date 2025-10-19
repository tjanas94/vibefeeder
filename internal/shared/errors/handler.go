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

		// Get error response using helper function
		resp := getErrorResponse(http.StatusInternalServerError, err)

		// Check if this is an HTMX request
		isHTMX := c.Request().Header.Get("HX-Request") == "true"

		var component templ.Component
		if isHTMX {
			// For HTMX requests, use error fragment with hx-swap-oob
			// Set HX-Reswap: none to prevent clearing the main target (e.g., modal content)
			c.Response().Header().Set("HX-Reswap", "none")
			component = sharedview.ErrorFragment(sharedview.ErrorFragmentProps{
				Message: resp.Message,
			})
		} else {
			// For regular requests, use the full error page
			component = sharedview.ErrorPage(sharedview.ErrorPageProps{
				Code:    resp.Code,
				Title:   resp.Title,
				Message: resp.Message,
			})
		}

		// Render the component
		if err := c.Render(resp.Code, "", component); err != nil {
			logger.Error("Failed to render error response", "error", err)
		}
	}
}
