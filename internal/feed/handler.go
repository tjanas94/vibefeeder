package feed

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/feed/view"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	sharedmodels "github.com/tjanas94/vibefeeder/internal/shared/models"
	"github.com/tjanas94/vibefeeder/internal/shared/validator"
	sharedview "github.com/tjanas94/vibefeeder/internal/shared/view/components"
)

// FeedFetcher is an interface for triggering immediate feed fetches
type FeedFetcher interface {
	FetchFeedNow(feedID string)
}

// Handler handles HTTP requests for feed operations
type Handler struct {
	service     *Service
	logger      *slog.Logger
	feedFetcher FeedFetcher
}

// NewHandler creates a new feed handler
func NewHandler(service *Service, logger *slog.Logger, feedFetcher FeedFetcher) *Handler {
	return &Handler{
		service:     service,
		logger:      logger,
		feedFetcher: feedFetcher,
	}
}

// ListFeeds handles GET /feeds endpoint
// Returns a list of feeds for the authenticated user with filtering and pagination
func (h *Handler) ListFeeds(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Bind and validate query parameters
	query := new(models.ListFeedsQuery)
	if err := c.Bind(query); err != nil {
		h.logger.Warn("failed to bind query parameters", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
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
		// Return empty state with error message instead of HTTP error
		vm = &models.FeedListViewModel{
			Feeds:          []models.FeedItemViewModel{},
			ShowEmptyState: true,
			ErrorMessage:   "Failed to load feed list. Please refresh the page.",
			Pagination:     sharedmodels.PaginationViewModel{},
		}
	}

	// Build URL for HX-Push-Url header to update browser history
	pushURL := "/dashboard"
	params := make(url.Values)

	if query.Search != "" {
		params.Set("search", query.Search)
	}
	if query.Status != "" && query.Status != "all" {
		params.Set("status", query.Status)
	}
	if query.Page > 1 {
		params.Set("page", fmt.Sprintf("%d", query.Page))
	}

	if len(params) > 0 {
		pushURL += "?" + params.Encode()
	}

	c.Response().Header().Set("HX-Push-Url", pushURL)

	// Set HX-Trigger header to notify about feed availability
	hasFeeds := !vm.ShowEmptyState
	if hasFeeds {
		c.Response().Header().Set("HX-Trigger", `{"feedsLoaded": {"hasFeeds": true}}`)
	} else {
		c.Response().Header().Set("HX-Trigger", `{"feedsLoaded": {"hasFeeds": false}}`)
	}

	// Success - render list view with view model
	return c.Render(http.StatusOK, "", view.List(*vm))
}

// HandleFeedAddForm handles GET /feeds/new endpoint
// Returns an HTML form for adding a new feed
func (h *Handler) HandleFeedAddForm(c echo.Context) error {
	// Get user ID from authenticated session (check auth first)
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Create view model for adding a new feed
	vm := models.NewFeedFormForAdd()

	// Success - add HX-Trigger header to open modal and render form with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "feed"}}`)
	return c.Render(http.StatusOK, "", view.FeedForm(vm))
}

// CreateFeed handles POST /feeds endpoint
// Creates a new feed for the authenticated user
func (h *Handler) CreateFeed(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Authentication required",
		}
		vm := models.NewFeedFormWithErrors("add", "", "", "", errorVM)
		return c.Render(http.StatusUnauthorized, "", view.FeedForm(vm))
	}

	// Bind form data to command
	cmd := new(models.CreateFeedCommand)
	if err := c.Bind(cmd); err != nil {
		h.logger.Warn("failed to bind form data", "error", err)
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Invalid form data",
		}
		vm := models.NewFeedFormWithErrors("add", "", "", "", errorVM)
		return c.Render(http.StatusBadRequest, "", view.FeedForm(vm))
	}

	// Validate command
	if err := c.Validate(cmd); err != nil {
		h.logger.Warn("validation failed", "error", err)
		// Parse validation errors into view model
		fieldErrors := validator.ParseFieldErrors(err)
		errorVM := models.NewFeedFormErrorFromFieldErrors(fieldErrors)
		vm := models.NewFeedFormWithErrors("add", "", cmd.Name, cmd.URL, errorVM)
		return c.Render(http.StatusBadRequest, "", view.FeedForm(vm))
	}

	// Call service to create feed
	feedID, err := h.service.CreateFeed(c.Request().Context(), *cmd, userID)
	if err != nil {
		// Handle specific error types
		if err == ErrFeedAlreadyExists {
			h.logger.Info("duplicate feed attempt", "user_id", userID, "url", cmd.URL)
			errorVM := models.FeedFormErrorViewModel{
				URLError: "You have already added this feed",
			}
			vm := models.NewFeedFormWithErrors("add", "", cmd.Name, cmd.URL, errorVM)
			return c.Render(http.StatusConflict, "", view.FeedForm(vm))
		}

		// Handle other errors
		h.logger.Error("failed to create feed", "user_id", userID, "error", err)
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Failed to create feed. Please try again.",
		}
		vm := models.NewFeedFormWithErrors("add", "", cmd.Name, cmd.URL, errorVM)
		return c.Render(http.StatusInternalServerError, "", view.FeedForm(vm))
	}

	// Trigger immediate fetch for the newly created feed
	if h.feedFetcher != nil {
		h.feedFetcher.FetchFeedNow(feedID)
		h.logger.Info("Triggered immediate fetch for new feed", "feed_id", feedID)
	}

	// Success - refresh feed list, close modal and show toast
	c.Response().Header().Set("HX-Trigger", `{"refreshFeedList": null, "closeModal": null}`)
	return c.Render(http.StatusOK, "", sharedview.Toast(sharedview.ToastProps{
		Type:    "success",
		Message: "Feed was added",
		UseOOB:  true,
	}))
}

