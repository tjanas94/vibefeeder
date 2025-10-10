package summary

import (
	"context"
	"fmt"
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
}

// NewService creates a new summary service
func NewService(db *database.Client, aiClient AIClient) *Service {
	return &Service{
		db:       db,
		aiClient: aiClient,
	}
}

// GenerateSummary generates a new AI summary from user's articles from the last 24 hours
func (s *Service) GenerateSummary(ctx context.Context, userID string) (*database.PublicSummariesSelect, error) {
	// Step 1: Check if user has any feeds
	hasFeeds, err := s.checkUserHasFeeds(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user feeds: %w", err)
	}
	if !hasFeeds {
		return nil, ErrNoFeeds
	}

	// Step 2: Fetch articles from last 24 hours
	articles, err := s.fetchRecentArticles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch articles: %w", err)
	}
	if len(articles) == 0 {
		return nil, ErrNoArticlesFound
	}

	// Step 3: Prepare prompt from articles
	prompt := buildPromptFromArticles(articles)

	// Step 4: Call AI service with timeout
	aiCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	summaryContent, err := s.aiClient.GenerateSummary(aiCtx, prompt)
	if err != nil {
		return nil, ErrAIServiceUnavailable
	}

	// Step 5: Save summary to database
	summary, err := s.saveSummary(ctx, userID, summaryContent)
	if err != nil {
		return nil, ErrDatabase
	}

	// Step 6: Record event
	if err := s.recordSummaryEvent(ctx, userID); err != nil {
		// Log error but don't fail the request
		// In production, you would use slog here
		fmt.Printf("failed to record summary event: %v\n", err)
	}

	return summary, nil
}

// checkUserHasFeeds verifies if the user has at least one feed configured
func (s *Service) checkUserHasFeeds(ctx context.Context, userID string) (bool, error) {
	var feeds []database.PublicFeedsSelect
	_, err := s.db.From("feeds").
		Select("id", "", false).
		Eq("user_id", userID).
		Limit(1, "").
		ExecuteTo(&feeds)

	if err != nil {
		return false, err
	}

	return len(feeds) > 0, nil
}

// fetchRecentArticles retrieves all articles published in the last 24 hours for the user's feeds
func (s *Service) fetchRecentArticles(ctx context.Context, userID string) ([]database.PublicArticlesSelect, error) {
	// Calculate 24 hours ago timestamp
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	var articles []database.PublicArticlesSelect

	// Query articles joined with feeds to filter by user_id
	// We need to get feed_ids for this user first
	var feeds []database.PublicFeedsSelect
	_, err := s.db.From("feeds").
		Select("id", "", false).
		Eq("user_id", userID).
		ExecuteTo(&feeds)

	if err != nil {
		return nil, err
	}

	if len(feeds) == 0 {
		return []database.PublicArticlesSelect{}, nil
	}

	// Extract feed IDs
	feedIDs := make([]string, len(feeds))
	for i, feed := range feeds {
		feedIDs[i] = feed.Id
	}

	// Fetch articles for these feeds published in last 24h
	_, err = s.db.From("articles").
		Select("*", "", false).
		In("feed_id", feedIDs).
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

	// Step 2: Check if user has at least one working feed (last_fetch_status = 'success')
	var feeds []database.PublicFeedsSelect
	_, err = s.db.From("feeds").
		Select("id", "", false).
		Eq("user_id", userID).
		Eq("last_fetch_status", "success").
		Limit(1, "").
		ExecuteTo(&feeds)

	if err != nil {
		return nil, fmt.Errorf("failed to check working feeds: %w", err)
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
