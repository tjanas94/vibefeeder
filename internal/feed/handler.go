package feed

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/feed/view"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
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
	feedFetcher FeedFetcher
}

// NewHandler creates a new feed handler
func NewHandler(service *Service, feedFetcher FeedFetcher) *Handler {
	return &Handler{
		service:     service,
		feedFetcher: feedFetcher,
	}
}

// ListFeeds handles GET /feeds endpoint
// Returns a list of feeds for the authenticated user with filtering and pagination
func (h *Handler) ListFeeds(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)

	// Bind and sanitize query parameters
	query := new(models.ListFeedsQuery)
	_ = c.Bind(query) // Ignore bind errors for query parameters

	// Sanitize and set defaults for invalid/missing values
	query.SetDefaults()

	// Set user ID from authenticated session
	query.UserID = userID

	// Call service to get feeds
	vm, err := h.service.ListFeeds(c.Request().Context(), *query)
	if err != nil {
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			errVM := &models.FeedListViewModel{
				Feeds:          []models.FeedItemViewModel{},
				ShowEmptyState: true,
				ErrorMessage:   serviceErr.Message,
				Pagination:     sharedmodels.PaginationViewModel{},
			}
			return c.Render(serviceErr.Code, "", view.List(*errVM))
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// Build URL for HX-Push-Url header to update browser history
	pushURL := buildDashboardURL(*query)
	c.Response().Header().Set("HX-Push-Url", pushURL)

	// Set HX-Trigger header to notify about feed availability
	hasFeeds := !vm.ShowEmptyState
	c.Response().Header().Set("HX-Trigger", fmt.Sprintf(`{"feedsLoaded": {"hasFeeds": %t}}`, hasFeeds))

	// Success - render list view with view model
	return c.Render(http.StatusOK, "", view.List(*vm))
}

// HandleFeedAddForm handles GET /feeds/new endpoint
// Returns an HTML form for adding a new feed
func (h *Handler) HandleFeedAddForm(c echo.Context) error {
	// Create view model for adding a new feed
	vm := models.NewFeedFormForAdd()

	// Success - add HX-Trigger header to open modal and render form with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "feed"}}`)
	return c.Render(http.StatusOK, "", view.FeedForm(vm))
}

