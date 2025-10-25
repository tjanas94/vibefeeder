package summary

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/shared/ai"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
)

// MockSummaryRepository is a mock implementation of SummaryRepository
type MockSummaryRepository struct {
	mock.Mock
}

func (m *MockSummaryRepository) FetchRecentArticles(ctx context.Context, userID string, limit int) ([]models.ArticleForPrompt, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ArticleForPrompt), args.Error(1)
}

func (m *MockSummaryRepository) SaveSummary(ctx context.Context, userID, content string) (*database.PublicSummariesSelect, error) {
	args := m.Called(ctx, userID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.PublicSummariesSelect), args.Error(1)
}

func (m *MockSummaryRepository) GetLatestSummary(ctx context.Context, userID string) (*database.PublicSummariesSelect, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.PublicSummariesSelect), args.Error(1)
}

func (m *MockSummaryRepository) HasFeeds(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// MockAIClient is a mock implementation of AIClient
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) GenerateChatCompletion(ctx context.Context, options ai.GenerateChatCompletionOptions) (*ai.ChatCompletionResponse, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.ChatCompletionResponse), args.Error(1)
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
func newTestArticle(title, content string) models.ArticleForPrompt {
	return models.ArticleForPrompt{
		Title:   title,
		Content: &content,
	}
}

func newTestSummary(id, userID, content string) *database.PublicSummariesSelect {
	now := time.Now().Format(time.RFC3339)
	return &database.PublicSummariesSelect{
		Id:        id,
		UserId:    userID,
		Content:   content,
		CreatedAt: now,
	}
}

func newTestAIResponse(content string) *ai.ChatCompletionResponse {
	return &ai.ChatCompletionResponse{
		ID: "response-123",
		Choices: []ai.Choice{
			{
				Index: 0,
				Message: ai.ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Model: "gpt-4o-mini",
	}
}

func newTestLogger() *slog.Logger {
	// Use io.Discard to suppress log output during tests
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Tests for GenerateSummary
func TestGenerateSummary_Success(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		newTestArticle("Article 1", "Content 1"),
		newTestArticle("Article 2", "Content 2"),
	}

	summaryContent := "This is a summary of recent articles."
	dbSummary := newTestSummary("summary-123", userID, summaryContent)

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	// Note: context is a timeout context created inside GenerateSummary, not the original context
	mockAI.On("GenerateChatCompletion", mock.Anything, mock.MatchedBy(func(opts ai.GenerateChatCompletionOptions) bool {
		return opts.Model == "openai/gpt-4o-mini" && opts.Temperature == 0.7 && opts.MaxTokens == 2000
	})).Return(newTestAIResponse(summaryContent), nil)

	mockRepo.On("SaveSummary", ctx, userID, summaryContent).
		Return(dbSummary, nil)

	mockEventRepo.On("RecordEvent", ctx, mock.MatchedBy(func(event database.PublicEventsInsert) bool {
		return event.EventType == events.EventSummaryGenerated && event.UserId != nil && *event.UserId == userID
	})).Return(nil)

	result, err := service.GenerateSummary(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, summaryContent, result.Summary.Content)
	assert.True(t, result.CanGenerate) // Default value since we didn't check feeds
	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestGenerateSummary_NoArticles(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return([]models.ArticleForPrompt{}, nil)

	result, err := service.GenerateSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 404, serviceErr.Code)
	mockAI.AssertNotCalled(t, "GenerateChatCompletion")
	mockRepo.AssertNotCalled(t, "SaveSummary")
}

func TestGenerateSummary_FetchArticlesError(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(nil, errors.New("database error"))

	result, err := service.GenerateSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 500, serviceErr.Code)
	mockAI.AssertNotCalled(t, "GenerateChatCompletion")
}

func TestGenerateSummary_AIServiceError(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		newTestArticle("Article 1", "Content 1"),
	}

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(nil, errors.New("AI service error"))

	result, err := service.GenerateSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 503, serviceErr.Code)
	mockRepo.AssertNotCalled(t, "SaveSummary")
}

