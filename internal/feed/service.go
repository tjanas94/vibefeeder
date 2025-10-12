package feed

import (
	"context"
	"fmt"

	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// Service handles business logic for feeds
type Service struct {
	repo *Repository
}

// NewService creates a new feed service
func NewService(db *database.Client) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

// ListFeeds retrieves and transforms feeds for display
func (s *Service) ListFeeds(ctx context.Context, query models.ListFeedsQuery) (*models.FeedListViewModel, error) {
	// Fetch feeds from repository
	result, err := s.repo.ListFeeds(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list feeds: %w", err)
	}

	// Build view model using pure function
	viewModel := buildFeedListViewModel(result, query)
	return &viewModel, nil
}

// buildFeedListViewModel is a pure function that transforms repository result to view model
func buildFeedListViewModel(result *ListFeedsResult, query models.ListFeedsQuery) models.FeedListViewModel {
	// Calculate pagination data
	totalPages := (result.TotalCount + query.Limit - 1) / query.Limit
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := models.PaginationViewModel{
		CurrentPage: query.Page,
		TotalPages:  totalPages,
		TotalItems:  result.TotalCount,
		HasPrevious: query.Page > 1,
		HasNext:     query.Page < totalPages,
	}

	// Transform database models to view models
	feedItems := make([]models.FeedItemViewModel, len(result.Feeds))
	for i, dbFeed := range result.Feeds {
		feedItems[i] = models.NewFeedItemFromDB(dbFeed)
	}

	// Determine if empty state should be shown
	// Show empty state only when there are no feeds at all (no filters applied)
	showEmptyState := result.TotalCount == 0 && query.Search == "" && query.Status == "all"

	return models.FeedListViewModel{
		Feeds:          feedItems,
		ShowEmptyState: showEmptyState,
		Pagination:     pagination,
	}
}
