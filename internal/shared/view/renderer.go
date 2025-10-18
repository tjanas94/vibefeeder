package view

import (
	"io"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/shared/csrf"
)

// TemplRenderer implements echo.Renderer interface for Templ components
type TemplRenderer struct{}

// NewTemplRenderer creates a new Templ renderer
func NewTemplRenderer() *TemplRenderer {
	return &TemplRenderer{}
}

// Render implements echo.Renderer interface
// The name parameter should be ignored as Templ uses type-safe components
// The data parameter must be a templ.Component
func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	component, ok := data.(templ.Component)
	if !ok {
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "data must be a templ.Component")
	}

	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	// Create a new context with CSRF token from Echo context
	ctx := c.Request().Context()
	if csrfToken, ok := c.Get(csrf.EchoContextKey).(string); ok && csrfToken != "" {
		ctx = csrf.WithToken(ctx, csrfToken)
	}

	if err := component.Render(ctx, buf); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}
