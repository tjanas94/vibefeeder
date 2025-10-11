package models

import (
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// FeedListViewModel represents the feed list display with empty state support.
// Used by: GET /feeds
type FeedListViewModel struct {
	Feeds          []FeedItemViewModel `json:"feeds"`
	ShowEmptyState bool                `json:"show_empty_state"`
	Pagination     PaginationViewModel `json:"pagination"`
}

// PaginationViewModel represents pagination information for feed list.
// Used by: GET /feeds
type PaginationViewModel struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasPrevious bool `json:"has_previous"`
	HasNext     bool `json:"has_next"`
}

// FeedItemViewModel represents a single feed item for display.
// Derived from database.PublicFeedsSelect with computed HasError field.
// Used by: GET /feeds, POST /feeds, POST /feeds/{id}, GET /dashboard
type FeedItemViewModel struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	HasError     bool      `json:"has_error"`     // Computed: LastFetchError != nil
	ErrorMessage string    `json:"error_message"` // From last_fetch_error
	UpdatedAt    time.Time `json:"updated_at"`
}

// FeedEditFormViewModel represents the feed edit form data.
// Derived from database.PublicFeedsSelect (subset of fields).
// Used by: GET /feeds/{id}/edit
type FeedEditFormViewModel struct {
	FeedID string `json:"feed_id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}

// FeedFormErrorViewModel represents validation errors for feed forms.
// Used by: POST /feeds, POST /feeds/{id}
type FeedFormErrorViewModel struct {
	NameError    string `json:"name_error,omitempty"`
	URLError     string `json:"url_error,omitempty"`
	GeneralError string `json:"general_error,omitempty"`
}

// NewFeedItemFromDB creates a FeedItemViewModel from database.PublicFeedsSelect.
// Computes HasError from LastFetchError and parses timestamps.
func NewFeedItemFromDB(dbFeed database.PublicFeedsSelect) FeedItemViewModel {
	vm := FeedItemViewModel{
		ID:   dbFeed.Id,
		Name: dbFeed.Name,
		URL:  dbFeed.Url,
	}

	// Compute HasError from last_fetch_error
	if dbFeed.LastFetchError != nil && *dbFeed.LastFetchError != "" {
		vm.HasError = true
		vm.ErrorMessage = *dbFeed.LastFetchError
	}

	// Parse updated_at timestamp
	if updatedAt, err := time.Parse(time.RFC3339, dbFeed.UpdatedAt); err == nil {
		vm.UpdatedAt = updatedAt
	}

	return vm
}

// NewFeedEditFormFromDB creates a FeedEditFormViewModel from database.PublicFeedsSelect.
func NewFeedEditFormFromDB(dbFeed database.PublicFeedsSelect) FeedEditFormViewModel {
	return FeedEditFormViewModel{
		FeedID: dbFeed.Id,
		Name:   dbFeed.Name,
		URL:    dbFeed.Url,
	}
}
