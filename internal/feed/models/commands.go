package models

import (
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// CreateFeedCommand represents the input for creating a new feed.
// Maps to database.PublicFeedsInsert.
// Used by: POST /feeds
type CreateFeedCommand struct {
	Name   string `form:"name" json:"name" validate:"required,max=255"`
	URL    string `form:"url" json:"url" validate:"required,http_url"`
	UserID string `param:"-"`
}

// UpdateFeedCommand represents the input for updating an existing feed.
// Maps to database.PublicFeedsUpdate (subset of fields).
// Used by: PATCH /feeds/{id}
type UpdateFeedCommand struct {
	ID     string `param:"id"`
	UserID string `param:"-"`
	Name   string `form:"name" json:"name" validate:"required,max=255"`
	URL    string `form:"url" json:"url" validate:"required,http_url"`
}

// ToInsert converts CreateFeedCommand to database.PublicFeedsInsert.
// UserID is automatically bound from authenticated session via custom binder.
// Sets fetch_after to NOW() + 5 minutes to prevent race conditions with background job.
func (c CreateFeedCommand) ToInsert() database.PublicFeedsInsert {
	fetchAfter := time.Now().Add(5 * time.Minute).Format(time.RFC3339)

	return database.PublicFeedsInsert{
		Name:       c.Name,
		Url:        c.URL,
		UserId:     c.UserID,
		FetchAfter: &fetchAfter,
		// CreatedAt, UpdatedAt, Id will be set by database
		// LastFetchStatus, LastFetchError will be set by background job
	}
}

// ToUpdate converts UpdateFeedCommand to database.PublicFeedsUpdate.
// Only updates Name field; other fields remain unchanged.
func (c UpdateFeedCommand) ToUpdate() database.PublicFeedsUpdate {
	return database.PublicFeedsUpdate{
		Name: &c.Name,
		// Other fields intentionally nil to avoid updating them
	}
}

// ToUpdateWithURLChange converts UpdateFeedCommand to database.PublicFeedsUpdate when URL has changed.
// Resets fetch-related fields
func (c UpdateFeedCommand) ToUpdateWithURLChange() database.PublicFeedsUpdate {
	fetchAfter := time.Now().Add(5 * time.Minute).Format(time.RFC3339)

	return database.PublicFeedsUpdate{
		Name:            &c.Name,
		Url:             &c.URL,
		LastFetchStatus: nil, // Reset status
		LastFetchError:  nil, // Reset error
		LastModified:    nil, // Reset Last-Modified header
		Etag:            nil, // Reset ETag header
		FetchAfter:      &fetchAfter,
	}
}
