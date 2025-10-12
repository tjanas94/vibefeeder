package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tjanas94/vibefeeder/internal/dashboard"
	"github.com/tjanas94/vibefeeder/internal/feed"
	"github.com/tjanas94/vibefeeder/internal/shared/ai"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	"github.com/tjanas94/vibefeeder/internal/summary"
	"golang.org/x/time/rate"
)

// setupRoutes configures all application routes
func (a *App) setupRoutes() {
	// Health check endpoint
	a.Echo.GET("/healthz", a.healthCheck)

	// Redirect root to dashboard
	a.Echo.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/dashboard")
	})

	// Dashboard routes (authenticated)
	dashboardHandler := dashboard.NewHandler(a.Logger)
	a.Echo.GET("/dashboard", dashboardHandler.ShowDashboard)

	// Feed routes (authenticated)
	feedService := feed.NewService(a.DB)
	feedHandler := feed.NewHandler(feedService, a.Logger)
	a.Echo.GET("/feeds", feedHandler.ListFeeds)
	a.Echo.POST("/feeds", feedHandler.CreateFeed)

	// Summary routes (authenticated with rate limiting)
	aiClient := ai.NewOpenRouterClient(a.Config.OpenRouter.APIKey)
	summaryService := summary.NewService(a.DB, aiClient, a.Logger)
	summaryHandler := summary.NewHandler(summaryService, a.Logger)

	// Configure rate limiter: 1 request per 5 minutes per user
	// rate.Every(5 * time.Minute) = 1 request every 5 minutes
	rateLimiterStore := middleware.NewRateLimiterMemoryStore(rate.Every(5 * 60)) // 1 request per 300 seconds

	// Create authenticated route group
	// Note: MockAuthMiddleware is already applied globally in setupMiddleware
	// TODO: Replace with real user ID from JWT when auth is implemented

	// Summary endpoints
	a.Echo.GET("/summaries/latest", summaryHandler.GetLatestSummary)

	// Apply rate limiting to summary generation endpoint (per user based on context)
	a.Echo.POST("/summaries", summaryHandler.GenerateSummary, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: rateLimiterStore,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Extract user ID from context for per-user rate limiting
			userID := auth.GetUserID(c)
			return userID, nil
		},
	}))
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
