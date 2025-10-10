package models

import "github.com/tjanas94/vibefeeder/internal/shared/database"

// CreateFeedCommand represents the input for creating a new feed.
// Maps to database.PublicFeedsInsert.
// Used by: POST /feeds
type CreateFeedCommand struct {
	Name string `form:"name" json:"name" validate:"required,max=255"`
	URL  string `form:"url" json:"url" validate:"required,url"`
}

// UpdateFeedCommand represents the input for updating an existing feed.
// Maps to database.PublicFeedsUpdate (subset of fields).
// Used by: POST /feeds/{id}
type UpdateFeedCommand struct {
	Name string `form:"name" json:"name" validate:"required,max=255"`
	URL  string `form:"url" json:"url" validate:"required,url"`
}

// ToInsert converts CreateFeedCommand to database.PublicFeedsInsert.
// UserID must be set separately by the handler from authenticated session.
func (c CreateFeedCommand) ToInsert(userID string) database.PublicFeedsInsert {
	return database.PublicFeedsInsert{
		Name:   c.Name,
		Url:    c.URL,
		UserId: userID,
		// CreatedAt, UpdatedAt, Id will be set by database
		// LastFetchStatus, LastFetchError will be set by background job
	}
}

// ToUpdate converts UpdateFeedCommand to database.PublicFeedsUpdate.
// Only updates Name and URL fields; other fields remain unchanged.
func (c UpdateFeedCommand) ToUpdate() database.PublicFeedsUpdate {
	return database.PublicFeedsUpdate{
		Name: &c.Name,
		Url:  &c.URL,
		// Other fields intentionally nil to avoid updating them
	}
}
