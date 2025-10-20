package summary

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
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

// contextWithToken adds the access token from Echo context to request context for RLS
func (h *Handler) contextWithToken(c echo.Context) context.Context {
	token := auth.GetAccessToken(c)
	return database.ContextWithToken(c.Request().Context(), token)
}

// GenerateSummary handles POST /summaries endpoint
// Generates a new AI summary from user's recent articles
func (h *Handler) GenerateSummary(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)

	// Add access token to context for RLS
	ctx := h.contextWithToken(c)

	// Call service to generate summary and get view model
	vm, err := h.service.GenerateSummary(ctx, userID)

	// Domain / expected errors are embedded into the Display view via ErrorMessage
	if err != nil {
		switch {
		case errors.Is(err, ErrNoArticlesFound):
			h.logger.Info("no articles found for user", "error", err)
			errVM := &models.SummaryDisplayViewModel{
				ErrorMessage: err.Error(),
				// User has at least one feed if they attempted generation; allow retry button.
				CanGenerate: true,
			}
			return c.Render(http.StatusNotFound, "", view.Display(*errVM))
		case errors.Is(err, ErrAIServiceUnavailable):
			h.logger.Warn("AI service unavailable", "error", err)
			errVM := &models.SummaryDisplayViewModel{
				ErrorMessage: err.Error(),
				CanGenerate:  true,
			}
			return c.Render(http.StatusServiceUnavailable, "", view.Display(*errVM))
		case errors.Is(err, ErrDatabase):
			h.logger.Error("database error", "error", err)
			errVM := &models.SummaryDisplayViewModel{
				ErrorMessage: err.Error(),
				CanGenerate:  true,
			}
			return c.Render(http.StatusInternalServerError, "", view.Display(*errVM))
		default:
			h.logger.Error("unexpected summary generation error", "error", err)
			errVM := &models.SummaryDisplayViewModel{
				ErrorMessage: "An unexpected error occurred",
				CanGenerate:  true,
			}
			return c.Render(http.StatusInternalServerError, "", view.Display(*errVM))
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

	// Add access token to context for RLS
	ctx := h.contextWithToken(c)

	// Call service to get latest summary and view model
	vm, err := h.service.GetLatestSummaryForUser(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get latest summary", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load summary")
	}

	// Success - add HX-Trigger header to open modal and render display view with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "summary"}}`)
	return c.Render(http.StatusOK, "", view.Display(*vm))
}