// HandleFeedEditForm handles GET /feeds/:id/edit endpoint
// Returns an HTML form pre-filled with the feed's current data for editing
func (h *Handler) HandleFeedEditForm(c echo.Context) error {
	// Get user ID from authenticated session (check auth first)
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		return err // Already returns echo.NewHTTPError
	}

	// Call service to get feed for editing
	vm, err := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
	if err != nil {
		// Handle specific error types
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			return echo.NewHTTPError(http.StatusNotFound, "Feed not found")
		}

		// Handle other errors
		h.logger.Error("failed to get feed for edit", "feed_id", feedID, "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load feed")
	}

	// Success - add HX-Trigger header to open modal and render form with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "feed"}}`)
	return c.Render(http.StatusOK, "", view.FeedForm(*vm))
}

// HandleUpdate handles PATCH /feeds/:id endpoint
// Updates an existing feed for the authenticated user
func (h *Handler) HandleUpdate(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Authentication required",
		}
		vm := models.NewFeedFormWithErrors("edit", "", "", "", errorVM)
		return c.Render(http.StatusUnauthorized, "", view.FeedForm(vm))
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Invalid feed ID",
		}
		vm := models.NewFeedFormWithErrors("edit", "", "", "", errorVM)
		return c.Render(http.StatusBadRequest, "", view.FeedForm(vm))
	}

	// Bind form data to command
	cmd := new(models.UpdateFeedCommand)
	if err := c.Bind(cmd); err != nil {
		h.logger.Warn("failed to bind form data", "feed_id", feedID, "error", err)
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Invalid form data",
		}
		vm := models.NewFeedFormWithErrors("edit", feedID, "", "", errorVM)
		return c.Render(http.StatusBadRequest, "", view.FeedForm(vm))
	}

	// Validate command
	if err := c.Validate(cmd); err != nil {
		h.logger.Warn("validation failed", "feed_id", feedID, "error", err)
		// Parse validation errors into view model
		fieldErrors := validator.ParseFieldErrors(err)
		errorVM := models.NewFeedFormErrorFromFieldErrors(fieldErrors)
		vm := models.NewFeedFormWithErrors("edit", feedID, cmd.Name, cmd.URL, errorVM)
		return c.Render(http.StatusBadRequest, "", view.FeedForm(vm))
	}

	// Call service to update feed
	urlChanged, err := h.service.UpdateFeed(c.Request().Context(), feedID, userID, *cmd)
	if err != nil {
		// Handle specific error types
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			errorVM := models.FeedFormErrorViewModel{
				GeneralError: "Feed not found",
			}
			vm := models.NewFeedFormWithErrors("edit", feedID, cmd.Name, cmd.URL, errorVM)
			return c.Render(http.StatusNotFound, "", view.FeedForm(vm))
		}

		if err == ErrFeedURLConflict {
			h.logger.Info("url already in use", "feed_id", feedID, "user_id", userID, "url", cmd.URL)
			errorVM := models.FeedFormErrorViewModel{
				URLError: "You have already added this feed",
			}
			vm := models.NewFeedFormWithErrors("edit", feedID, cmd.Name, cmd.URL, errorVM)
			return c.Render(http.StatusConflict, "", view.FeedForm(vm))
		}

		// Handle other errors
		h.logger.Error("failed to update feed", "feed_id", feedID, "user_id", userID, "error", err)
		errorVM := models.FeedFormErrorViewModel{
			GeneralError: "Failed to update feed. Please try again.",
		}
		vm := models.NewFeedFormWithErrors("edit", feedID, cmd.Name, cmd.URL, errorVM)
		return c.Render(http.StatusInternalServerError, "", view.FeedForm(vm))
	}

	// Trigger immediate fetch if URL changed
	if urlChanged && h.feedFetcher != nil {
		h.feedFetcher.FetchFeedNow(feedID)
		h.logger.Info("Triggered immediate fetch after URL change", "feed_id", feedID)
	}

	// Success - refresh feed list, close modal and show toast
	c.Response().Header().Set("HX-Trigger", `{"refreshFeedList": null, "closeModal": null}`)
	return c.Render(http.StatusOK, "", sharedview.Toast(sharedview.ToastProps{
		Type:    "success",
		Message: "Feed was updated",
		UseOOB:  true,
	}))
}

