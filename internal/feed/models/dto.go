package models

import (
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
	sharedmodels "github.com/tjanas94/vibefeeder/internal/shared/models"
)

// FeedListViewModel represents the feed list display with empty state support.
// Used by: GET /feeds
type FeedListViewModel struct {
	Feeds          []FeedItemViewModel              `json:"feeds"`
	ShowEmptyState bool                             `json:"show_empty_state"`
	Pagination     sharedmodels.PaginationViewModel `json:"pagination"`
}

// FeedItemViewModel represents a single feed item for display.
// Derived from database.PublicFeedsSelect with computed HasError field.
// Used by: GET /feeds, POST /feeds, POST /feeds/{id}, GET /dashboard
type FeedItemViewModel struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	HasError      bool      `json:"has_error"`     // Computed: LastFetchError != nil
	ErrorMessage  string    `json:"error_message"` // From last_fetch_error
	LastFetchedAt time.Time `json:"last_fetched_at"`
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

// FeedListErrorViewModel represents errors during feed list retrieval.
// Used by: GET /feeds
type FeedListErrorViewModel struct {
	ErrorMessage string `json:"error_message"`
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

	// Parse last_fetched_at timestamp
	if dbFeed.LastFetchedAt != nil {
		if lastFetchedAt, err := time.Parse(time.RFC3339, *dbFeed.LastFetchedAt); err == nil {
			vm.LastFetchedAt = lastFetchedAt
		}
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

// NewFeedFormErrorFromFieldErrors creates FeedFormErrorViewModel from field error map.
// Accepts a map of field names to error messages (from validator.ParseFieldErrors).
// Returns a view model with errors mapped to the appropriate fields.
func NewFeedFormErrorFromFieldErrors(fieldErrors map[string]string) FeedFormErrorViewModel {
	vm := FeedFormErrorViewModel{}

	if fieldErrors == nil {
		vm.GeneralError = "Invalid request"
		return vm
	}

	// Map parsed errors to view model fields
	if nameErr, ok := fieldErrors["Name"]; ok {
		vm.NameError = nameErr
	}
	if urlErr, ok := fieldErrors["URL"]; ok {
		vm.URLError = urlErr
	}

	return vm
}
