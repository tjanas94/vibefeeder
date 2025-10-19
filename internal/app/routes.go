package app

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/supabase-community/gotrue-go"
	authModule "github.com/tjanas94/vibefeeder/internal/auth"
	"github.com/tjanas94/vibefeeder/internal/dashboard"
	"github.com/tjanas94/vibefeeder/internal/feed"
	"github.com/tjanas94/vibefeeder/internal/shared/ai"
	"github.com/tjanas94/vibefeeder/internal/shared/auth"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
	"github.com/tjanas94/vibefeeder/internal/summary"
	"golang.org/x/time/rate"
)

// setupRoutes configures all application routes
func (a *App) setupRoutes() {
	// Health check endpoint (public)
	a.Echo.GET("/healthz", a.healthCheck)

	// Initialize shared repositories
	eventsRepo := events.NewRepository(a.DB)

	// Initialize auth components
	// Create gotrue client with custom URL
	authClient := gotrue.New("dummy", a.Config.Supabase.Key).WithCustomGoTrueURL(a.Config.Supabase.URL + "/auth/v1")
	authService := authModule.NewService(authClient, eventsRepo, &a.Config.Auth, a.Logger)
	sessionManager := authModule.NewSessionManager(&a.Config.Auth)
	requireRegCode := a.Config.Auth.RegistrationCode != ""
	authHandler := authModule.NewHandler(authService, sessionManager, a.Logger, requireRegCode)

	// Create middleware adapter
	authMiddlewareAdapter := authModule.NewMiddlewareAdapter(authService)

	// Public routes (no authentication required)
	publicGroup := a.Echo.Group("/auth")
	publicGroup.GET("/login", authHandler.ShowLoginPage)
	publicGroup.POST("/login", authHandler.HandleLogin)
	publicGroup.GET("/register", authHandler.ShowRegisterPage)
	publicGroup.POST("/register", authHandler.HandleRegister)
	publicGroup.GET("/confirm", authHandler.HandleConfirm)
	publicGroup.GET("/forgot-password", authHandler.ShowForgotPasswordPage)
	publicGroup.POST("/forgot-password", authHandler.HandleForgotPassword)
	publicGroup.GET("/reset-password", authHandler.ShowResetPasswordPage)
	publicGroup.POST("/reset-password", authHandler.HandleResetPassword)

	// Protected routes (authentication required)
	protectedGroup := a.Echo.Group("")
	protectedGroup.Use(auth.AuthMiddleware(authMiddlewareAdapter, sessionManager, a.Logger))
	protectedGroup.Use(a.csrfMiddleware())

	// Root redirect to dashboard
	protectedGroup.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/dashboard")
	})

	// Logout endpoint
	protectedGroup.POST("/auth/logout", authHandler.HandleLogout)

	// Dashboard routes
	dashboardHandler := dashboard.NewHandler(a.Logger)
	protectedGroup.GET("/dashboard", dashboardHandler.ShowDashboard)

	// Feed routes
	feedRepo := feed.NewRepository(a.DB)
	feedService := feed.NewService(feedRepo, eventsRepo, a.Logger)
	feedHandler := feed.NewHandler(feedService, a.Logger, a.FeedFetcher)
	protectedGroup.GET("/feeds", feedHandler.ListFeeds)
	protectedGroup.GET("/feeds/new", feedHandler.HandleFeedAddForm)
	protectedGroup.POST("/feeds", feedHandler.CreateFeed)
	protectedGroup.GET("/feeds/:id/edit", feedHandler.HandleFeedEditForm)
	protectedGroup.PATCH("/feeds/:id", feedHandler.HandleUpdate)
	protectedGroup.GET("/feeds/:id/delete", feedHandler.HandleDeleteConfirmation)
	protectedGroup.DELETE("/feeds/:id", feedHandler.DeleteFeed)

	// Summary routes with rate limiting
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
	}

	aiService, err := ai.NewOpenRouterService(a.Config.OpenRouter, httpClient)
	if err != nil {
		a.Logger.Error("Failed to initialize OpenRouter service", "error", err)
		panic("OpenRouter service is required but failed to initialize: " + err.Error())
	}

	summaryRepo := summary.NewRepository(a.DB)
	summaryService := summary.NewService(summaryRepo, aiService, a.Logger, eventsRepo)
	summaryHandler := summary.NewHandler(summaryService, a.Logger)

	// Configure rate limiter from config
	rateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:  rate.Every(a.Config.RateLimit.SummaryGenerationInterval),
		Burst: 1,
	})

	protectedGroup.GET("/summaries/latest", summaryHandler.GetLatestSummary)
	protectedGroup.POST("/summaries", summaryHandler.GenerateSummary, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: rateLimiterStore,
		IdentifierExtractor: func(c echo.Context) (string, error) {
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
