package feed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
	sharedmodels "github.com/tjanas94/vibefeeder/internal/shared/models"
)

var (
	// ErrFeedAlreadyExists indicates that a feed with the same URL already exists for the user
	ErrFeedAlreadyExists = errors.New("feed already exists for this user")

	// ErrFeedNotFound indicates that the feed was not found or doesn't belong to the user
	ErrFeedNotFound = errors.New("feed not found")

	// ErrFeedURLConflict indicates that the new URL is already in use by another feed
	ErrFeedURLConflict = errors.New("feed URL already in use")
)

// FeedRepository defines the interface for feed data access
type FeedRepository interface {
	ListFeeds(ctx context.Context, query models.ListFeedsQuery) (*ListFeedsResult, error)
	InsertFeed(ctx context.Context, feed database.PublicFeedsInsert) (string, error)
	FindFeedByIDAndUser(ctx context.Context, feedID, userID string) (*database.PublicFeedsSelect, error)
	IsURLTaken(ctx context.Context, userID, url, excludeFeedID string) (bool, error)
	UpdateFeed(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error
	DeleteFeed(ctx context.Context, id, userID string) error
}

// Service handles business logic for feeds
type Service struct {
	repo      FeedRepository
	eventRepo events.EventRepository
	logger    *slog.Logger
}

// NewService creates a new feed service
func NewService(repo FeedRepository, eventRepo events.EventRepository, logger *slog.Logger) *Service {
	return &Service{
		repo:      repo,
		eventRepo: eventRepo,
		logger:    logger,
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

// CreateFeed creates a new feed for the authenticated user and returns the feed ID
func (s *Service) CreateFeed(ctx context.Context, cmd models.CreateFeedCommand, userID string) (string, error) {
	// Convert command to insert model
	feedInsert := cmd.ToInsert(userID)

	// Insert feed into database
	feedID, err := s.repo.InsertFeed(ctx, feedInsert)
	if err != nil {
		// Check if error is due to unique constraint violation (duplicate URL for user)
		if database.IsUniqueViolationError(err) {
			return "", ErrFeedAlreadyExists
		}
		return "", fmt.Errorf("failed to create feed: %w", err)
	}

	// Log feed_added event
	if err := s.eventRepo.RecordEvent(ctx, database.PublicEventsInsert{
		EventType: events.EventFeedAdded,
		UserId:    &userID,
		Metadata: map[string]any{
			"feed_name": cmd.Name,
			"feed_url":  cmd.URL,
		},
	}); err != nil {
		s.logger.Warn("Failed to log event", "event_type", events.EventFeedAdded, "error", err, "user_id", userID)
	}

	return feedID, nil
}

// GetFeedForEdit retrieves a feed for editing
func (s *Service) GetFeedForEdit(ctx context.Context, feedID, userID string) (*models.FeedFormViewModel, error) {
	// Fetch feed from repository (includes authorization check via user_id)
	dbFeed, err := s.repo.FindFeedByIDAndUser(ctx, feedID, userID)
	if err != nil {
		// Check if error indicates feed not found
		if database.IsNotFoundError(err) {
			return nil, ErrFeedNotFound
		}
		return nil, fmt.Errorf("failed to get feed for edit: %w", err)
	}

	// Map database model to view model
	vm := models.NewFeedFormForEdit(*dbFeed)
	return &vm, nil
}

// UpdateFeed updates an existing feed with validation and conflict detection
// Returns true if URL was changed, false otherwise
func (s *Service) UpdateFeed(ctx context.Context, feedID, userID string, cmd models.UpdateFeedCommand) (bool, error) {
	// Get existing feed to verify ownership
	existingFeed, err := s.repo.FindFeedByIDAndUser(ctx, feedID, userID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, ErrFeedNotFound
		}
		return false, fmt.Errorf("failed to get feed for update: %w", err)
	}

	// Prepare update data based on whether URL changed
	var updateData database.PublicFeedsUpdate
	urlChanged := existingFeed.Url != cmd.URL

	if !urlChanged {
		// Name-only update - preserves fetch-related fields
		updateData = cmd.ToUpdate()
	} else {
		// URL changed - validate and reset fetch-related fields
		if err := s.validateURLChange(ctx, userID, feedID, cmd.URL); err != nil {
			return false, err
		}
		updateData = cmd.ToUpdateWithURLChange()
	}

	// Perform update
	if err := s.repo.UpdateFeed(ctx, feedID, updateData); err != nil {
		return false, fmt.Errorf("failed to update feed: %w", err)
	}

	return urlChanged, nil
}

// validateURLChange checks if new URL can be used
func (s *Service) validateURLChange(ctx context.Context, userID, feedID, newURL string) error {
	isTaken, err := s.repo.IsURLTaken(ctx, userID, newURL, feedID)
	if err != nil {
		return fmt.Errorf("failed to check URL availability: %w", err)
	}

	if isTaken {
		return ErrFeedURLConflict
	}

	return nil
}

// DeleteFeed deletes a feed for the authenticated user
func (s *Service) DeleteFeed(ctx context.Context, id, userID string) error {
	// Delete feed from repository (includes authorization check via user_id)
	if err := s.repo.DeleteFeed(ctx, id, userID); err != nil {
		// Check if error indicates feed not found
		if database.IsNotFoundError(err) {
			return ErrFeedNotFound
		}
		// Return generic error for other cases
		return fmt.Errorf("failed to delete feed: %w", err)
	}

	return nil
}

// buildFeedListViewModel is a pure function that transforms repository result to view model
func buildFeedListViewModel(result *ListFeedsResult, query models.ListFeedsQuery) models.FeedListViewModel {
	// Transform database models to view models
	feedItems := make([]models.FeedItemViewModel, len(result.Feeds))
	for i, dbFeed := range result.Feeds {
		feedItems[i] = models.NewFeedItemFromDB(dbFeed)
	}

	// Show empty state only when there are no feeds at all (no filters applied)
	showEmptyState := result.TotalCount == 0 && query.Search == "" && query.Status == "all"

	return models.FeedListViewModel{
		Feeds:          feedItems,
		ShowEmptyState: showEmptyState,
		Pagination:     sharedmodels.BuildPagination(result.TotalCount, query.Page, pageSize),
	}
}
