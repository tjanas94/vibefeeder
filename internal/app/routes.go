package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/home"
)

// setupRoutes configures all application routes
func (a *App) setupRoutes() {
	// Health check endpoint
	a.Echo.GET("/healthz", a.healthCheck)

	// Home routes
	homeHandler := home.NewHandler()
	a.Echo.GET("/", homeHandler.Index)
}

// healthCheck handler checks application and database health
func (a *App) healthCheck(c echo.Context) error {
	if err := a.DB.Health(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
