package summary

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
	"github.com/tjanas94/vibefeeder/internal/summary/view"
)

// Handler handles HTTP requests for summary operations
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new summary handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GenerateSummary handles POST /summaries endpoint
// Generates a new AI summary from user's recent articles
func (h *Handler) GenerateSummary(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Call service to generate summary and get view model
	vm, err := h.service.GenerateSummary(c.Request().Context(), userID)

	// Domain / expected errors are embedded into the Display view via ErrorMessage
	if err != nil {
		switch {
		case errors.Is(err, ErrNoArticlesFound),
			errors.Is(err, ErrAIServiceUnavailable):
			errVM := &models.SummaryDisplayViewModel{
				ErrorMessage: err.Error(),
				// User has at least one feed if they attempted generation; allow retry button.
				CanGenerate: true,
			}
			return c.Render(http.StatusOK, "", view.Display(*errVM))
		default:
			h.logger.Error("unexpected summary generation error", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process summary request")
		}
	}

	// Success - render display view with view model
	return c.Render(http.StatusOK, "", view.Display(*vm))
}

// GetLatestSummary handles GET /summaries/latest endpoint
// Retrieves and displays the latest summary for the authenticated user
func (h *Handler) GetLatestSummary(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Call service to get latest summary and view model
	vm, err := h.service.GetLatestSummaryForUser(c.Request().Context(), userID)
	if err != nil {
		// Treat load failure as embedded error view instead of global handler
		h.logger.Error("failed to get latest summary", "user_id", userID, "error", err)
		errVM := models.SummaryDisplayViewModel{
			ErrorMessage:   "Failed to load summary",
			CanGenerate:    false,
			ShowEmptyState: false,
		}
		// Still open modal so user sees the error
		c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "summary"}}`)
		return c.Render(http.StatusOK, "", view.Display(errVM))
	}

	// Success - add HX-Trigger header to open modal and render display view with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "summary"}}`)
	return c.Render(http.StatusOK, "", view.Display(*vm))
}