// HandleDeleteConfirmation handles GET /feeds/:id/delete endpoint
// Returns the delete confirmation modal content for the specified feed
func (h *Handler) HandleDeleteConfirmation(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		return err // Already returns echo.NewHTTPError
	}

	// Get feed name for confirmation (we need to check if feed exists and belongs to user)
	feed, err := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
	if err != nil {
		// Handle specific error types
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			return echo.NewHTTPError(http.StatusNotFound, "Feed not found")
		}

		// Handle other errors
		h.logger.Error("failed to get feed for delete confirmation", "feed_id", feedID, "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load feed")
	}

	// Create view model
	vm := models.DeleteConfirmationViewModel{
		FeedID:   feedID,
		FeedName: feed.Name,
	}

	// Success - add HX-Trigger header to open modal and render delete confirmation modal
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "delete"}}`)
	return c.Render(http.StatusOK, "", view.DeleteConfirmation(vm))
}

// DeleteFeed handles DELETE /feeds/:id endpoint
// Deletes a feed for the authenticated user
func (h *Handler) DeleteFeed(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)
	if userID == "" {
		h.logger.Error("missing user_id in context")
		vm := models.DeleteConfirmationViewModel{
			FeedID:       c.Param("id"),
			FeedName:     "",
			ErrorMessage: "Authentication required",
		}
		return c.Render(http.StatusUnauthorized, "", view.DeleteConfirmation(vm))
	}

	// Get and validate feed ID from path parameter
	feedID, err := h.getFeedID(c)
	if err != nil {
		// Get feed name for error display
		feed, feedErr := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
		if feedErr != nil {
			// If can't get feed, just use generic message
			vm := models.DeleteConfirmationViewModel{
				FeedID:       feedID,
				FeedName:     "",
				ErrorMessage: "Invalid feed ID",
			}
			return c.Render(http.StatusBadRequest, "", view.DeleteConfirmation(vm))
		}
		vm := models.DeleteConfirmationViewModel{
			FeedID:       feedID,
			FeedName:     feed.Name,
			ErrorMessage: "Invalid feed ID",
		}
		return c.Render(http.StatusBadRequest, "", view.DeleteConfirmation(vm))
	}

	// Call service to delete feed
	if err := h.service.DeleteFeed(c.Request().Context(), feedID, userID); err != nil {
		// Get feed name for error display
		feed, feedErr := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
		if feedErr != nil {
			// If can't get feed, use generic message
			vm := models.DeleteConfirmationViewModel{
				FeedID:       feedID,
				FeedName:     "",
				ErrorMessage: "Failed to delete feed",
			}
			return c.Render(http.StatusInternalServerError, "", view.DeleteConfirmation(vm))
		}

		// Handle specific error types
		var errorMessage string
		var statusCode int
		if err == ErrFeedNotFound {
			h.logger.Info("feed not found or unauthorized", "feed_id", feedID, "user_id", userID)
			errorMessage = "Feed not found"
			statusCode = http.StatusNotFound
		} else {
			// Handle other errors
			h.logger.Error("failed to delete feed", "feed_id", feedID, "user_id", userID, "error", err)
			errorMessage = "Failed to delete feed"
			statusCode = http.StatusInternalServerError
		}

		vm := models.DeleteConfirmationViewModel{
			FeedID:       feedID,
			FeedName:     feed.Name,
			ErrorMessage: errorMessage,
		}
		return c.Render(statusCode, "", view.DeleteConfirmation(vm))
	}

	// Success - refresh feed list, close modal and show toast
	c.Response().Header().Set("HX-Trigger", `{"refreshFeedList": null, "closeModal": null}`)
	return c.Render(http.StatusOK, "", sharedview.Toast(sharedview.ToastProps{
		Type:    "success",
		Message: "Feed was deleted",
		UseOOB:  true,
	}))
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
