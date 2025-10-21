package feed

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/feed/models"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
)

// MockFeedRepository is a mock implementation of FeedRepository
type MockFeedRepository struct {
	mock.Mock
}

func (m *MockFeedRepository) ListFeeds(ctx context.Context, query models.ListFeedsQuery) (*ListFeedsResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ListFeedsResult), args.Error(1)
}

func (m *MockFeedRepository) InsertFeed(ctx context.Context, feed database.PublicFeedsInsert) (string, error) {
	args := m.Called(ctx, feed)
	return args.String(0), args.Error(1)
}

func (m *MockFeedRepository) FindFeedByIDAndUser(ctx context.Context, feedID, userID string) (*database.PublicFeedsSelect, error) {
	args := m.Called(ctx, feedID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.PublicFeedsSelect), args.Error(1)
}

func (m *MockFeedRepository) IsURLTaken(ctx context.Context, userID, url, excludeFeedID string) (bool, error) {
	args := m.Called(ctx, userID, url, excludeFeedID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFeedRepository) UpdateFeed(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error {
	args := m.Called(ctx, feedID, update)
	return args.Error(0)
}

func (m *MockFeedRepository) DeleteFeed(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// MockEventRepository is a mock implementation of events.EventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) RecordEvent(ctx context.Context, event database.PublicEventsInsert) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Helper functions for test data
func newTestFeed(id, userID, name, url string) *database.PublicFeedsSelect {
	return &database.PublicFeedsSelect{
		Id:     id,
		UserId: userID,
		Name:   name,
		Url:    url,
	}
}

func newTestLogger() *slog.Logger {
	// Use io.Discard to suppress log output during tests
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Tests for ListFeeds
func TestListFeeds_Success(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "",
		Status: "all",
		Page:   1,
	}

	feeds := []database.PublicFeedsSelect{
		*newTestFeed("feed-1", "user-123", "Test Feed 1", "https://example.com/feed1"),
		*newTestFeed("feed-2", "user-123", "Test Feed 2", "https://example.com/feed2"),
	}

	expectedResult := &ListFeedsResult{
		Feeds:      feeds,
		TotalCount: 2,
	}

	mockRepo.On("ListFeeds", ctx, query).Return(expectedResult, nil)

	result, err := service.ListFeeds(ctx, query)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Feeds, 2)
	assert.False(t, result.ShowEmptyState)
	mockRepo.AssertExpectations(t)
}

func TestListFeeds_Empty(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "",
		Status: "all",
		Page:   1,
	}

	expectedResult := &ListFeedsResult{
		Feeds:      []database.PublicFeedsSelect{},
		TotalCount: 0,
	}

	mockRepo.On("ListFeeds", ctx, query).Return(expectedResult, nil)

	result, err := service.ListFeeds(ctx, query)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Feeds, 0)
	assert.True(t, result.ShowEmptyState)
	mockRepo.AssertExpectations(t)
}

func TestListFeeds_WithSearchFilter(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "tech",
		Status: "all",
		Page:   1,
	}

	expectedResult := &ListFeedsResult{
		Feeds:      []database.PublicFeedsSelect{},
		TotalCount: 0,
	}

	mockRepo.On("ListFeeds", ctx, query).Return(expectedResult, nil)

	result, err := service.ListFeeds(ctx, query)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.ShowEmptyState) // Don't show empty state when filters are applied
	mockRepo.AssertExpectations(t)
}

