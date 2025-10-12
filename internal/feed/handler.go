package feed

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/feed/view"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
)

// Handler handles HTTP requests for feed operations
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new feed handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// ListFeeds handles GET /feeds endpoint
// Returns a list of feeds for the authenticated user with filtering and pagination
func (h *Handler) ListFeeds(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return h.renderError(c, http.StatusUnauthorized, "Authentication required")
	}

	// Bind and validate query parameters
	query := new(models.ListFeedsQuery)
	if err := c.Bind(query); err != nil {
		h.logger.Warn("failed to bind query parameters", "error", err)
		return h.renderError(c, http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults for optional parameters
	query.SetDefaults()

	// Validate query parameters
	if err := c.Validate(query); err != nil {
		h.logger.Warn("invalid query parameters", "error", err)
		return err // Echo validator returns formatted error
	}

	// Set user ID from authenticated session
	query.UserID = userID

	// Call service to get feeds
	vm, err := h.service.ListFeeds(c.Request().Context(), *query)
	if err != nil {
		h.logger.Error("failed to list feeds", "user_id", userID, "error", err)
		return h.renderError(c, http.StatusInternalServerError, "Failed to load feeds")
	}

	// Success - render list view with view model
	return c.Render(http.StatusOK, "", view.List(*vm))
}

// renderError renders the error view with appropriate error message
func (h *Handler) renderError(c echo.Context, statusCode int, message string) error {
	vm := models.FeedListErrorViewModel{
		ErrorMessage: message,
	}
	return c.Render(statusCode, "", view.Error(vm))
}
