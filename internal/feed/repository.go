package feed

import (
	"context"
	"fmt"

	"github.com/supabase-community/postgrest-go"
	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// Repository handles data access for feeds
type Repository struct {
	db *database.Client
}

// NewRepository creates a new feed repository
func NewRepository(db *database.Client) *Repository {
	return &Repository{db: db}
}

// ListFeedsResult contains the feeds and total count for pagination
type ListFeedsResult struct {
	Feeds      []database.PublicFeedsSelect
	TotalCount int
}

// ListFeeds retrieves feeds with filtering and pagination
func (r *Repository) ListFeeds(ctx context.Context, query models.ListFeedsQuery) (*ListFeedsResult, error) {
	// Calculate offset for pagination
	offset := (query.Page - 1) * pageSize

	// Build query with exact count
	feedQuery := r.db.From("feeds").Select("*", "exact", false)

	// Apply user_id filter (required for security)
	feedQuery = feedQuery.Eq("user_id", query.UserID)

	// Apply search filter if provided
	if query.Search != "" {
		// Use ILIKE for case-insensitive search
		feedQuery = feedQuery.Ilike("name", fmt.Sprintf("%%%s%%", query.Search))
	}

	// Apply status filter
	if statusFilter, hasFilter := query.GetStatusFilter(); hasFilter {
		switch statusFilter.FilterType {
		case "IN":
			feedQuery = feedQuery.In(statusFilter.Column, statusFilter.Values)
		case "IS_NULL":
			feedQuery = feedQuery.Is(statusFilter.Column, "null")
		}
	}

	// Execute query with pagination and ordering
	// The count is returned in Content-Range header by PostgREST when using "exact"
	var feeds []database.PublicFeedsSelect
	count, err := feedQuery.
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Range(offset, offset+pageSize-1, "").
		ExecuteTo(&feeds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feeds: %w", err)
	}

	return &ListFeedsResult{
		Feeds:      feeds,
		TotalCount: int(count),
	}, nil
}

// InsertFeed creates a new feed in the database and returns the created feed ID
func (r *Repository) InsertFeed(ctx context.Context, feed database.PublicFeedsInsert) (string, error) {
	var result []database.PublicFeedsSelect
	_, err := r.db.From("feeds").Insert(feed, false, "", "", "").ExecuteTo(&result)
	if err != nil {
		return "", fmt.Errorf("failed to insert feed: %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("no feed returned after insert")
	}

	return result[0].Id, nil
}

// InsertEvent creates a new event in the database
func (r *Repository) InsertEvent(ctx context.Context, event database.PublicEventsInsert) error {
	var result []database.PublicEventsSelect
	_, err := r.db.From("events").Insert(event, false, "", "", "").ExecuteTo(&result)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

// FindFeedByIDAndUser retrieves a single feed by ID and user ID
// Returns error if feed is not found or doesn't belong to the user
// Used for: edit form display, update operations
func (r *Repository) FindFeedByIDAndUser(ctx context.Context, feedID, userID string) (*database.PublicFeedsSelect, error) {
	var feed database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Select("*", "", false).
		Eq("id", feedID).
		Eq("user_id", userID).
		Single().
		ExecuteTo(&feed)

	if err != nil {
		return nil, fmt.Errorf("failed to find feed: %w", err)
	}

	return &feed, nil
}

// IsURLTaken checks if a URL is already in use by another feed for the same user
// excludeFeedID is used to exclude the current feed being updated from the check
func (r *Repository) IsURLTaken(ctx context.Context, userID, url, excludeFeedID string) (bool, error) {
	var feeds []database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Select("id", "", false).
		Eq("user_id", userID).
		Eq("url", url).
		Neq("id", excludeFeedID).
		ExecuteTo(&feeds)

	if err != nil {
		return false, fmt.Errorf("failed to check if URL is taken: %w", err)
	}

	return len(feeds) > 0, nil
}

// UpdateFeed updates an existing feed in the database
func (r *Repository) UpdateFeed(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error {
	var result []database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Update(update, "", "").
		Eq("id", feedID).
		ExecuteTo(&result)

	if err != nil {
		return fmt.Errorf("failed to update feed: %w", err)
	}

	if len(result) == 0 {
		return fmt.Errorf("feed not found")
	}

	return nil
}

// DeleteFeed deletes a feed from the database by ID and user ID
// Returns error that can be checked with database.IsNotFoundError if feed doesn't exist or doesn't belong to user
func (r *Repository) DeleteFeed(ctx context.Context, id, userID string) error {
	var result []database.PublicFeedsSelect
	_, err := r.db.From("feeds").
		Delete("", "").
		Eq("id", id).
		Eq("user_id", userID).
		ExecuteTo(&result)

	if err != nil {
		return fmt.Errorf("failed to delete feed: %w", err)
	}

	// Check if any rows were affected
	if len(result) == 0 {
		return fmt.Errorf("feed not found")
	}

	return nil
}
