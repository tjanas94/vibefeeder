package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/dashboard/models"
	"github.com/tjanas94/vibefeeder/internal/dashboard/view"
	sharedView "github.com/tjanas94/vibefeeder/internal/shared/view"
)

// Handler handles dashboard requests
type Handler struct {
	logger *slog.Logger
}

// NewHandler creates a new dashboard handler
func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// ShowDashboard renders the main dashboard page
// Returns empty layout with htmx-enabled containers that load content dynamically
func (h *Handler) ShowDashboard(c echo.Context) error {
	// Create empty view model - all data loaded via htmx
	vm := models.DashboardViewModel{}

	// Render dashboard template
	if err := c.Render(http.StatusOK, "", view.Index(vm)); err != nil {
		h.logger.Error("failed to render dashboard",
			"error", err,
			"path", c.Request().URL.Path,
		)
		return c.Render(
			http.StatusInternalServerError,
			"",
			sharedView.ErrorPage(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to load dashboard. Please refresh the page.",
			),
		)
	}

	return nil
}
