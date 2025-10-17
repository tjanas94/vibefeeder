package app

import (
	"net/http"
	"time"

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
	feedService := feed.NewService(a.DB, a.Logger)
	feedHandler := feed.NewHandler(feedService, a.Logger, a.FeedFetcher)
	a.Echo.GET("/feeds", feedHandler.ListFeeds)
	a.Echo.GET("/feeds/new", feedHandler.HandleFeedAddForm)
	a.Echo.POST("/feeds", feedHandler.CreateFeed)
	a.Echo.GET("/feeds/:id/edit", feedHandler.HandleFeedEditForm)
	a.Echo.PATCH("/feeds/:id", feedHandler.HandleUpdate)
	a.Echo.GET("/feeds/:id/delete", feedHandler.HandleDeleteConfirmation)
	a.Echo.DELETE("/feeds/:id", feedHandler.DeleteFeed)

	// Summary routes (authenticated with rate limiting)
	// Create HTTP client with timeout for OpenRouter API
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
	}

	aiService, err := ai.NewOpenRouterService(a.Config.OpenRouter, httpClient)
	if err != nil {
		a.Logger.Error("Failed to initialize OpenRouter service", "error", err)
		panic("OpenRouter service is required but failed to initialize: " + err.Error())
	}

	summaryService := summary.NewService(a.DB, aiService, a.Logger)
	summaryHandler := summary.NewHandler(summaryService, a.Logger)

	// Configure rate limiter from config (defaults to 30 seconds for testing, 300 for production)
	// Burst must be at least 1 to allow any requests (rate.Every returns fractional rate < 1)
	rateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:  rate.Every(a.Config.RateLimit.SummaryGenerationInterval),
		Burst: 1, // Allow 1 immediate request, then enforce rate limit
	})

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
