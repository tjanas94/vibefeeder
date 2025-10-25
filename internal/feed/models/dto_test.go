package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	sharedmodels "github.com/tjanas94/vibefeeder/internal/shared/models"
)

// TestNewFeedItemFromDB tests the NewFeedItemFromDB function
func TestNewFeedItemFromDB(t *testing.T) {
	successStatus := "success"
	tempErrorStatus := "temporary_error"
	permErrorStatus := "permanent_error"
	unauthorizedStatus := "unauthorized"
	pendingStatus := "pending"
	errorMessage := "Failed to fetch feed"
	lastFetchedAt := "2024-01-15T12:00:00Z"

	tests := []struct {
		name     string
		dbFeed   database.PublicFeedsSelect
		validate func(t *testing.T, vm FeedItemViewModel)
	}{
		{
			name: "feed with success status",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-1",
				Name:            "Tech Blog",
				Url:             "https://example.com/feed",
				LastFetchStatus: &successStatus,
				LastFetchedAt:   &lastFetchedAt,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.Equal(t, "feed-1", vm.ID)
				assert.Equal(t, "Tech Blog", vm.Name)
				assert.Equal(t, "https://example.com/feed", vm.URL)
				assert.False(t, vm.HasError)
				assert.Empty(t, vm.ErrorMessage)
				expectedTime, _ := time.Parse(time.RFC3339, lastFetchedAt)
				assert.Equal(t, expectedTime, vm.LastFetchedAt)
			},
		},
		{
			name: "feed with temporary error status",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-2",
				Name:            "News Site",
				Url:             "https://news.example.com/rss",
				LastFetchStatus: &tempErrorStatus,
				LastFetchError:  &errorMessage,
				LastFetchedAt:   &lastFetchedAt,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.Equal(t, "feed-2", vm.ID)
				assert.True(t, vm.HasError)
				assert.Equal(t, errorMessage, vm.ErrorMessage)
			},
		},
		{
			name: "feed with permanent error status",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-3",
				Name:            "Broken Feed",
				Url:             "https://broken.example.com/feed",
				LastFetchStatus: &permErrorStatus,
				LastFetchError:  &errorMessage,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.Equal(t, "feed-3", vm.ID)
				assert.True(t, vm.HasError)
				assert.Equal(t, errorMessage, vm.ErrorMessage)
			},
		},
		{
			name: "feed with unauthorized status",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-4",
				Name:            "Protected Feed",
				Url:             "https://protected.example.com/feed",
				LastFetchStatus: &unauthorizedStatus,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.Equal(t, "feed-4", vm.ID)
				assert.False(t, vm.HasError, "unauthorized should not set HasError")
			},
		},
		{
			name: "feed with pending status",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-5",
				Name:            "Pending Feed",
				Url:             "https://pending.example.com/feed",
				LastFetchStatus: &pendingStatus,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.False(t, vm.HasError)
			},
		},
		{
			name: "feed with nil last_fetch_status",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-6",
				Name:            "New Feed",
				Url:             "https://new.example.com/feed",
				LastFetchStatus: nil,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.False(t, vm.HasError)
				assert.Empty(t, vm.ErrorMessage)
			},
		},
		{
			name: "feed with error status but no error message",
			dbFeed: database.PublicFeedsSelect{
				Id:              "feed-7",
				Name:            "Error Feed",
				Url:             "https://error.example.com/feed",
				LastFetchStatus: &tempErrorStatus,
				LastFetchError:  nil,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.True(t, vm.HasError)
				assert.Empty(t, vm.ErrorMessage)
			},
		},
		{
			name: "feed with nil last_fetched_at",
			dbFeed: database.PublicFeedsSelect{
				Id:            "feed-8",
				Name:          "Never Fetched",
				Url:           "https://never.example.com/feed",
				LastFetchedAt: nil,
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.True(t, vm.LastFetchedAt.IsZero())
			},
		},
		{
			name: "feed with invalid timestamp format",
			dbFeed: database.PublicFeedsSelect{
				Id:            "feed-9",
				Name:          "Invalid Time",
				Url:           "https://invalid.example.com/feed",
				LastFetchedAt: stringPtr("invalid-timestamp"),
			},
			validate: func(t *testing.T, vm FeedItemViewModel) {
				assert.True(t, vm.LastFetchedAt.IsZero(), "should remain zero on parse error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewFeedItemFromDB(tt.dbFeed)
			tt.validate(t, vm)
		})
	}
}

