package fetcher

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// MockFetcherRepository implements FetcherRepository for testing
type MockFetcherRepository struct {
	UpdateFeedAfterFetchFunc func(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error
	InsertArticlesFunc       func(ctx context.Context, articles []database.PublicArticlesInsert) error

	// Tracking for assertions
	UpdateFeedCalls   []UpdateFeedCall
	InsertArticleCall *InsertArticleCall
}

type UpdateFeedCall struct {
	FeedID string
	Update database.PublicFeedsUpdate
}

type InsertArticleCall struct {
	Articles []database.PublicArticlesInsert
}

func (m *MockFetcherRepository) FindFeedsDueForFetch(ctx context.Context, limit int) ([]database.PublicFeedsSelect, error) {
	return nil, nil
}

func (m *MockFetcherRepository) FindFeedByID(ctx context.Context, feedID string) (*database.PublicFeedsSelect, error) {
	return nil, nil
}

func (m *MockFetcherRepository) UpdateFeedAfterFetch(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error {
	m.UpdateFeedCalls = append(m.UpdateFeedCalls, UpdateFeedCall{FeedID: feedID, Update: update})
	if m.UpdateFeedAfterFetchFunc != nil {
		return m.UpdateFeedAfterFetchFunc(ctx, feedID, update)
	}
	return nil
}

func (m *MockFetcherRepository) InsertArticles(ctx context.Context, articles []database.PublicArticlesInsert) error {
	m.InsertArticleCall = &InsertArticleCall{Articles: articles}
	if m.InsertArticlesFunc != nil {
		return m.InsertArticlesFunc(ctx, articles)
	}
	return nil
}

// TestNewFeedStatusManager tests FeedStatusManager constructor
func TestNewFeedStatusManager(t *testing.T) {
	tests := []struct {
		name        string
		repo        FetcherRepository
		logger      *slog.Logger
		description string
	}{
		{
			name:        "creates manager with provided logger",
			repo:        &MockFetcherRepository{},
			logger:      slog.Default(),
			description: "Should create FeedStatusManager with provided logger",
		},
		{
			name:        "uses default logger when nil",
			repo:        &MockFetcherRepository{},
			logger:      nil,
			description: "Should use slog.Default() when logger is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsm := NewFeedStatusManager(tt.repo, tt.logger)

			assert.NotNil(t, fsm, tt.description)
			assert.NotNil(t, fsm.logger, "logger should never be nil")
			assert.Equal(t, tt.repo, fsm.repo)
		})
	}
}

// TestApplyDecisionSuccess tests successful feed fetch with article save
func TestApplyDecisionSuccess(t *testing.T) {
	tests := []struct {
		name        string
		articles    []Article
		description string
	}{
		{
			name: "success with articles",
			articles: []Article{
				{
					Title:       "Article 1",
					URL:         "https://example.com/article1",
					Content:     strPtr("Content 1"),
					PublishedAt: time.Now().UTC(),
				},
				{
					Title:       "Article 2",
					URL:         "https://example.com/article2",
					Content:     strPtr("Content 2"),
					PublishedAt: time.Now().UTC(),
				},
			},
			description: "Should save articles and update feed status to success",
		},
		{
			name:        "success without articles",
			articles:    []Article{},
			description: "Should update feed status to success without saving articles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFetcherRepository{}
			fsm := NewFeedStatusManager(mockRepo, slog.Default())

			feed := database.PublicFeedsSelect{
				Id:         "feed-1",
				Url:        "https://example.com/feed.xml",
				RetryCount: 3,
			}

			decision := FetchDecision{
				Status:        "success",
				Articles:      tt.articles,
				NextFetchTime: time.Now().UTC().Add(15 * time.Minute),
			}

			err := fsm.ApplyDecision(context.Background(), feed, decision)

			assert.NoError(t, err, tt.description)
			assert.Len(t, mockRepo.UpdateFeedCalls, 1, "Should call UpdateFeedAfterFetch once")

			updateCall := mockRepo.UpdateFeedCalls[0]
			assert.Equal(t, feed.Id, updateCall.FeedID)
			assert.Equal(t, "success", *updateCall.Update.LastFetchStatus)
			assert.Nil(t, updateCall.Update.LastFetchError)

			// Verify retry count reset to 0
			require.NotNil(t, updateCall.Update.RetryCount)
			assert.Equal(t, 0, *updateCall.Update.RetryCount)

			// Verify last_fetched_at is set
			assert.NotNil(t, updateCall.Update.LastFetchedAt)

			// Verify articles were saved if present
			if len(tt.articles) > 0 {
				assert.NotNil(t, mockRepo.InsertArticleCall)
				assert.Len(t, mockRepo.InsertArticleCall.Articles, len(tt.articles))
			} else {
				assert.Nil(t, mockRepo.InsertArticleCall, "Should not call InsertArticles when no articles")
			}
		})
	}
}