func TestGenerateSummary_AIResponseEmpty(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		newTestArticle("Article 1", "Content 1"),
	}

	// Return response with no choices
	emptyResponse := &ai.ChatCompletionResponse{
		ID:      "response-123",
		Choices: []ai.Choice{},
		Model:   "gpt-4o-mini",
	}

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(emptyResponse, nil)

	result, err := service.GenerateSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 503, serviceErr.Code)
}

func TestGenerateSummary_SaveSummaryError(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		newTestArticle("Article 1", "Content 1"),
	}

	summaryContent := "This is a summary"

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(newTestAIResponse(summaryContent), nil)

	mockRepo.On("SaveSummary", ctx, userID, summaryContent).
		Return(nil, errors.New("save failed"))

	result, err := service.GenerateSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 500, serviceErr.Code)
}

func TestGenerateSummary_EventLogError(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		newTestArticle("Article 1", "Content 1"),
	}

	summaryContent := "This is a summary"
	dbSummary := newTestSummary("summary-123", userID, summaryContent)

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(newTestAIResponse(summaryContent), nil)

	mockRepo.On("SaveSummary", ctx, userID, summaryContent).
		Return(dbSummary, nil)

	mockEventRepo.On("RecordEvent", ctx, mock.AnythingOfType("database.PublicEventsInsert")).
		Return(errors.New("event log failed"))

	// Should not fail if event logging fails, only warn
	result, err := service.GenerateSummary(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, summaryContent, result.Summary.Content)
}

func TestGenerateSummary_ManyArticles(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	// Create many articles (should be limited by maxArticlesForSummary)
	articles := make([]models.ArticleForPrompt, maxArticlesForSummary)
	for i := 0; i < maxArticlesForSummary; i++ {
		content := "Content " + string(rune(i))
		articles[i] = newTestArticle("Article "+string(rune(i)), content)
	}

	summaryContent := "Summary of many articles"
	dbSummary := newTestSummary("summary-123", userID, summaryContent)

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(newTestAIResponse(summaryContent), nil)

	mockRepo.On("SaveSummary", ctx, userID, summaryContent).
		Return(dbSummary, nil)

	mockEventRepo.On("RecordEvent", ctx, mock.AnythingOfType("database.PublicEventsInsert")).
		Return(nil)

	result, err := service.GenerateSummary(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, summaryContent, result.Summary.Content)
}

// Tests for GetLatestSummaryForUser
func TestGetLatestSummaryForUser_WithSummaryAndFeeds(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	dbSummary := newTestSummary("summary-123", userID, "Recent summary content")

	mockRepo.On("GetLatestSummary", ctx, userID).Return(dbSummary, nil)
	mockRepo.On("HasFeeds", ctx, userID).Return(true, nil)

	result, err := service.GetLatestSummaryForUser(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Summary)
	assert.Equal(t, "Recent summary content", result.Summary.Content)
	assert.True(t, result.CanGenerate)
	mockRepo.AssertExpectations(t)
}

func TestGetLatestSummaryForUser_NoSummaryWithFeeds(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("GetLatestSummary", ctx, userID).Return(nil, nil)
	mockRepo.On("HasFeeds", ctx, userID).Return(true, nil)

	result, err := service.GetLatestSummaryForUser(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.Summary)
	assert.True(t, result.CanGenerate)
	mockRepo.AssertExpectations(t)
}

func TestGetLatestSummaryForUser_WithSummaryNoFeeds(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	dbSummary := newTestSummary("summary-123", userID, "Old summary")

	mockRepo.On("GetLatestSummary", ctx, userID).Return(dbSummary, nil)
	mockRepo.On("HasFeeds", ctx, userID).Return(false, nil)

	result, err := service.GetLatestSummaryForUser(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Summary)
	assert.False(t, result.CanGenerate) // Can't generate without feeds
	mockRepo.AssertExpectations(t)
}

func TestGetLatestSummaryForUser_NoSummaryNoFeeds(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("GetLatestSummary", ctx, userID).Return(nil, nil)
	mockRepo.On("HasFeeds", ctx, userID).Return(false, nil)

	result, err := service.GetLatestSummaryForUser(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.Summary)
	assert.False(t, result.CanGenerate)
	mockRepo.AssertExpectations(t)
}

