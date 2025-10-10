package home

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/home/view"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// Handler handles home page requests
type Handler struct {
	db *database.Client
}

// NewHandler creates a new home handler
func NewHandler(db *database.Client) *Handler {
	return &Handler{
		db: db,
	}
}

// Index handles the home page request
func (h *Handler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "", view.Index())
}
