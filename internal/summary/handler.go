package summary

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

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
		return h.renderError(c, http.StatusUnauthorized, "Authentication required")
	}

	// Call service to generate summary
	summary, err := h.service.GenerateSummary(c.Request().Context(), userID)

	// Handle errors with appropriate HTTP status codes
	if err != nil {
		return h.handleServiceError(c, err)
	}

	// Success - render display view with summary data
	vm := models.SummaryDisplayViewModel{
		Summary: &models.SummaryViewModel{
			ID:        summary.Id,
			Content:   summary.Content,
			CreatedAt: parseTimestamp(summary.CreatedAt),
		},
		ShowEmptyState: false,
		CanGenerate:    true,
	}

	return c.Render(http.StatusOK, "", view.Display(vm))
}

// handleServiceError maps service errors to appropriate HTTP responses
func (h *Handler) handleServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNoFeeds):
		h.logger.Info("user has no feeds", "error", err)
		return h.renderError(c, http.StatusBadRequest, err.Error())

	case errors.Is(err, ErrNoArticlesFound):
		h.logger.Info("no articles found for user", "error", err)
		return h.renderError(c, http.StatusNotFound, err.Error())

	case errors.Is(err, ErrAIServiceUnavailable):
		h.logger.Warn("AI service unavailable", "error", err)
		return h.renderError(c, http.StatusServiceUnavailable, err.Error())

	case errors.Is(err, ErrDatabase):
		h.logger.Error("database error", "error", err)
		return h.renderError(c, http.StatusInternalServerError, err.Error())

	default:
		h.logger.Error("unexpected error", "error", err)
		return h.renderError(c, http.StatusInternalServerError, "An unexpected error occurred")
	}
}

// renderError renders the error view with appropriate error message
func (h *Handler) renderError(c echo.Context, statusCode int, message string) error {
	vm := models.SummaryErrorViewModel{
		ErrorMessage: message,
	}
	return c.Render(statusCode, "", view.Error(vm))
}

// parseTimestamp parses an RFC3339 timestamp string to time.Time
func parseTimestamp(timestamp string) time.Time {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return time.Time{}
	}
	return t
}
