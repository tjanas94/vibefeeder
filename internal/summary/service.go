package summary

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/supabase-community/postgrest-go"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
)

// AIClient defines the interface for AI service communication
type AIClient interface {
	GenerateSummary(ctx context.Context, prompt string) (string, error)
}

// Service handles business logic for summary generation
type Service struct {
	db       *database.Client
	aiClient AIClient
	logger   *slog.Logger
}

// NewService creates a new summary service
func NewService(db *database.Client, aiClient AIClient, logger *slog.Logger) *Service {
	return &Service{
		db:       db,
		aiClient: aiClient,
		logger:   logger,
	}
}

// GenerateSummary generates a new AI summary from user's articles from the last 24 hours
func (s *Service) GenerateSummary(ctx context.Context, userID string) (*models.SummaryDisplayViewModel, error) {
	// Step 1: Fetch articles from last 24 hours
	articles, err := s.fetchRecentArticles(ctx, userID)
	if err != nil {
		s.logger.Error("failed to fetch articles", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to fetch articles: %w", err)
	}
	if len(articles) == 0 {
		return nil, ErrNoArticlesFound
	}

	// Step 2: Prepare prompt from articles
	prompt := buildPromptFromArticles(articles)

	// Step 3: Call AI service with timeout
	aiCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	summaryContent, err := s.aiClient.GenerateSummary(aiCtx, prompt)
	if err != nil {
		s.logger.Error("AI service failed to generate summary", "user_id", userID, "error", err)
		return nil, ErrAIServiceUnavailable
	}

	// Step 4: Save summary to database
	dbSummary, err := s.saveSummary(ctx, userID, summaryContent)
	if err != nil {
		s.logger.Error("failed to save summary to database", "user_id", userID, "error", err)
		return nil, ErrDatabase
	}

	// Step 5: Record event
	if err := s.recordSummaryEvent(ctx, userID); err != nil {
		// Log error but don't fail the request
		s.logger.Warn("failed to record summary event", "user_id", userID, "error", err)
	}

	// Convert database type to view model
	summaryVM := models.NewSummaryFromDB(*dbSummary)
	return &models.SummaryDisplayViewModel{
		Summary:        &summaryVM,
		ShowEmptyState: false,
		CanGenerate:    true, // User just generated, so they have feeds
	}, nil
}

// fetchRecentArticles retrieves all articles published in the last 24 hours for the user's feeds
func (s *Service) fetchRecentArticles(ctx context.Context, userID string) ([]models.ArticleForPrompt, error) {
	// Calculate 24 hours ago timestamp
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	var articles []models.ArticleForPrompt

	// Query articles joined with feeds to filter by user_id
	_, err := s.db.From("articles").
		Select("title, content, feeds!inner(user_id)", "", false).
		Eq("feeds.user_id", userID).
		Gte("published_at", twentyFourHoursAgo).
		Order("published_at", nil).
		ExecuteTo(&articles)

	if err != nil {
		return nil, err
	}

	return articles, nil
}

// saveSummary stores the generated summary in the database
func (s *Service) saveSummary(ctx context.Context, userID, content string) (*database.PublicSummariesSelect, error) {
	insert := models.ToInsert(userID, content)

	var result []database.PublicSummariesSelect
	_, err := s.db.From("summaries").
		Insert(insert, false, "", "*", "").
		ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no summary returned after insert")
	}

	return &result[0], nil
}

// recordSummaryEvent records a summary_generated event in the events table
func (s *Service) recordSummaryEvent(ctx context.Context, userID string) error {
	event := database.PublicEventsInsert{
		EventType: "summary_generated",
		UserId:    &userID,
		Metadata:  map[string]interface{}{},
	}

	var result []database.PublicEventsSelect
	_, err := s.db.From("events").
		Insert(event, false, "", "*", "").
		ExecuteTo(&result)

	return err
}

// GetLatestSummaryForUser retrieves the latest summary for a user and determines if they can generate new ones
func (s *Service) GetLatestSummaryForUser(ctx context.Context, userID string) (*models.SummaryDisplayViewModel, error) {
	// Step 1: Fetch the latest summary for the user
	var summaries []database.PublicSummariesSelect
	_, err := s.db.From("summaries").
		Select("*", "", false).
		Eq("user_id", userID).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Limit(1, "").
		ExecuteTo(&summaries)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest summary: %w", err)
	}

	// Step 2: Check if user has at least one feed
	var feeds []database.PublicFeedsSelect
	_, err = s.db.From("feeds").
		Select("id", "", false).
		Eq("user_id", userID).
		Limit(1, "").
		ExecuteTo(&feeds)

	if err != nil {
		return nil, fmt.Errorf("failed to check user feeds: %w", err)
	}

	canGenerate := len(feeds) > 0

	// Step 3: Build the view model
	vm := &models.SummaryDisplayViewModel{
		ShowEmptyState: len(summaries) == 0,
		CanGenerate:    canGenerate,
	}

	// If summary exists, convert it to view model
	if len(summaries) > 0 {
		summaryVM := models.NewSummaryFromDB(summaries[0])
		vm.Summary = &summaryVM
	}

	return vm, nil
}
