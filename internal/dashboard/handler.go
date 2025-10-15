package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/dashboard/models"
	"github.com/tjanas94/vibefeeder/internal/dashboard/view"
	feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"
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
	// TODO: Fetch actual user data from database when user service is implemented
	// For now, use mock email based on user ID
	mockEmail := "user@example.com"

	// Bind and validate query params using the same struct as GET /feeds
	query := new(feedmodels.ListFeedsQuery)
	if err := c.Bind(query); err != nil {
		h.logger.Warn("failed to bind query parameters", "error", err)
		// Use defaults on binding error
		query = &feedmodels.ListFeedsQuery{}
	}

	query.SetDefaults()

	if err := c.Validate(query); err != nil {
		h.logger.Warn("invalid query parameters", "error", err)
		// Use defaults on validation error
		query = &feedmodels.ListFeedsQuery{}
		query.SetDefaults()
	}

	// Prepare view model
	vm := models.DashboardViewModel{
		Title:     "Dashboard - VibeFeeder",
		UserEmail: mockEmail,
		Query:     query,
	}

	// Render dashboard template
	return c.Render(http.StatusOK, "", view.Index(vm))
}
