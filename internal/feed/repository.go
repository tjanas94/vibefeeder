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
	offset := (query.Page - 1) * query.Limit

	// Build query with exact count
	feedQuery := r.db.From("feeds").Select("*", "exact", false)

	// Apply user_id filter (required for security)
	feedQuery = feedQuery.Eq("user_id", query.UserID)

	// Apply search filter if provided
	if query.Search != "" {
		// Use ILIKE for case-insensitive search
		feedQuery = feedQuery.Ilike("name", fmt.Sprintf("%%%s%%", query.Search))
	}

	// Apply status filter if needed
	if statuses, hasFilter := query.GetStatusFilter(); hasFilter {
		feedQuery = feedQuery.In("last_fetch_status", statuses)
	}

	// Execute query with pagination and ordering
	// The count is returned in Content-Range header by PostgREST when using "exact"
	var feeds []database.PublicFeedsSelect
	count, err := feedQuery.
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Range(offset, offset+query.Limit-1, "").
		ExecuteTo(&feeds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feeds: %w", err)
	}

	return &ListFeedsResult{
		Feeds:      feeds,
		TotalCount: int(count),
	}, nil
}
