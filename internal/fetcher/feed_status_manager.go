package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// FeedStatusManager handles database updates after fetch decisions
type FeedStatusManager struct {
	repo   FetcherRepository
	logger *slog.Logger
}

// NewFeedStatusManager creates a new feed status manager
func NewFeedStatusManager(repo FetcherRepository, logger *slog.Logger) *FeedStatusManager {
	if logger == nil {
		logger = slog.Default()
	}

	return &FeedStatusManager{
		repo:   repo,
		logger: logger,
	}
}

// ApplyDecision applies a FetchDecision to the database
// Saves articles if present and updates feed status
// Article save failures don't block status update
func (fsm *FeedStatusManager) ApplyDecision(
	ctx context.Context,
	feed database.PublicFeedsSelect,
	decision FetchDecision,
) error {
	// Save articles if any (non-blocking failure)
	if len(decision.Articles) > 0 {
		if err := fsm.saveArticles(ctx, feed.Id, decision.Articles); err != nil {
			fsm.logger.Error("Failed to save articles", "feed_id", feed.Id, "error", err)
			// Don't return error - we still want to update status
		}
	}

	// Build update params
	update := database.PublicFeedsUpdate{
		LastFetchStatus: &decision.Status,
		LastFetchError:  decision.ErrorMessage,
	}

	// Set fetch_after for scheduling next fetch
	if !decision.NextFetchTime.IsZero() {
		fetchAfterStr := decision.NextFetchTime.UTC().Format(time.RFC3339)
		update.FetchAfter = &fetchAfterStr
	}

	// Set conditional headers if provided
	if decision.ETag != nil {
		update.Etag = decision.ETag
	}
	if decision.LastModified != nil {
		update.LastModified = decision.LastModified
	}

	// Set new URL if provided (permanent redirect)
	if decision.NewURL != nil {
		update.Url = decision.NewURL
	}

	// Reset retry count on success, increment on failure
	switch decision.Status {
	case "success":
		retryCount := 0
		update.RetryCount = &retryCount
	case "temporary_error":
		newRetryCount := feed.RetryCount + 1
		update.RetryCount = &newRetryCount
	}

	// Set last_fetched_at timestamp
	now := time.Now().UTC().Format(time.RFC3339)
	update.LastFetchedAt = &now

	// Update feed in database
	if err := fsm.repo.UpdateFeedAfterFetch(ctx, feed.Id, update); err != nil {
		return fmt.Errorf("failed to update feed status: %w", err)
	}

	return nil
}

// saveArticles saves parsed articles to database, skipping duplicates
func (fsm *FeedStatusManager) saveArticles(
	ctx context.Context,
	feedID string,
	articles []Article,
) error {
	if len(articles) == 0 {
		return nil
	}

	fsm.logger.Info("Saving articles", "feed_id", feedID, "count", len(articles))

	// Transform articles to database insert format
	dbArticles := make([]database.PublicArticlesInsert, 0, len(articles))
	for _, article := range articles {
		dbArticle := database.PublicArticlesInsert{
			FeedId:      feedID,
			Title:       article.Title,
			Url:         article.URL,
			Content:     article.Content,
			PublishedAt: article.PublishedAt.UTC().Format(time.RFC3339),
		}
		dbArticles = append(dbArticles, dbArticle)
	}

	// Insert articles (duplicates are ignored by UNIQUE constraint on feed_id, url)
	if err := fsm.repo.InsertArticles(ctx, dbArticles); err != nil {
		return fmt.Errorf("failed to insert articles: %w", err)
	}

	fsm.logger.Info("Articles saved successfully", "feed_id", feedID, "count", len(articles))
	return nil
}
