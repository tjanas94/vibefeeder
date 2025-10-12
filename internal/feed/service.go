package feed

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

var (
	// ErrFeedAlreadyExists indicates that a feed with the same URL already exists for the user
	ErrFeedAlreadyExists = errors.New("feed already exists for this user")
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

// CreateFeed creates a new feed for the authenticated user
func (s *Service) CreateFeed(ctx context.Context, cmd models.CreateFeedCommand, userID string) error {
	// Convert command to insert model
	feedInsert := cmd.ToInsert(userID)

	// Insert feed into database
	if err := s.repo.InsertFeed(ctx, feedInsert); err != nil {
		// Check if error is due to unique constraint violation (duplicate URL for user)
		if isUniqueViolationError(err) {
			return ErrFeedAlreadyExists
		}
		return fmt.Errorf("failed to create feed: %w", err)
	}

	// Log feed_added event
	event := database.PublicEventsInsert{
		EventType: "feed_added",
		UserId:    &userID,
		Metadata: map[string]interface{}{
			"feed_name": cmd.Name,
			"feed_url":  cmd.URL,
		},
	}

	// Insert event (non-critical operation, just log if it fails)
	if err := s.repo.InsertEvent(ctx, event); err != nil {
		// Event logging failure should not prevent feed creation
		// This is logged but not returned as error
		return fmt.Errorf("feed created but event logging failed: %w", err)
	}

	return nil
}

// isUniqueViolationError checks if the error is due to unique constraint violation
// PostgREST returns 409 status code wrapped in error message for unique violations
func isUniqueViolationError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// PostgREST returns errors with "409" or "duplicate key" in the message
	return strings.Contains(errMsg, "409") ||
		strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint")
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
