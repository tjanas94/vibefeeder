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
		return h.renderError(c, http.StatusUnauthorized, "Authentication required")
	}

	// Call service to generate summary and get view model
	vm, err := h.service.GenerateSummary(c.Request().Context(), userID)

	// Handle errors with appropriate HTTP status codes
	if err != nil {
		return h.handleServiceError(c, err)
	}

	// Success - render display view with view model
	return c.Render(http.StatusOK, "", view.Display(*vm))
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

// GetLatestSummary handles GET /summaries/latest endpoint
// Retrieves and displays the latest summary for the authenticated user
func (h *Handler) GetLatestSummary(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return h.renderError(c, http.StatusUnauthorized, "Authentication required")
	}

	// Call service to get latest summary and view model
	vm, err := h.service.GetLatestSummaryForUser(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("failed to get latest summary", "user_id", userID, "error", err)
		return h.renderError(c, http.StatusInternalServerError, "Failed to load summary")
	}

	// Success - render display view with view model
	return c.Render(http.StatusOK, "", view.Display(*vm))
}
