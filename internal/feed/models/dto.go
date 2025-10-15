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
	ErrorMessage   string                           `json:"error_message,omitempty"`
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

// FeedFormViewModel represents the unified form for adding or editing feeds.
// Used by: GET /feeds/new, GET /feeds/{id}/edit
type FeedFormViewModel struct {
	Mode         string                 `json:"mode"`           // "add" or "edit"
	PostURL      string                 `json:"post_url"`       // "/feeds" or "/feeds/{id}"
	FormTargetID string                 `json:"form_target_id"` // "feed-add-form-errors" or "feed-edit-form-errors-{id}"
	FeedID       string                 `json:"feed_id"`        // ID of the feed being edited (optional)
	Name         string                 `json:"name"`           // Current name
	URL          string                 `json:"url"`            // Current URL
	Errors       FeedFormErrorViewModel `json:"errors"`         // Validation errors
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

	// Parse last_fetched_at timestamp
	if dbFeed.LastFetchedAt != nil {
		if lastFetchedAt, err := time.Parse(time.RFC3339, *dbFeed.LastFetchedAt); err == nil {
			vm.LastFetchedAt = lastFetchedAt
		}
	}

	return vm
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

// NewFeedFormForAdd creates a FeedFormViewModel for adding a new feed.
func NewFeedFormForAdd() FeedFormViewModel {
	return FeedFormViewModel{
		Mode:         "add",
		PostURL:      "/feeds",
		FormTargetID: "feed-add-form-errors",
		Errors:       FeedFormErrorViewModel{},
	}
}

// NewFeedFormForEdit creates a FeedFormViewModel for editing an existing feed.
func NewFeedFormForEdit(dbFeed database.PublicFeedsSelect) FeedFormViewModel {
	return FeedFormViewModel{
		Mode:         "edit",
		PostURL:      "/feeds/" + dbFeed.Id,
		FormTargetID: "feed-edit-form-errors-" + dbFeed.Id,
		FeedID:       dbFeed.Id,
		Name:         dbFeed.Name,
		URL:          dbFeed.Url,
		Errors:       FeedFormErrorViewModel{},
	}
}

// NewFeedFormWithErrors creates a FeedFormViewModel with validation errors.
// Used to re-render the form after failed validation.
func NewFeedFormWithErrors(mode, feedID, name, url string, errors FeedFormErrorViewModel) FeedFormViewModel {
	vm := FeedFormViewModel{
		Mode:   mode,
		FeedID: feedID,
		Name:   name,
		URL:    url,
		Errors: errors,
	}

	if mode == "add" {
		vm.PostURL = "/feeds"
		vm.FormTargetID = "feed-add-form-errors"
	} else {
		vm.PostURL = "/feeds/" + feedID
		vm.FormTargetID = "feed-edit-form-errors-" + feedID
	}

	return vm
}

// DeleteConfirmationViewModel holds data for the delete confirmation modal.
// Used by: GET /feeds/{id}/delete
type DeleteConfirmationViewModel struct {
	FeedID       string `json:"feed_id"`
	FeedName     string `json:"feed_name"`
	ErrorMessage string `json:"error_message,omitempty"`
}
