package dashboard

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/dashboard/models"
	"github.com/tjanas94/vibefeeder/internal/dashboard/view"
	feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
)

// Handler handles dashboard requests
type Handler struct{}

// NewHandler creates a new dashboard handler
func NewHandler() *Handler {
	return &Handler{}
}

// ShowDashboard renders the main dashboard page
// Returns empty layout with htmx-enabled containers that load content dynamically
func (h *Handler) ShowDashboard(c echo.Context) error {
	// Get authenticated user email from context
	userEmail := auth.GetUserEmail(c)

	// Bind and sanitize query parameters
	query := new(feedmodels.ListFeedsQuery)
	_ = c.Bind(query) // Ignore bind errors for query parameters

	// Sanitize and set defaults for invalid/missing values
	query.SetDefaults()

	// Prepare view model
	vm := models.DashboardViewModel{
		Title:     "Dashboard - VibeFeeder",
		UserEmail: userEmail,
		Query:     query,
	}

	// Render dashboard template
	return c.Render(http.StatusOK, "", view.Index(vm))
}
