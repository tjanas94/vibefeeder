package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
)

// setupRoutes configures all application routes
func (a *App) setupRoutes() {
	c := a.Container

	// Health check endpoint (public)
	a.Echo.GET("/healthz", a.healthCheck)

	// Public routes (no authentication required)
	publicGroup := a.Echo.Group("/auth")
	publicGroup.GET("/login", c.AuthHandler.ShowLoginPage)
	publicGroup.POST("/login", c.AuthHandler.HandleLogin)
	publicGroup.GET("/register", c.AuthHandler.ShowRegisterPage)
	publicGroup.POST("/register", c.AuthHandler.HandleRegister)
	publicGroup.GET("/confirm", c.AuthHandler.HandleConfirm)
	publicGroup.GET("/forgot-password", c.AuthHandler.ShowForgotPasswordPage)
	publicGroup.POST("/forgot-password", c.AuthHandler.HandleForgotPassword)
	publicGroup.GET("/reset-password", c.AuthHandler.ShowResetPasswordPage)
	publicGroup.POST("/reset-password", c.AuthHandler.HandleResetPassword)

	// Protected routes (authentication required)
	protectedGroup := a.Echo.Group("")
	protectedGroup.Use(auth.AuthMiddleware(c.AuthService, c.SessionManager, c.Logger))
	protectedGroup.Use(a.csrfMiddleware())

	// Root redirect to dashboard
	protectedGroup.GET("/", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusFound, "/dashboard")
	})

	// Logout endpoint
	protectedGroup.POST("/auth/logout", c.AuthHandler.HandleLogout)

	// Dashboard routes
	protectedGroup.GET("/dashboard", c.DashboardHandler.ShowDashboard)

	// Feed routes
	protectedGroup.GET("/feeds", c.FeedHandler.ListFeeds)
	protectedGroup.GET("/feeds/new", c.FeedHandler.HandleFeedAddForm)
	protectedGroup.POST("/feeds", c.FeedHandler.CreateFeed)
	protectedGroup.GET("/feeds/:id/edit", c.FeedHandler.HandleFeedEditForm)
	protectedGroup.PATCH("/feeds/:id", c.FeedHandler.HandleUpdate)
	protectedGroup.GET("/feeds/:id/delete", c.FeedHandler.HandleDeleteConfirmation)
	protectedGroup.DELETE("/feeds/:id", c.FeedHandler.DeleteFeed)

	// Summary routes with rate limiting
	protectedGroup.GET("/summaries/latest", c.SummaryHandler.GetLatestSummary)
	protectedGroup.POST("/summaries", c.SummaryHandler.GenerateSummary, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: c.RateLimiterStore,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			userID := auth.GetUserID(ctx)
			return userID, nil
		},
	}))
}

// healthCheck handler checks application and database health
func (a *App) healthCheck(c echo.Context) error {
	if err := a.Container.DB.Health(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
