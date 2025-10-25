package summary

import (
	"context"
	"log/slog"
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/ai"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
)

const (
	// maxArticlesForSummary limits the number of articles fetched for summary generation
	maxArticlesForSummary = 100
)

// SummaryRepository defines the interface for summary data access
type SummaryRepository interface {
	FetchRecentArticles(ctx context.Context, userID string, limit int) ([]models.ArticleForPrompt, error)
	SaveSummary(ctx context.Context, userID, content string) (*database.PublicSummariesSelect, error)
	GetLatestSummary(ctx context.Context, userID string) (*database.PublicSummariesSelect, error)
	HasFeeds(ctx context.Context, userID string) (bool, error)
}

// AIClient defines the interface for AI service communication
type AIClient interface {
	GenerateChatCompletion(ctx context.Context, options ai.GenerateChatCompletionOptions) (*ai.ChatCompletionResponse, error)
}

// Service handles business logic for summary generation
type Service struct {
	repo       SummaryRepository
	aiClient   AIClient
	logger     *slog.Logger
	eventsRepo events.EventRepository
}

// NewService creates a new summary service
func NewService(repo SummaryRepository, aiClient AIClient, logger *slog.Logger, eventsRepo events.EventRepository) *Service {
	return &Service{
		repo:       repo,
		aiClient:   aiClient,
		logger:     logger,
		eventsRepo: eventsRepo,
	}
}

// GenerateSummary generates a new AI summary from user's articles from the last 24 hours
func (s *Service) GenerateSummary(ctx context.Context, userID string) (*models.SummaryDisplayViewModel, error) {
	// Step 1: Fetch articles from last 24 hours
	articles, err := s.repo.FetchRecentArticles(ctx, userID, maxArticlesForSummary)
	if err != nil {
		s.logger.Error("failed to fetch articles", "user_id", userID, "error", err)
		return nil, NewDatabaseError(err)
	}
	if len(articles) == 0 {
		return nil, NewNoArticlesFoundError()
	}

	// Step 2: Prepare prompt from articles
	prompt := buildPromptFromArticles(articles)

	// Step 3: Call AI service with timeout
	aiCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Prepare AI request options
	options := ai.GenerateChatCompletionOptions{
		Model:        "openai/gpt-4o-mini",
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
		Temperature:  0.7,
		MaxTokens:    2000,
	}

	response, err := s.aiClient.GenerateChatCompletion(aiCtx, options)
	if err != nil {
		s.logger.Error("AI service failed to generate summary", "user_id", userID, "error", err)
		return nil, NewAIServiceUnavailableError()
	}

	// Extract summary content from response
	summaryContent, err := extractSummaryContent(response)
	if err != nil {
		s.logger.Error("AI service returned empty response", "user_id", userID, "error", err)
		return nil, NewAIServiceUnavailableError()
	}

	// Step 4: Save summary to database
	dbSummary, err := s.repo.SaveSummary(ctx, userID, summaryContent)
	if err != nil {
		s.logger.Error("failed to save summary to database", "user_id", userID, "error", err)
		return nil, NewDatabaseError(err)
	}

	// Log summary_generated event
	if err := s.eventsRepo.RecordEvent(ctx, database.PublicEventsInsert{
		EventType: events.EventSummaryGenerated,
		UserId:    &userID,
		Metadata:  nil,
	}); err != nil {
		s.logger.Warn("Failed to log event", "event_type", events.EventSummaryGenerated, "error", err, "user_id", userID)
	}

	// Convert database type to view model
	vm := buildSummaryDisplayViewModel(dbSummary, true)
	return &vm, nil
}

// GetLatestSummaryForUser retrieves the latest summary for a user and determines if they can generate new ones
func (s *Service) GetLatestSummaryForUser(ctx context.Context, userID string) (*models.SummaryDisplayViewModel, error) {
	// Step 1: Fetch the latest summary for the user
	summary, err := s.repo.GetLatestSummary(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get latest summary", "user_id", userID, "error", err)
		return nil, NewDatabaseError(err)
	}

	// Step 2: Check if user has at least one feed
	canGenerate, err := s.repo.HasFeeds(ctx, userID)
	if err != nil {
		s.logger.Error("failed to check user feeds", "user_id", userID, "error", err)
		return nil, NewDatabaseError(err)
	}

	// Step 3: Build the view model
	vm := buildSummaryDisplayViewModel(summary, canGenerate)
	return &vm, nil
}

// buildSummaryDisplayViewModel is a pure function that transforms database result to view model
func buildSummaryDisplayViewModel(summary *database.PublicSummariesSelect, canGenerate bool) models.SummaryDisplayViewModel {
	vm := models.SummaryDisplayViewModel{
		CanGenerate: canGenerate,
	}

	if summary != nil {
		summaryVM := models.NewSummaryFromDB(*summary)
		vm.Summary = &summaryVM
	}

	return vm
}