func TestListFeeds_RepositoryError(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	query := models.ListFeedsQuery{
		UserID: "user-123",
	}

	mockRepo.On("ListFeeds", ctx, query).Return(nil, errors.New("database error"))

	result, err := service.ListFeeds(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list feeds")
	mockRepo.AssertExpectations(t)
}

// Tests for CreateFeed
func TestCreateFeed_Success(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	userID := "user-123"
	cmd := models.CreateFeedCommand{
		Name: "Tech News",
		URL:  "https://example.com/feed",
	}

	mockRepo.On("InsertFeed", ctx, mock.MatchedBy(func(feed database.PublicFeedsInsert) bool {
		return feed.Name == cmd.Name && feed.Url == cmd.URL && feed.UserId == userID
	})).Return("feed-123", nil)

	mockEventRepo.On("RecordEvent", ctx, mock.MatchedBy(func(event database.PublicEventsInsert) bool {
		return event.EventType == events.EventFeedAdded && event.UserId != nil && *event.UserId == userID
	})).Return(nil)

	feedID, err := service.CreateFeed(ctx, cmd, userID)

	require.NoError(t, err)
	assert.Equal(t, "feed-123", feedID)
	mockRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestCreateFeed_URLAlreadyExists(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	userID := "user-123"
	cmd := models.CreateFeedCommand{
		Name: "Tech News",
		URL:  "https://example.com/feed",
	}

	// Simulate unique constraint violation error from database
	mockRepo.On("InsertFeed", ctx, mock.AnythingOfType("database.PublicFeedsInsert")).
		Return("", errors.New("unique constraint violation"))

	feedID, err := service.CreateFeed(ctx, cmd, userID)

	assert.Error(t, err)
	assert.Equal(t, "", feedID)
	assert.Equal(t, ErrFeedAlreadyExists, err)
	mockEventRepo.AssertNotCalled(t, "RecordEvent")
}

func TestCreateFeed_RepositoryError(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	userID := "user-123"
	cmd := models.CreateFeedCommand{
		Name: "Tech News",
		URL:  "https://example.com/feed",
	}

	mockRepo.On("InsertFeed", ctx, mock.AnythingOfType("database.PublicFeedsInsert")).
		Return("", errors.New("unexpected error"))

	feedID, err := service.CreateFeed(ctx, cmd, userID)

	assert.Error(t, err)
	assert.Equal(t, "", feedID)
	assert.Contains(t, err.Error(), "failed to create feed")
}

func TestCreateFeed_EventLogError(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	userID := "user-123"
	cmd := models.CreateFeedCommand{
		Name: "Tech News",
		URL:  "https://example.com/feed",
	}

	mockRepo.On("InsertFeed", ctx, mock.AnythingOfType("database.PublicFeedsInsert")).
		Return("feed-123", nil)

	mockEventRepo.On("RecordEvent", ctx, mock.AnythingOfType("database.PublicEventsInsert")).
		Return(errors.New("event log failed"))

	// Should not fail if event logging fails, only warn
	feedID, err := service.CreateFeed(ctx, cmd, userID)

	require.NoError(t, err)
	assert.Equal(t, "feed-123", feedID)
	mockRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// Tests for GetFeedForEdit
func TestGetFeedForEdit_Success(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	dbFeed := newTestFeed(feedID, userID, "Test Feed", "https://example.com/feed")
	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).Return(dbFeed, nil)

	result, err := service.GetFeedForEdit(ctx, feedID, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "edit", result.Mode)
	assert.Equal(t, dbFeed.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetFeedForEdit_NotFound(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	// Simulate not found error from database
	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).
		Return(nil, errors.New("no rows"))

	result, err := service.GetFeedForEdit(ctx, feedID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrFeedNotFound, err)
}

func TestGetFeedForEdit_RepositoryError(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).
		Return(nil, errors.New("database error"))

	result, err := service.GetFeedForEdit(ctx, feedID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get feed for edit")
}

// Tests for UpdateFeed
func TestUpdateFeed_NameOnlyUpdate(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"
	oldURL := "https://example.com/feed"

	cmd := models.UpdateFeedCommand{
		Name: "Updated Feed Name",
		URL:  oldURL, // Same URL
	}

	existingFeed := &database.PublicFeedsSelect{
		Id:     feedID,
		UserId: userID,
		Name:   "Old Feed Name",
		Url:    oldURL,
	}

	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).Return(existingFeed, nil)
	mockRepo.On("UpdateFeed", ctx, feedID, mock.AnythingOfType("database.PublicFeedsUpdate")).Return(nil)

	urlChanged, err := service.UpdateFeed(ctx, feedID, userID, cmd)

	require.NoError(t, err)
	assert.False(t, urlChanged)
	mockRepo.AssertExpectations(t)
}

func TestUpdateFeed_URLChange(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"
	oldURL := "https://example.com/feed"
	newURL := "https://newexample.com/feed"

	cmd := models.UpdateFeedCommand{
		Name: "Updated Feed Name",
		URL:  newURL,
	}

	existingFeed := &database.PublicFeedsSelect{
		Id:     feedID,
		UserId: userID,
		Name:   "Old Feed Name",
		Url:    oldURL,
	}

	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).Return(existingFeed, nil)
	mockRepo.On("IsURLTaken", ctx, userID, newURL, feedID).Return(false, nil)
	mockRepo.On("UpdateFeed", ctx, feedID, mock.AnythingOfType("database.PublicFeedsUpdate")).Return(nil)

	urlChanged, err := service.UpdateFeed(ctx, feedID, userID, cmd)

	require.NoError(t, err)
	assert.True(t, urlChanged)
	mockRepo.AssertExpectations(t)
}

func TestUpdateFeed_URLConflict(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"
	oldURL := "https://example.com/feed"
	newURL := "https://newexample.com/feed"

	cmd := models.UpdateFeedCommand{
		Name: "Updated Feed Name",
		URL:  newURL,
	}

	existingFeed := &database.PublicFeedsSelect{
		Id:     feedID,
		UserId: userID,
		Name:   "Old Feed Name",
		Url:    oldURL,
	}

	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).Return(existingFeed, nil)
	mockRepo.On("IsURLTaken", ctx, userID, newURL, feedID).Return(true, nil)

	urlChanged, err := service.UpdateFeed(ctx, feedID, userID, cmd)

	assert.Error(t, err)
	assert.False(t, urlChanged)
	assert.Equal(t, ErrFeedURLConflict, err)
	mockRepo.AssertNotCalled(t, "UpdateFeed")
}

