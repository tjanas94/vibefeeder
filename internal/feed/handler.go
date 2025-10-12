package feed

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/feed/view"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	"github.com/tjanas94/vibefeeder/internal/shared/validator"
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

// CreateFeed handles POST /feeds endpoint
// Creates a new feed for the authenticated user
func (h *Handler) CreateFeed(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return h.renderFormError(c, http.StatusUnauthorized, models.FeedFormErrorViewModel{
			GeneralError: "Authentication required",
		})
	}

	// Bind form data to command
	cmd := new(models.CreateFeedCommand)
	if err := c.Bind(cmd); err != nil {
		h.logger.Warn("failed to bind form data", "error", err)
		return h.renderFormError(c, http.StatusBadRequest, models.FeedFormErrorViewModel{
			GeneralError: "Invalid form data",
		})
	}

	// Validate command
	if err := c.Validate(cmd); err != nil {
		h.logger.Warn("validation failed", "error", err)
		// Parse validation errors into view model
		vm := parseValidationErrors(err)
		return h.renderFormError(c, http.StatusBadRequest, vm)
	}

	// Call service to create feed
	if err := h.service.CreateFeed(c.Request().Context(), *cmd, userID); err != nil {
		// Handle specific error types
		if err == ErrFeedAlreadyExists {
			h.logger.Info("duplicate feed attempt", "user_id", userID, "url", cmd.URL)
			return h.renderFormError(c, http.StatusConflict, models.FeedFormErrorViewModel{
				URLError: "You have already added this feed",
			})
		}

		// Handle other errors
		h.logger.Error("failed to create feed", "user_id", userID, "error", err)
		return h.renderFormError(c, http.StatusInternalServerError, models.FeedFormErrorViewModel{
			GeneralError: "Failed to create feed. Please try again.",
		})
	}

	// Success - return 204 No Content with HX-Trigger header
	c.Response().Header().Set("HX-Trigger", "refreshFeedList")
	return c.NoContent(http.StatusNoContent)
}

// parseValidationErrors converts validation errors to FeedFormErrorViewModel
func parseValidationErrors(err error) models.FeedFormErrorViewModel {
	vm := models.FeedFormErrorViewModel{}

	// Parse field errors using shared validator package
	fieldErrors := validator.ParseFieldErrors(err)
	if fieldErrors == nil {
		vm.GeneralError = "Invalid request"
		return vm
	}

	// Map parsed errors to view model fields
	if nameErr, ok := fieldErrors["Name"]; ok {
		vm.NameError = nameErr
	}
	if urlErr, ok := fieldErrors["URL"]; ok {
		vm.URLError = urlErr
	}

	return vm
}

// renderFormError renders the form error view
func (h *Handler) renderFormError(c echo.Context, statusCode int, vm models.FeedFormErrorViewModel) error {
	// Set htmx headers for form error handling
	c.Response().Header().Set("HX-Retarget", "#feed-add-form-errors")
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return c.Render(statusCode, "", view.FeedFormErrors(vm))
}

// renderError renders the error view with appropriate error message
func (h *Handler) renderError(c echo.Context, statusCode int, message string) error {
	vm := models.FeedListErrorViewModel{
		ErrorMessage: message,
	}
	return c.Render(statusCode, "", view.Error(vm))
}
