package home

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/home/view"
)

// Handler handles home page requests
type Handler struct{}

// NewHandler creates a new home handler
func NewHandler() *Handler {
	return &Handler{}
}

// Index handles the home page request
func (h *Handler) Index(c echo.Context) error {
	if err := view.Index().Render(c.Request().Context(), c.Response().Writer); err != nil {
		return fmt.Errorf("failed to render home page: %w", err)
	}
	return nil
}
