package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/dashboard/models"
	"github.com/tjanas94/vibefeeder/internal/dashboard/view"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	sharedview "github.com/tjanas94/vibefeeder/internal/shared/view"
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
	// Get user ID from context (set by auth middleware)
	userID := auth.GetUserID(c)

	// TODO: Fetch actual user data from database when user service is implemented
	// For now, use mock email based on user ID
	mockEmail := "user@example.com"

	// Prepare view model
	vm := models.DashboardViewModel{
		Title:     "Dashboard - VibeFeeder",
		UserEmail: mockEmail,
	}

	// Render dashboard template
	if err := c.Render(http.StatusOK, "", view.Index(vm)); err != nil {
		h.logger.Error("failed to render dashboard",
			"error", err,
			"path", c.Request().URL.Path,
			"user_id", userID,
		)
		return c.Render(
			http.StatusInternalServerError,
			"",
			sharedview.ErrorPage(sharedview.ErrorPageProps{
				Code:    http.StatusInternalServerError,
				Title:   "Internal Server Error",
				Message: "Failed to load dashboard. Please refresh the page.",
			}),
		)
	}

	return nil
}