func TestGetLatestSummaryForUser_GetSummaryError(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("GetLatestSummary", ctx, userID).
		Return(nil, errors.New("database error"))

	result, err := service.GetLatestSummaryForUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "HasFeeds")
}

func TestGetLatestSummaryForUser_HasFeedsError(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	dbSummary := newTestSummary("summary-123", userID, "Summary")

	mockRepo.On("GetLatestSummary", ctx, userID).Return(dbSummary, nil)
	mockRepo.On("HasFeeds", ctx, userID).Return(false, errors.New("check failed"))

	result, err := service.GetLatestSummaryForUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 500, serviceErr.Code)
}

// Tests for buildSummaryDisplayViewModel (pure function)
func TestBuildSummaryDisplayViewModel_WithSummary(t *testing.T) {
	dbSummary := newTestSummary("summary-123", "user-123", "Summary content")

	viewModel := buildSummaryDisplayViewModel(dbSummary, true)

	assert.NotNil(t, viewModel.Summary)
	assert.Equal(t, "Summary content", viewModel.Summary.Content)
	assert.True(t, viewModel.CanGenerate)
}

func TestBuildSummaryDisplayViewModel_WithoutSummary(t *testing.T) {
	viewModel := buildSummaryDisplayViewModel(nil, true)

	assert.Nil(t, viewModel.Summary)
	assert.True(t, viewModel.CanGenerate)
}

func TestBuildSummaryDisplayViewModel_CannotGenerate(t *testing.T) {
	dbSummary := newTestSummary("summary-123", "user-123", "Old summary")

	viewModel := buildSummaryDisplayViewModel(dbSummary, false)

	assert.NotNil(t, viewModel.Summary)
	assert.False(t, viewModel.CanGenerate)
}

func TestBuildSummaryDisplayViewModel_NilWithoutCanGenerate(t *testing.T) {
	viewModel := buildSummaryDisplayViewModel(nil, false)

	assert.Nil(t, viewModel.Summary)
	assert.False(t, viewModel.CanGenerate)
}

// Tests for AIClient context timeout
func TestGenerateSummary_ContextTimeout(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		newTestArticle("Article 1", "Content 1"),
	}

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	// Simulate AI service timing out
	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(nil, errors.New("context deadline exceeded"))

	result, err := service.GenerateSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Check if error is a ServiceError with the correct HTTP code
	serviceErr, ok := sharederrors.AsServiceError(err)
	assert.True(t, ok, "error should be a ServiceError")
	assert.Equal(t, 503, serviceErr.Code)
}

// Tests for Articles with nil content
func TestGenerateSummary_ArticlesWithNilContent(t *testing.T) {
	mockRepo := new(MockSummaryRepository)
	mockAI := new(MockAIClient)
	mockEventRepo := new(MockEventRepository)
	logger := newTestLogger()
	service := NewService(mockRepo, mockAI, logger, mockEventRepo)

	ctx := context.Background()
	userID := "user-123"

	articles := []models.ArticleForPrompt{
		{Title: "Article 1", Content: nil},
		{Title: "Article 2", Content: nil},
	}

	summaryContent := "Summary even with nil content"
	dbSummary := newTestSummary("summary-123", userID, summaryContent)

	mockRepo.On("FetchRecentArticles", ctx, userID, maxArticlesForSummary).
		Return(articles, nil)

	mockAI.On("GenerateChatCompletion", mock.Anything, mock.AnythingOfType("ai.GenerateChatCompletionOptions")).
		Return(newTestAIResponse(summaryContent), nil)

	mockRepo.On("SaveSummary", ctx, userID, summaryContent).
		Return(dbSummary, nil)

	mockEventRepo.On("RecordEvent", ctx, mock.AnythingOfType("database.PublicEventsInsert")).
		Return(nil)

	result, err := service.GenerateSummary(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, summaryContent, result.Summary.Content)
}
