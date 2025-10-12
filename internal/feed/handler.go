package feed

import (
	"fmt"
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
		fieldErrors := validator.ParseFieldErrors(err)
		vm := models.NewFeedFormErrorFromFieldErrors(fieldErrors)
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

// HandleFeedEditForm handles GET /feeds/:id/edit endpoint
// Returns an HTML form pre-filled with the feed's current data for editing
func (h *Handler) HandleFeedEditForm(c echo.Context) error {
	// Get user ID from authenticated session (check auth first)
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return h.renderError(c, http.StatusUnauthorized, "Authentication required")
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		return h.renderError(c, http.StatusBadRequest, "Invalid feed ID")
	}

	// Call service to get feed for editing
	vm, err := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
	if err != nil {
		// Handle specific error types
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			return h.renderError(c, http.StatusNotFound, "Feed not found")
		}

		// Handle other errors
		h.logger.Error("failed to get feed for edit", "feed_id", feedID, "user_id", userID, "error", err)
		return h.renderError(c, http.StatusInternalServerError, "Failed to load feed")
	}

	// Success - render edit form with view model
	return c.Render(http.StatusOK, "", view.FeedEditForm(*vm))
}

// HandleUpdate handles POST /feeds/:id endpoint
// Updates an existing feed for the authenticated user
func (h *Handler) HandleUpdate(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return h.renderUpdateFormError(c, "", http.StatusUnauthorized, models.FeedFormErrorViewModel{
			GeneralError: "Authentication required",
		})
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		return h.renderUpdateFormError(c, "", http.StatusBadRequest, models.FeedFormErrorViewModel{
			GeneralError: "Invalid feed ID",
		})
	}

	// Bind form data to command
	cmd := new(models.UpdateFeedCommand)
	if err := c.Bind(cmd); err != nil {
		h.logger.Warn("failed to bind form data", "feed_id", feedID, "error", err)
		return h.renderUpdateFormError(c, feedID, http.StatusBadRequest, models.FeedFormErrorViewModel{
			GeneralError: "Invalid form data",
		})
	}

	// Validate command
	if err := c.Validate(cmd); err != nil {
		h.logger.Warn("validation failed", "feed_id", feedID, "error", err)
		// Parse validation errors into view model
		fieldErrors := validator.ParseFieldErrors(err)
		vm := models.NewFeedFormErrorFromFieldErrors(fieldErrors)
		return h.renderUpdateFormError(c, feedID, http.StatusBadRequest, vm)
	}

	// Call service to update feed
	if err := h.service.UpdateFeed(c.Request().Context(), feedID, userID, *cmd); err != nil {
		// Handle specific error types
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			return h.renderUpdateFormError(c, feedID, http.StatusNotFound, models.FeedFormErrorViewModel{
				GeneralError: "Feed not found",
			})
		}

		if err == ErrFeedURLConflict {
			h.logger.Info("duplicate feed URL attempt", "feed_id", feedID, "user_id", userID, "url", cmd.URL)
			return h.renderUpdateFormError(c, feedID, http.StatusConflict, models.FeedFormErrorViewModel{
				URLError: "You have already added this feed",
			})
		}

		// Handle other errors
		h.logger.Error("failed to update feed", "feed_id", feedID, "user_id", userID, "error", err)
		return h.renderUpdateFormError(c, feedID, http.StatusInternalServerError, models.FeedFormErrorViewModel{
			GeneralError: "Failed to update feed. Please try again.",
		})
	}

	// Success - return 204 No Content with HX-Trigger header
	c.Response().Header().Set("HX-Trigger", "refreshFeedList")
	return c.NoContent(http.StatusNoContent)
}

// getFeedID extracts and validates feed ID from path parameter
func (h *Handler) getFeedID(c echo.Context) (string, error) {
	feedID := c.Param("id")
	if !validator.IsValidUUID(feedID) {
		h.logger.Warn("invalid feed id format", "feed_id", feedID)
		return "", echo.NewHTTPError(http.StatusBadRequest, "Invalid feed ID")
	}
	return feedID, nil
}

// renderFormError renders the form error view
func (h *Handler) renderFormError(c echo.Context, statusCode int, vm models.FeedFormErrorViewModel) error {
	// Set htmx headers for form error handling
	c.Response().Header().Set("HX-Retarget", "#feed-add-form-errors")
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return c.Render(statusCode, "", view.FeedFormErrors(vm))
}

// renderUpdateFormError renders the form error view for feed update
func (h *Handler) renderUpdateFormError(c echo.Context, feedID string, statusCode int, vm models.FeedFormErrorViewModel) error {
	// Set htmx headers for form error handling with dynamic feed ID
	targetID := fmt.Sprintf("#feed-edit-form-errors-%s", feedID)
	c.Response().Header().Set("HX-Retarget", targetID)
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return c.Render(statusCode, "", view.FeedFormErrors(vm))
}

// DeleteFeed handles DELETE /feeds/:id endpoint
// Deletes a feed for the authenticated user
func (h *Handler) DeleteFeed(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return c.NoContent(http.StatusUnauthorized)
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		return h.renderDeleteError(c, "", http.StatusBadRequest, "Invalid feed ID")
	}

	// Call service to delete feed
	if err := h.service.DeleteFeed(c.Request().Context(), feedID, userID); err != nil {
		// Handle specific error types
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			return h.renderDeleteError(c, feedID, http.StatusNotFound, "Feed not found")
		}

		// Handle other errors
		h.logger.Error("failed to delete feed", "feed_id", feedID, "user_id", userID, "error", err)
		return h.renderDeleteError(c, feedID, http.StatusInternalServerError, "Failed to delete feed")
	}

	// Success - return 204 No Content with HX-Trigger header
	c.Response().Header().Set("HX-Trigger", "refreshFeedList")
	return c.NoContent(http.StatusNoContent)
}

// renderDeleteError renders the error view for feed deletion with appropriate retargeting
func (h *Handler) renderDeleteError(c echo.Context, feedID string, statusCode int, message string) error {
	// Set htmx headers for error handling with dynamic feed ID
	if feedID != "" {
		targetID := fmt.Sprintf("#feed-item-%s-errors", feedID)
		c.Response().Header().Set("HX-Retarget", targetID)
		c.Response().Header().Set("HX-Reswap", "innerHTML")
	}

	vm := models.FeedListErrorViewModel{
		ErrorMessage: message,
	}
	return c.Render(statusCode, "", view.Error(vm))
}

// renderError renders the error view with appropriate error message
func (h *Handler) renderError(c echo.Context, statusCode int, message string) error {
	vm := models.FeedListErrorViewModel{
		ErrorMessage: message,
	}
	return c.Render(statusCode, "", view.Error(vm))
}
