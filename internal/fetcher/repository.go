package fetcher

import (
	"context"
	"fmt"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// Repository handles data access for feed fetching operations
type Repository struct {
	db *database.Client
}

// NewRepository creates a new fetcher repository
func NewRepository(db *database.Client) *Repository {
	return &Repository{db: db}
}

// FindFeedsDueForFetch retrieves feeds that are ready to be fetched
// Returns feeds where fetch_after is NULL or in the past, ordered by last_fetched_at
func (r *Repository) FindFeedsDueForFetch(ctx context.Context, limit int) ([]database.PublicFeedsSelect, error) {
	var feeds []database.PublicFeedsSelect

	// Query for feeds where:
	// - fetch_after IS NULL (never fetched or no schedule set)
	// - fetch_after <= NOW() (scheduled time has passed)
	// Order by last_fetched_at ASC NULLS FIRST to prioritize never-fetched feeds
	_, err := r.db.From("feeds").
		Select("*", "", false).
		Or("fetch_after.is.null,fetch_after.lte.now()", "").
		Order("last_fetched_at", nil). // NULLS FIRST is default for ascending order
		Limit(limit, "").
		ExecuteTo(&feeds)

	if err != nil {
		return nil, fmt.Errorf("failed to find feeds due for fetch: %w", err)
	}

	return feeds, nil
}

// FindFeedByID retrieves a single feed by ID (for immediate fetch)
func (r *Repository) FindFeedByID(ctx context.Context, feedID string) (*database.PublicFeedsSelect, error) {
	var feed database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Select("*", "", false).
		Eq("id", feedID).
		Single().
		ExecuteTo(&feed)

	if err != nil {
		return nil, fmt.Errorf("failed to find feed by id: %w", err)
	}

	return &feed, nil
}

// UpdateFeedAfterFetch updates feed status after a fetch attempt
func (r *Repository) UpdateFeedAfterFetch(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error {
	var result []database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Update(update, "", "").
		Eq("id", feedID).
		ExecuteTo(&result)

	if err != nil {
		return fmt.Errorf("failed to update feed after fetch: %w", err)
	}

	if len(result) == 0 {
		return fmt.Errorf("feed not found")
	}

	return nil
}

// InsertArticles inserts multiple articles into the database
// Uses upsert with ON CONFLICT to skip duplicates based on (feed_id, url) unique constraint
// Setting upsert=true with onConflict tells PostgREST to use ON CONFLICT DO NOTHING behavior
func (r *Repository) InsertArticles(ctx context.Context, articles []database.PublicArticlesInsert) error {
	if len(articles) == 0 {
		return nil
	}

	var result []database.PublicArticlesSelect
	_, err := r.db.From("articles").
		Insert(articles, true, "feed_id,url", "", "").
		ExecuteTo(&result)

	if err != nil {
		return fmt.Errorf("failed to insert articles: %w", err)
	}

	return nil
}