// TestApplyDecisionTemporaryError tests temporary error handling with retry increment
func TestApplyDecisionTemporaryError(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:         "feed-1",
		Url:        "https://example.com/feed.xml",
		RetryCount: 2,
	}

	errorMsg := "connection timeout"
	decision := FetchDecision{
		Status:        "temporary_error",
		ErrorMessage:  &errorMsg,
		NextFetchTime: time.Now().UTC().Add(30 * time.Minute),
		Articles:      []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	assert.Len(t, mockRepo.UpdateFeedCalls, 1)

	updateCall := mockRepo.UpdateFeedCalls[0]
	assert.Equal(t, "temporary_error", *updateCall.Update.LastFetchStatus)
	assert.Equal(t, &errorMsg, updateCall.Update.LastFetchError)

	// Verify retry count incremented
	require.NotNil(t, updateCall.Update.RetryCount)
	assert.Equal(t, 3, *updateCall.Update.RetryCount, "retry count should be incremented from 2 to 3")
}

// TestApplyDecisionPermanentError tests permanent error handling without retry increment
func TestApplyDecisionPermanentError(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:         "feed-1",
		Url:        "https://example.com/feed.xml",
		RetryCount: 5,
	}

	errorMsg := "Invalid feed URL"
	decision := FetchDecision{
		Status:       "permanent_error",
		ErrorMessage: &errorMsg,
		Articles:     []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	assert.Len(t, mockRepo.UpdateFeedCalls, 1)

	updateCall := mockRepo.UpdateFeedCalls[0]
	assert.Equal(t, "permanent_error", *updateCall.Update.LastFetchStatus)
	assert.Equal(t, &errorMsg, updateCall.Update.LastFetchError)

	// Verify retry count is NOT incremented for permanent errors
	assert.Nil(t, updateCall.Update.RetryCount, "retry count should not be set for permanent errors")
}

