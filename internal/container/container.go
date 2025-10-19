package container

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"github.com/supabase-community/gotrue-go"
	authModule "github.com/tjanas94/vibefeeder/internal/auth"
	"github.com/tjanas94/vibefeeder/internal/dashboard"
	"github.com/tjanas94/vibefeeder/internal/feed"
	"github.com/tjanas94/vibefeeder/internal/fetcher"
	"github.com/tjanas94/vibefeeder/internal/shared/ai"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
	"github.com/tjanas94/vibefeeder/internal/summary"
	"golang.org/x/time/rate"
)

// Container holds all application dependencies
type Container struct {
	// Configuration and infrastructure
	Config *config.Config
	DB     *database.Client
	Logger *slog.Logger
	Ctx    context.Context

	// Repositories
	EventsRepo  *events.Repository
	FeedRepo    *feed.Repository
	SummaryRepo *summary.Repository
	FetcherRepo *fetcher.Repository

	// Services
	AuthService    *authModule.Service
	FeedService    *feed.Service
	SummaryService *summary.Service
	AIService      *ai.OpenRouterService
	FeedFetcher    *fetcher.FeedFetcherService

	// Handlers
	AuthHandler      *authModule.Handler
	DashboardHandler *dashboard.Handler
	FeedHandler      *feed.Handler
	SummaryHandler   *summary.Handler

	// Middleware and utilities
	SessionManager        *authModule.SessionManager
	AuthMiddlewareAdapter *authModule.MiddlewareAdapter
	RateLimiterStore      *middleware.RateLimiterMemoryStore
}

// New creates and initializes a new dependency injection container
func New(cfg *config.Config, db *database.Client, logger *slog.Logger, ctx context.Context) (*Container, error) {
	c := &Container{
		Config: cfg,
		DB:     db,
		Logger: logger,
		Ctx:    ctx,
	}

	// Initialize repositories
	if err := c.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize services (depends on repositories)
	if err := c.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize middleware and utilities (depends on services)
	if err := c.initMiddleware(); err != nil {
		return nil, fmt.Errorf("failed to initialize middleware: %w", err)
	}

	// Initialize handlers (depends on services and middleware)
	if err := c.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	return c, nil
}

// initRepositories initializes all repository instances
func (c *Container) initRepositories() error {
	c.EventsRepo = events.NewRepository(c.DB)
	c.FeedRepo = feed.NewRepository(c.DB)
	c.SummaryRepo = summary.NewRepository(c.DB)
	c.FetcherRepo = fetcher.NewRepository(c.DB)

	return nil
}

// initServices initializes all service instances
func (c *Container) initServices() error {
	// Initialize auth service
	authClient := gotrue.New("dummy", c.Config.Supabase.Key).
		WithCustomGoTrueURL(c.Config.Supabase.URL + "/auth/v1")
	c.AuthService = authModule.NewService(authClient, c.EventsRepo, &c.Config.Auth, c.Logger)

	// Initialize feed service
	c.FeedService = feed.NewService(c.FeedRepo, c.EventsRepo, c.Logger)

	// Initialize AI service
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
	}
	aiService, err := ai.NewOpenRouterService(c.Config.OpenRouter, httpClient)
	if err != nil {
		c.Logger.Error("Failed to initialize OpenRouter service", "error", err)
		return fmt.Errorf("OpenRouter service is required but failed to initialize: %w", err)
	}
	c.AIService = aiService

	// Initialize summary service
	c.SummaryService = summary.NewService(c.SummaryRepo, c.AIService, c.Logger, c.EventsRepo)

	// Initialize feed fetcher service
	fetcherHTTPClient := fetcher.NewHTTPClient(fetcher.HTTPClientConfig{
		Timeout:         c.Config.Fetcher.RequestTimeout,
		FollowRedirects: false,
		Logger:          c.Logger,
	})
	c.FeedFetcher = fetcher.NewFeedFetcherService(
		c.FetcherRepo,
		fetcherHTTPClient,
		c.Logger,
		c.Config.Fetcher,
		c.Ctx,
	)

	return nil
}

// initMiddleware initializes middleware and utility instances
func (c *Container) initMiddleware() error {
	// Initialize session manager
	c.SessionManager = authModule.NewSessionManager(&c.Config.Auth)

	// Initialize auth middleware adapter
	c.AuthMiddlewareAdapter = authModule.NewMiddlewareAdapter(c.AuthService)

	// Initialize rate limiter store
	c.RateLimiterStore = middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:  rate.Every(c.Config.RateLimit.SummaryGenerationInterval),
			Burst: 1,
		},
	)

	return nil
}

// initHandlers initializes all handler instances
func (c *Container) initHandlers() error {
	// Initialize auth handler
	requireRegCode := c.Config.Auth.RegistrationCode != ""
	c.AuthHandler = authModule.NewHandler(c.AuthService, c.SessionManager, c.Logger, requireRegCode)

	// Initialize dashboard handler
	c.DashboardHandler = dashboard.NewHandler(c.Logger)

	// Initialize feed handler
	c.FeedHandler = feed.NewHandler(c.FeedService, c.Logger, c.FeedFetcher)

	// Initialize summary handler
	c.SummaryHandler = summary.NewHandler(c.SummaryService, c.Logger)

	return nil
}
