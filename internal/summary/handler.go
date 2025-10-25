package summary

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
	"github.com/tjanas94/vibefeeder/internal/summary/view"
)

// Handler handles HTTP requests for summary operations
type Handler struct {
	service *Service
}

// NewHandler creates a new summary handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GenerateSummary handles POST /summaries endpoint
// Generates a new AI summary from user's recent articles
func (h *Handler) GenerateSummary(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)

	// Call service to generate summary and get view model
	vm, err := h.service.GenerateSummary(c.Request().Context(), userID)
	if err != nil {
		return h.handleServiceError(c, err, "service error during summary generation", "code", "message")
	}

	// Success - render display view with view model
	return c.Render(http.StatusOK, "", view.Display(*vm))
}

// GetLatestSummary handles GET /summaries/latest endpoint
// Retrieves and displays the latest summary for the authenticated user
func (h *Handler) GetLatestSummary(c echo.Context) error {
	// Get user ID from authenticated session
	userID := auth.GetUserID(c)

	// Call service to get latest summary and view model
	vm, err := h.service.GetLatestSummaryForUser(c.Request().Context(), userID)
	if err != nil {
		return h.handleServiceError(c, err, "service error getting latest summary", "user_id", userID, "code")
	}

	// Success - add HX-Trigger header to open modal and render display view with view model
	c.Response().Header().Set("HX-Trigger", `{"openModal": {"modal": "summary"}}`)
	return c.Render(http.StatusOK, "", view.Display(*vm))
}

// handleServiceError handles ServiceError responses with logging and error view rendering.
// If err is a ServiceError, logs a warning and renders an error view.
// If err is not a ServiceError, returns the error for global error handler processing.
func (h *Handler) handleServiceError(c echo.Context, err error, logMsg string, logAttrs ...any) error {
	var serviceErr *sharederrors.ServiceError
	if errors.As(err, &serviceErr) {
		errVM := &models.SummaryDisplayViewModel{
			ErrorMessage: serviceErr.Message,
			CanGenerate:  true,
		}
		return c.Render(serviceErr.Code, "", view.Display(*errVM))
	}
	return err
}