// CreateFeed handles POST /feeds endpoint
// Creates a new feed for the authenticated user
func (h *Handler) CreateFeed(c echo.Context) error {
	cmd := new(models.CreateFeedCommand)
	// Path 1: Handle bind errors (invalid request format)
	if err := c.Bind(cmd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Sanitize URL input
	cmd.URL = strings.TrimSpace(cmd.URL)

	// Get user ID from authenticated session
	cmd.UserID = auth.GetUserID(c)

	// Path 2: Handle validation errors (invalid data)
	if err := c.Validate(cmd); err != nil {
		// Parse validation errors into view model
		fieldErrors := validator.ParseFieldErrors(err)
		errorVM := models.NewFeedFormErrorFromFieldErrors(fieldErrors)
		vm := models.NewFeedFormWithErrors("add", "", cmd.Name, cmd.URL, errorVM)
		return c.Render(http.StatusUnprocessableEntity, "", view.FeedForm(vm))
	}

	// Call service to create feed
	feedID, err := h.service.CreateFeed(c.Request().Context(), *cmd)
	if err != nil {
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			return h.renderFormServiceError(c, serviceErr, "add", "", cmd.Name, cmd.URL)
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// Trigger immediate fetch for the newly created feed
	if h.feedFetcher != nil {
		h.feedFetcher.FetchFeedNow(feedID)
	}

	// Success - refresh feed list, close modal and show toast
	return h.renderSuccessToast(c, "Feed was added")
}

// HandleFeedEditForm handles GET /feeds/:id/edit endpoint
// Returns an HTML form pre-filled with the feed's current data for editing
func (h *Handler) HandleFeedEditForm(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)

	// Get feed ID from path parameter
	feedID := c.Param("id")

	// Call service to get feed for editing
	vm, err := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
	if err != nil {
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			return h.renderErrorToast(c, serviceErr.Code, serviceErr.Message)
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// Success - add HX-Trigger header to open modal and render form with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "feed"}}`)
	return c.Render(http.StatusOK, "", view.FeedForm(*vm))
}

// HandleUpdate handles PATCH /feeds/:id endpoint
// Updates an existing feed for the authenticated user
func (h *Handler) HandleUpdate(c echo.Context) error {
	cmd := new(models.UpdateFeedCommand)
	// Path 1: Handle bind errors (invalid request format)
	if err := c.Bind(cmd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Sanitize URL input
	cmd.URL = strings.TrimSpace(cmd.URL)

	// Get user ID from authenticated session
	cmd.UserID = auth.GetUserID(c)

	// Path 2: Handle validation errors (invalid data)
	if err := c.Validate(cmd); err != nil {
		// Parse validation errors into view model
		fieldErrors := validator.ParseFieldErrors(err)
		errorVM := models.NewFeedFormErrorFromFieldErrors(fieldErrors)
		vm := models.NewFeedFormWithErrors("edit", cmd.ID, cmd.Name, cmd.URL, errorVM)
		return c.Render(http.StatusUnprocessableEntity, "", view.FeedForm(vm))
	}

	// Call service to update feed
	urlChanged, err := h.service.UpdateFeed(c.Request().Context(), *cmd)
	if err != nil {
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			// For 404 errors, show error toast, close modal and refresh feed list
			if serviceErr.Code == http.StatusNotFound {
				return h.renderErrorToast(c, serviceErr.Code, serviceErr.Message)
			}

			// For other errors, show form with errors
			return h.renderFormServiceError(c, serviceErr, "edit", cmd.ID, cmd.Name, cmd.URL)
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// Trigger immediate fetch if URL changed
	if urlChanged && h.feedFetcher != nil {
		h.feedFetcher.FetchFeedNow(cmd.ID)
	}

	// Success - refresh feed list, close modal and show toast
	return h.renderSuccessToast(c, "Feed was updated")
}

// HandleDeleteConfirmation handles GET /feeds/:id/delete endpoint
// Returns the delete confirmation modal content for the specified feed
func (h *Handler) HandleDeleteConfirmation(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)

	// Get feed ID from path parameter
	feedID := c.Param("id")

	// Get feed name for confirmation (we need to check if feed exists and belongs to user)
	feed, err := h.service.GetFeedForEdit(c.Request().Context(), feedID, userID)
	if err != nil {
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			return h.renderErrorToast(c, serviceErr.Code, serviceErr.Message)
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
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

	// Get feed ID from path parameter
	feedID := c.Param("id")

	if err := h.service.DeleteFeed(c.Request().Context(), feedID, userID); err != nil {
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			return h.renderErrorToast(c, serviceErr.Code, serviceErr.Message)
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// Success - refresh feed list, close modal and show toast
	return h.renderSuccessToast(c, "Feed was deleted")
}

// renderErrorToast renders error toast with modal close and refresh headers
func (h *Handler) renderErrorToast(c echo.Context, statusCode int, message string) error {
	c.Response().Header().Set("HX-Reswap", "none")
	c.Response().Header().Set("HX-Trigger", `{"closeModal": null, "refreshFeedList": null}`)
	return c.Render(statusCode, "", sharedview.Toast(sharedview.ToastProps{
		Type:    "error",
		Message: message,
		UseOOB:  true,
	}))
}

// renderSuccessToast renders success toast with modal close and refresh headers
func (h *Handler) renderSuccessToast(c echo.Context, message string) error {
	c.Response().Header().Set("HX-Reswap", "none")
	c.Response().Header().Set("HX-Trigger", `{"closeModal": null, "refreshFeedList": null}`)
	return c.Render(http.StatusOK, "", sharedview.Toast(sharedview.ToastProps{
		Type:    "success",
		Message: message,
		UseOOB:  true,
	}))
}

// renderFormServiceError renders form with ServiceError details
func (h *Handler) renderFormServiceError(
	c echo.Context,
	serviceErr *sharederrors.ServiceError,
	mode string, // "add" or "edit"
	feedID string,
	name string,
	url string,
) error {
	errorVM := models.FeedFormErrorViewModel{
		GeneralError: serviceErr.Message,
		URLError:     serviceErr.FieldErrors["URL"],
	}
	vm := models.NewFeedFormWithErrors(mode, feedID, name, url, errorVM)
	return c.Render(serviceErr.Code, "", view.FeedForm(vm))
}