// TestNewFeedFormErrorFromFieldErrors tests the NewFeedFormErrorFromFieldErrors function
func TestNewFeedFormErrorFromFieldErrors(t *testing.T) {
	tests := []struct {
		name        string
		fieldErrors map[string]string
		want        FeedFormErrorViewModel
	}{
		{
			name: "name and url errors",
			fieldErrors: map[string]string{
				"Name": "Name is required",
				"URL":  "Invalid URL format",
			},
			want: FeedFormErrorViewModel{
				NameError: "Name is required",
				URLError:  "Invalid URL format",
			},
		},
		{
			name: "only name error",
			fieldErrors: map[string]string{
				"Name": "Name must be at least 3 characters",
			},
			want: FeedFormErrorViewModel{
				NameError: "Name must be at least 3 characters",
			},
		},
		{
			name: "only url error",
			fieldErrors: map[string]string{
				"URL": "URL must be https",
			},
			want: FeedFormErrorViewModel{
				URLError: "URL must be https",
			},
		},
		{
			name:        "nil field errors",
			fieldErrors: nil,
			want: FeedFormErrorViewModel{
				GeneralError: "Invalid request",
			},
		},
		{
			name:        "empty field errors map",
			fieldErrors: map[string]string{},
			want:        FeedFormErrorViewModel{},
		},
		{
			name: "unknown field errors are ignored",
			fieldErrors: map[string]string{
				"Name":        "Name is required",
				"Description": "This field doesn't exist",
				"Unknown":     "This too",
			},
			want: FeedFormErrorViewModel{
				NameError: "Name is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFeedFormErrorFromFieldErrors(tt.fieldErrors)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestNewFeedFormForAdd tests the NewFeedFormForAdd function
func TestNewFeedFormForAdd(t *testing.T) {
	vm := NewFeedFormForAdd()

	assert.Equal(t, "add", vm.Mode)
	assert.Equal(t, "/feeds", vm.PostURL)
	assert.Equal(t, "feed-add-form-errors", vm.FormTargetID)
	assert.Empty(t, vm.FeedID)
	assert.Empty(t, vm.Name)
	assert.Empty(t, vm.URL)
	assert.Equal(t, FeedFormErrorViewModel{}, vm.Errors)
}

// TestNewFeedFormForEdit tests the NewFeedFormForEdit function
func TestNewFeedFormForEdit(t *testing.T) {
	dbFeed := database.PublicFeedsSelect{
		Id:   "feed-123",
		Name: "My Feed",
		Url:  "https://example.com/feed",
	}

	vm := NewFeedFormForEdit(dbFeed)

	assert.Equal(t, "edit", vm.Mode)
	assert.Equal(t, "/feeds/feed-123", vm.PostURL)
	assert.Equal(t, "feed-edit-form-errors-feed-123", vm.FormTargetID)
	assert.Equal(t, "feed-123", vm.FeedID)
	assert.Equal(t, "My Feed", vm.Name)
	assert.Equal(t, "https://example.com/feed", vm.URL)
	assert.Equal(t, FeedFormErrorViewModel{}, vm.Errors)
}

// TestNewFeedFormWithErrors tests the NewFeedFormWithErrors function
func TestNewFeedFormWithErrors(t *testing.T) {
	errors := FeedFormErrorViewModel{
		NameError: "Name is required",
		URLError:  "Invalid URL",
	}

	tests := []struct {
		name     string
		mode     string
		feedID   string
		feedName string
		url      string
		errors   FeedFormErrorViewModel
		validate func(t *testing.T, vm FeedFormViewModel)
	}{
		{
			name:     "add mode with errors",
			mode:     "add",
			feedID:   "",
			feedName: "Test",
			url:      "invalid-url",
			errors:   errors,
			validate: func(t *testing.T, vm FeedFormViewModel) {
				assert.Equal(t, "add", vm.Mode)
				assert.Equal(t, "/feeds", vm.PostURL)
				assert.Equal(t, "feed-add-form-errors", vm.FormTargetID)
				assert.Empty(t, vm.FeedID)
				assert.Equal(t, "Test", vm.Name)
				assert.Equal(t, "invalid-url", vm.URL)
				assert.Equal(t, errors, vm.Errors)
			},
		},
		{
			name:     "edit mode with errors",
			mode:     "edit",
			feedID:   "feed-456",
			feedName: "Updated Feed",
			url:      "https://updated.example.com/feed",
			errors:   errors,
			validate: func(t *testing.T, vm FeedFormViewModel) {
				assert.Equal(t, "edit", vm.Mode)
				assert.Equal(t, "/feeds/feed-456", vm.PostURL)
				assert.Equal(t, "feed-edit-form-errors-feed-456", vm.FormTargetID)
				assert.Equal(t, "feed-456", vm.FeedID)
				assert.Equal(t, "Updated Feed", vm.Name)
				assert.Equal(t, "https://updated.example.com/feed", vm.URL)
				assert.Equal(t, errors, vm.Errors)
			},
		},
		{
			name:     "unknown mode treated as edit with empty feedID",
			mode:     "unknown",
			feedID:   "",
			feedName: "Test",
			url:      "test-url",
			errors:   FeedFormErrorViewModel{},
			validate: func(t *testing.T, vm FeedFormViewModel) {
				assert.Equal(t, "unknown", vm.Mode)
				// Unknown mode != "add", so it goes to else branch which creates edit-like URLs
				assert.Equal(t, "/feeds/", vm.PostURL)
				assert.Equal(t, "feed-edit-form-errors-", vm.FormTargetID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewFeedFormWithErrors(tt.mode, tt.feedID, tt.feedName, tt.url, tt.errors)
			tt.validate(t, vm)
		})
	}
}

// TestFeedListViewModel tests the FeedListViewModel structure
func TestFeedListViewModel(t *testing.T) {
	t.Run("empty feed list", func(t *testing.T) {
		vm := FeedListViewModel{
			Feeds:          []FeedItemViewModel{},
			ShowEmptyState: true,
			Pagination:     sharedmodels.BuildPagination(0, 1, 20),
		}

		assert.Empty(t, vm.Feeds)
		assert.True(t, vm.ShowEmptyState)
		assert.Empty(t, vm.ErrorMessage)
	})

	t.Run("feed list with items", func(t *testing.T) {
		feeds := []FeedItemViewModel{
			{ID: "1", Name: "Feed 1", URL: "https://feed1.com"},
			{ID: "2", Name: "Feed 2", URL: "https://feed2.com"},
		}

		vm := FeedListViewModel{
			Feeds:          feeds,
			ShowEmptyState: false,
			Pagination:     sharedmodels.BuildPagination(2, 1, 20),
		}

		assert.Len(t, vm.Feeds, 2)
		assert.False(t, vm.ShowEmptyState)
	})

	t.Run("feed list with error", func(t *testing.T) {
		vm := FeedListViewModel{
			Feeds:          []FeedItemViewModel{},
			ShowEmptyState: false,
			ErrorMessage:   "Failed to load feeds",
		}

		assert.Empty(t, vm.Feeds)
		assert.Equal(t, "Failed to load feeds", vm.ErrorMessage)
	})
}

// BenchmarkNewFeedItemFromDB benchmarks feed item creation
func BenchmarkNewFeedItemFromDB(b *testing.B) {
	status := "success"
	lastFetchedAt := "2024-01-15T12:00:00Z"

	dbFeed := database.PublicFeedsSelect{
		Id:              "feed-1",
		Name:            "Tech Blog",
		Url:             "https://example.com/feed",
		LastFetchStatus: &status,
		LastFetchedAt:   &lastFetchedAt,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFeedItemFromDB(dbFeed)
	}
}

// BenchmarkNewFeedFormErrorFromFieldErrors benchmarks error mapping
func BenchmarkNewFeedFormErrorFromFieldErrors(b *testing.B) {
	fieldErrors := map[string]string{
		"Name": "Name is required",
		"URL":  "Invalid URL format",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFeedFormErrorFromFieldErrors(fieldErrors)
	}
}

// BenchmarkNewFeedFormWithErrors benchmarks form creation with errors
func BenchmarkNewFeedFormWithErrors(b *testing.B) {
	errors := FeedFormErrorViewModel{
		NameError: "Name is required",
		URLError:  "Invalid URL",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFeedFormWithErrors("edit", "feed-123", "Test Feed", "https://example.com/feed", errors)
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}