// TestApplyDecisionConditionalHeaders tests ETag and Last-Modified header preservation
func TestApplyDecisionConditionalHeaders(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	etag := "\"abc123\""
	lastModified := "Mon, 15 Jan 2024 10:00:00 GMT"
	decision := FetchDecision{
		Status:       "success",
		ETag:         &etag,
		LastModified: &lastModified,
		Articles:     []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	updateCall := mockRepo.UpdateFeedCalls[0]
	assert.Equal(t, &etag, updateCall.Update.Etag)
	assert.Equal(t, &lastModified, updateCall.Update.LastModified)
}

// TestApplyDecisionPermanentRedirect tests permanent redirect URL tracking
func TestApplyDecisionPermanentRedirect(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/old-feed.xml",
	}

	newURL := "https://example.com/new-feed.xml"
	decision := FetchDecision{
		Status:   "success",
		NewURL:   &newURL,
		Articles: []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	updateCall := mockRepo.UpdateFeedCalls[0]
	assert.Equal(t, &newURL, updateCall.Update.Url, "should update feed URL for permanent redirect")
}

// TestApplyDecisionNextFetchTime tests fetch_after scheduling
func TestApplyDecisionNextFetchTime(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	nextFetch := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	decision := FetchDecision{
		Status:        "success",
		NextFetchTime: nextFetch,
		Articles:      []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	updateCall := mockRepo.UpdateFeedCalls[0]
	require.NotNil(t, updateCall.Update.FetchAfter)
	assert.Equal(t, nextFetch.Format(time.RFC3339), *updateCall.Update.FetchAfter)
}

// TestApplyDecisionZeroNextFetchTime tests that zero timestamp doesn't set fetch_after
func TestApplyDecisionZeroNextFetchTime(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	decision := FetchDecision{
		Status:        "success",
		NextFetchTime: time.Time{}, // Zero timestamp
		Articles:      []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	updateCall := mockRepo.UpdateFeedCalls[0]
	assert.Nil(t, updateCall.Update.FetchAfter, "should not set fetch_after for zero timestamp")
}

// TestApplyDecisionArticleInsertionFailureNonBlocking tests that article save failure doesn't block status update
func TestApplyDecisionArticleInsertionFailureNonBlocking(t *testing.T) {
	mockRepo := &MockFetcherRepository{
		InsertArticlesFunc: func(ctx context.Context, articles []database.PublicArticlesInsert) error {
			return errors.New("database insertion failed")
		},
	}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	articles := []Article{
		{
			Title:       "Article 1",
			URL:         "https://example.com/article1",
			PublishedAt: time.Now().UTC(),
		},
	}

	decision := FetchDecision{
		Status:   "success",
		Articles: articles,
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	// ApplyDecision should NOT return error even though InsertArticles failed
	assert.NoError(t, err, "article insertion failure should be non-blocking")

	// But UpdateFeedAfterFetch should still be called
	assert.Len(t, mockRepo.UpdateFeedCalls, 1, "should still update feed status despite article insertion error")
}

// TestApplyDecisionUpdateStatusFailure tests that UpdateFeedAfterFetch errors are propagated
func TestApplyDecisionUpdateStatusFailure(t *testing.T) {
	mockRepo := &MockFetcherRepository{
		UpdateFeedAfterFetchFunc: func(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error {
			return errors.New("failed to update feed in database")
		},
	}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	decision := FetchDecision{
		Status:   "success",
		Articles: []Article{},
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.Error(t, err, "should propagate UpdateFeedAfterFetch error")
	assert.Contains(t, err.Error(), "failed to update feed status")
}

// TestApplyDecisionAllFields tests ApplyDecision with all fields populated
func TestApplyDecisionAllFields(t *testing.T) {
	mockRepo := &MockFetcherRepository{}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:         "feed-1",
		Url:        "https://example.com/old-feed.xml",
		RetryCount: 1,
	}

	etag := "\"updated-etag\""
	lastModified := "Tue, 16 Jan 2024 12:00:00 GMT"
	newURL := "https://example.com/new-feed.xml"
	errorMsg := "some error"
	nextFetch := time.Date(2024, 1, 17, 14, 0, 0, 0, time.UTC)

	articles := []Article{
		{
			Title:       "Article 1",
			URL:         "https://example.com/article1",
			Content:     strPtr("Content"),
			PublishedAt: time.Now().UTC(),
		},
	}

	decision := FetchDecision{
		Status:        "temporary_error",
		ErrorMessage:  &errorMsg,
		ETag:          &etag,
		LastModified:  &lastModified,
		NewURL:        &newURL,
		NextFetchTime: nextFetch,
		Articles:      articles,
	}

	err := fsm.ApplyDecision(context.Background(), feed, decision)

	assert.NoError(t, err)
	assert.Len(t, mockRepo.UpdateFeedCalls, 1)

	updateCall := mockRepo.UpdateFeedCalls[0]
	assert.Equal(t, "temporary_error", *updateCall.Update.LastFetchStatus)
	assert.Equal(t, &errorMsg, updateCall.Update.LastFetchError)
	assert.Equal(t, &etag, updateCall.Update.Etag)
	assert.Equal(t, &lastModified, updateCall.Update.LastModified)
	assert.Equal(t, &newURL, updateCall.Update.Url)
	assert.NotNil(t, updateCall.Update.FetchAfter)
	assert.Equal(t, 2, *updateCall.Update.RetryCount)

	// Verify articles were saved
	assert.NotNil(t, mockRepo.InsertArticleCall)
	assert.Len(t, mockRepo.InsertArticleCall.Articles, 1)
}

// TestSaveArticles tests article transformation and insertion
func TestSaveArticles(t *testing.T) {
	tests := []struct {
		name        string
		articles    []Article
		validate    func(t *testing.T, dbArticles []database.PublicArticlesInsert)
		description string
	}{
		{
			name: "transforms articles correctly",
			articles: []Article{
				{
					Title:       "Test Article",
					URL:         "https://example.com/article",
					Content:     strPtr("Article content"),
					PublishedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				},
			},
			validate: func(t *testing.T, dbArticles []database.PublicArticlesInsert) {
				assert.Len(t, dbArticles, 1)
				assert.Equal(t, "feed-1", dbArticles[0].FeedId)
				assert.Equal(t, "Test Article", dbArticles[0].Title)
				assert.Equal(t, "https://example.com/article", dbArticles[0].Url)
				// Content is *string in database.PublicArticlesInsert
				require.NotNil(t, dbArticles[0].Content)
				assert.Equal(t, "Article content", *dbArticles[0].Content, "Content should match article content")
				assert.Equal(t, "2024-01-15T10:00:00Z", dbArticles[0].PublishedAt)
			},
			description: "Should transform article to database format",
		},
		{
			name: "handles nil content",
			articles: []Article{
				{
					Title:       "Article without content",
					URL:         "https://example.com/article",
					Content:     nil,
					PublishedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				},
			},
			validate: func(t *testing.T, dbArticles []database.PublicArticlesInsert) {
				assert.Len(t, dbArticles, 1)
				// When content is nil, it should be nil in the database insert struct
				assert.Nil(t, dbArticles[0].Content, "nil content should remain nil")
			},
			description: "Should handle nil content as nil",
		},
		{
			name:     "handles empty list",
			articles: []Article{},
			validate: func(t *testing.T, dbArticles []database.PublicArticlesInsert) {
				// Should not be called for empty articles
			},
			description: "Should not call InsertArticles for empty list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFetcherRepository{}
			fsm := NewFeedStatusManager(mockRepo, slog.Default())

			err := fsm.saveArticles(context.Background(), "feed-1", tt.articles)

			if len(tt.articles) == 0 {
				assert.NoError(t, err)
				assert.Nil(t, mockRepo.InsertArticleCall, "should not call InsertArticles for empty list")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, mockRepo.InsertArticleCall)
				tt.validate(t, mockRepo.InsertArticleCall.Articles)
			}
		})
	}
}

// TestSaveArticlesInsertionError tests error handling in saveArticles
func TestSaveArticlesInsertionError(t *testing.T) {
	mockRepo := &MockFetcherRepository{
		InsertArticlesFunc: func(ctx context.Context, articles []database.PublicArticlesInsert) error {
			return errors.New("insertion failed")
		},
	}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	articles := []Article{
		{
			Title:       "Article",
			URL:         "https://example.com/article",
			PublishedAt: time.Now().UTC(),
		},
	}

	err := fsm.saveArticles(context.Background(), "feed-1", articles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert articles")
}

// TestApplyDecisionContextCancellation tests handling of cancelled context
func TestApplyDecisionContextCancellation(t *testing.T) {
	mockRepo := &MockFetcherRepository{
		UpdateFeedAfterFetchFunc: func(ctx context.Context, feedID string, update database.PublicFeedsUpdate) error {
			// Simulate context-aware operation
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		},
	}
	fsm := NewFeedStatusManager(mockRepo, slog.Default())

	feed := database.PublicFeedsSelect{
		Id:  "feed-1",
		Url: "https://example.com/feed.xml",
	}

	decision := FetchDecision{
		Status:   "success",
		Articles: []Article{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context

	err := fsm.ApplyDecision(ctx, feed, decision)

	// Should return error (context.Canceled will be wrapped)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
