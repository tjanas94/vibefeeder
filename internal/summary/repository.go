package summary

import (
	"context"
	"fmt"
	"time"

	"github.com/supabase-community/postgrest-go"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
)

// Repository handles data access for summaries
type Repository struct {
	db *database.Client
}

// Ensure Repository implements SummaryRepository interface at compile time
var _ SummaryRepository = (*Repository)(nil)

// NewRepository creates a new summary repository
func NewRepository(db *database.Client) *Repository {
	return &Repository{db: db}
}

// FetchRecentArticles retrieves articles published in the last 24 hours for the user's feeds
// Limited to maxArticlesForSummary most recent articles
func (r *Repository) FetchRecentArticles(ctx context.Context, userID string, limit int) ([]models.ArticleForPrompt, error) {
	// Calculate 24 hours ago timestamp
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	var articles []models.ArticleForPrompt

	// Query articles joined with feeds to filter by user_id
	_, err := r.db.From("articles").
		Select("title, content, feeds!inner(user_id)", "", false).
		Eq("feeds.user_id", userID).
		Gte("published_at", twentyFourHoursAgo).
		Order("published_at", &postgrest.OrderOpts{Ascending: false}).
		Limit(limit, "").
		ExecuteTo(&articles)

	if err != nil {
		return nil, err
	}

	return articles, nil
}

// SaveSummary stores the generated summary in the database
func (r *Repository) SaveSummary(ctx context.Context, userID, content string) (*database.PublicSummariesSelect, error) {
	insert := models.ToInsert(userID, content)

	var result []database.PublicSummariesSelect
	_, err := r.db.From("summaries").
		Insert(insert, false, "", "", "").
		ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no summary returned after insert")
	}

	return &result[0], nil
}

// GetLatestSummary retrieves the most recent summary for a user
func (r *Repository) GetLatestSummary(ctx context.Context, userID string) (*database.PublicSummariesSelect, error) {
	var summaries []database.PublicSummariesSelect
	_, err := r.db.From("summaries").
		Select("*", "", false).
		Eq("user_id", userID).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Limit(1, "").
		ExecuteTo(&summaries)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest summary: %w", err)
	}

	if len(summaries) == 0 {
		return nil, nil
	}

	return &summaries[0], nil
}

// HasFeeds checks if a user has at least one feed
func (r *Repository) HasFeeds(ctx context.Context, userID string) (bool, error) {
	var feeds []database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Select("id", "", false).
		Eq("user_id", userID).
		Limit(1, "").
		ExecuteTo(&feeds)

	if err != nil {
		return false, fmt.Errorf("failed to check user feeds: %w", err)
	}

	return len(feeds) > 0, nil
}