func TestUpdateFeed_FeedNotFound(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	cmd := models.UpdateFeedCommand{
		Name: "Updated Feed Name",
		URL:  "https://example.com/feed",
	}

	// Simulate not found error from database
	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).
		Return(nil, errors.New("not found"))

	urlChanged, err := service.UpdateFeed(ctx, feedID, userID, cmd)

	assert.Error(t, err)
	assert.False(t, urlChanged)
	assert.Equal(t, ErrFeedNotFound, err)
}

func TestUpdateFeed_UpdateError(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"
	oldURL := "https://example.com/feed"

	cmd := models.UpdateFeedCommand{
		Name: "Updated Feed Name",
		URL:  oldURL,
	}

	existingFeed := &database.PublicFeedsSelect{
		Id:     feedID,
		UserId: userID,
		Name:   "Old Feed Name",
		Url:    oldURL,
	}

	mockRepo.On("FindFeedByIDAndUser", ctx, feedID, userID).Return(existingFeed, nil)
	mockRepo.On("UpdateFeed", ctx, feedID, mock.AnythingOfType("database.PublicFeedsUpdate")).
		Return(errors.New("update failed"))

	urlChanged, err := service.UpdateFeed(ctx, feedID, userID, cmd)

	assert.Error(t, err)
	assert.False(t, urlChanged)
	assert.Contains(t, err.Error(), "failed to update feed")
}

// Tests for DeleteFeed
func TestDeleteFeed_Success(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	mockRepo.On("DeleteFeed", ctx, feedID, userID).Return(nil)

	err := service.DeleteFeed(ctx, feedID, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteFeed_NotFound(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	// Simulate not found error from database
	mockRepo.On("DeleteFeed", ctx, feedID, userID).
		Return(errors.New("404 not found"))

	err := service.DeleteFeed(ctx, feedID, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrFeedNotFound, err)
}

func TestDeleteFeed_RepositoryError(t *testing.T) {
	mockRepo := new(MockFeedRepository)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockEventRepo, logger)

	ctx := context.Background()
	feedID := "feed-123"
	userID := "user-123"

	mockRepo.On("DeleteFeed", ctx, feedID, userID).
		Return(errors.New("database error"))

	err := service.DeleteFeed(ctx, feedID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete feed")
}

// Tests for buildFeedListViewModel (pure function)
func TestBuildFeedListViewModel_WithFeeds(t *testing.T) {
	feeds := []database.PublicFeedsSelect{
		*newTestFeed("feed-1", "user-123", "Feed 1", "https://example.com/feed1"),
		*newTestFeed("feed-2", "user-123", "Feed 2", "https://example.com/feed2"),
	}

	result := &ListFeedsResult{
		Feeds:      feeds,
		TotalCount: 2,
	}

	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "",
		Status: "all",
		Page:   1,
	}

	viewModel := buildFeedListViewModel(result, query)

	assert.Len(t, viewModel.Feeds, 2)
	assert.False(t, viewModel.ShowEmptyState)
	assert.Equal(t, 1, viewModel.Pagination.CurrentPage)
}

func TestBuildFeedListViewModel_EmptyWithFilters(t *testing.T) {
	result := &ListFeedsResult{
		Feeds:      []database.PublicFeedsSelect{},
		TotalCount: 0,
	}

	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "tech",
		Status: "all",
		Page:   1,
	}

	viewModel := buildFeedListViewModel(result, query)

	assert.Len(t, viewModel.Feeds, 0)
	assert.False(t, viewModel.ShowEmptyState)
}

func TestBuildFeedListViewModel_EmptyNoFilters(t *testing.T) {
	result := &ListFeedsResult{
		Feeds:      []database.PublicFeedsSelect{},
		TotalCount: 0,
	}

	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "",
		Status: "all",
		Page:   1,
	}

	viewModel := buildFeedListViewModel(result, query)

	assert.Len(t, viewModel.Feeds, 0)
	assert.True(t, viewModel.ShowEmptyState)
}

func TestBuildFeedListViewModel_Pagination(t *testing.T) {
	// Simulate 100 total items with page size 20
	feeds := make([]database.PublicFeedsSelect, 20)
	for i := 0; i < 20; i++ {
		feeds[i] = *newTestFeed("feed-"+string(rune(i)), "user-123", "Feed", "https://example.com")
	}

	result := &ListFeedsResult{
		Feeds:      feeds,
		TotalCount: 100,
	}

	query := models.ListFeedsQuery{
		UserID: "user-123",
		Search: "",
		Status: "all",
		Page:   2,
	}

	viewModel := buildFeedListViewModel(result, query)

	assert.Equal(t, 2, viewModel.Pagination.CurrentPage)
	assert.True(t, viewModel.Pagination.HasPrevious)
	assert.True(t, viewModel.Pagination.HasNext)
}
