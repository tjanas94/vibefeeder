package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListFeedsQuery_SetDefaults tests the SetDefaults method
func TestListFeedsQuery_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		initial  ListFeedsQuery
		expected ListFeedsQuery
	}{
		{
			name:    "all fields empty",
			initial: ListFeedsQuery{},
			expected: ListFeedsQuery{
				Status: "all",
				Page:   1,
			},
		},
		{
			name: "status already set",
			initial: ListFeedsQuery{
				Status: "working",
				Page:   0,
			},
			expected: ListFeedsQuery{
				Status: "working",
				Page:   1,
			},
		},
		{
			name: "page already set",
			initial: ListFeedsQuery{
				Status: "",
				Page:   5,
			},
			expected: ListFeedsQuery{
				Status: "all",
				Page:   5,
			},
		},
		{
			name: "both already set",
			initial: ListFeedsQuery{
				Status: "error",
				Page:   3,
			},
			expected: ListFeedsQuery{
				Status: "error",
				Page:   3,
			},
		},
		{
			name: "with user id and search",
			initial: ListFeedsQuery{
				UserID: "user-123",
				Search: "golang",
			},
			expected: ListFeedsQuery{
				UserID: "user-123",
				Search: "golang",
				Status: "all",
				Page:   1,
			},
		},
		{
			name: "status 'all' is preserved",
			initial: ListFeedsQuery{
				Status: "all",
				Page:   2,
			},
			expected: ListFeedsQuery{
				Status: "all",
				Page:   2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.initial
			query.SetDefaults()
			assert.Equal(t, tt.expected, query)
		})
	}
}

// TestListFeedsQuery_GetStatusFilter tests the GetStatusFilter method
func TestListFeedsQuery_GetStatusFilter(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		expectFilter   bool
		expectedFilter *StatusFilter
	}{
		{
			name:         "status 'working' returns IN filter with success",
			status:       "working",
			expectFilter: true,
			expectedFilter: &StatusFilter{
				FilterType: "IN",
				Column:     "last_fetch_status",
				Values:     []string{"success"},
			},
		},
		{
			name:         "status 'error' returns IN filter with error statuses",
			status:       "error",
			expectFilter: true,
			expectedFilter: &StatusFilter{
				FilterType: "IN",
				Column:     "last_fetch_status",
				Values:     []string{"temporary_error", "permanent_error", "unauthorized"},
			},
		},
		{
			name:         "status 'pending' returns IS_NULL filter",
			status:       "pending",
			expectFilter: true,
			expectedFilter: &StatusFilter{
				FilterType: "IS_NULL",
				Column:     "last_fetched_at",
			},
		},
		{
			name:           "status 'all' returns no filter",
			status:         "all",
			expectFilter:   false,
			expectedFilter: nil,
		},
		{
			name:           "empty status returns no filter",
			status:         "",
			expectFilter:   false,
			expectedFilter: nil,
		},
		{
			name:           "unknown status returns no filter",
			status:         "unknown",
			expectFilter:   false,
			expectedFilter: nil,
		},
		{
			name:           "invalid status returns no filter",
			status:         "invalid-status",
			expectFilter:   false,
			expectedFilter: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ListFeedsQuery{Status: tt.status}
			filter, hasFilter := query.GetStatusFilter()

			assert.Equal(t, tt.expectFilter, hasFilter)
			if tt.expectFilter {
				require.NotNil(t, filter)
				assert.Equal(t, tt.expectedFilter.FilterType, filter.FilterType)
				assert.Equal(t, tt.expectedFilter.Column, filter.Column)
				assert.Equal(t, tt.expectedFilter.Values, filter.Values)
			} else {
				assert.Nil(t, filter)
			}
		})
	}
}

// TestStatusFilter_Values tests that the filter values are correct
func TestStatusFilter_Values(t *testing.T) {
	t.Run("working filter has only success", func(t *testing.T) {
		query := ListFeedsQuery{Status: "working"}
		filter, ok := query.GetStatusFilter()
		require.True(t, ok)
		assert.Len(t, filter.Values, 1)
		assert.Contains(t, filter.Values, "success")
	})

	t.Run("error filter has all error types", func(t *testing.T) {
		query := ListFeedsQuery{Status: "error"}
		filter, ok := query.GetStatusFilter()
		require.True(t, ok)
		assert.Len(t, filter.Values, 3)
		assert.Contains(t, filter.Values, "temporary_error")
		assert.Contains(t, filter.Values, "permanent_error")
		assert.Contains(t, filter.Values, "unauthorized")
	})

	t.Run("pending filter has no values (IS_NULL)", func(t *testing.T) {
		query := ListFeedsQuery{Status: "pending"}
		filter, ok := query.GetStatusFilter()
		require.True(t, ok)
		assert.Equal(t, "IS_NULL", filter.FilterType)
		assert.Nil(t, filter.Values)
	})
}

// TestListFeedsQuery_Complete tests a complete workflow
func TestListFeedsQuery_Complete(t *testing.T) {
	t.Run("typical query for working feeds on page 2", func(t *testing.T) {
		query := ListFeedsQuery{
			UserID: "user-123",
			Search: "tech",
			Status: "working",
			Page:   2,
		}

		query.SetDefaults()

		assert.Equal(t, "user-123", query.UserID)
		assert.Equal(t, "tech", query.Search)
		assert.Equal(t, "working", query.Status)
		assert.Equal(t, 2, query.Page)

		filter, ok := query.GetStatusFilter()
		require.True(t, ok)
		assert.Equal(t, "IN", filter.FilterType)
		assert.Equal(t, "last_fetch_status", filter.Column)
		assert.Equal(t, []string{"success"}, filter.Values)
	})

	t.Run("query with no parameters gets defaults", func(t *testing.T) {
		query := ListFeedsQuery{
			UserID: "user-456",
		}

		query.SetDefaults()

		assert.Equal(t, "user-456", query.UserID)
		assert.Empty(t, query.Search)
		assert.Equal(t, "all", query.Status)
		assert.Equal(t, 1, query.Page)

		_, ok := query.GetStatusFilter()
		assert.False(t, ok, "all status should not return a filter")
	})

	t.Run("query for error feeds with search", func(t *testing.T) {
		query := ListFeedsQuery{
			UserID: "user-789",
			Search: "broken",
			Status: "error",
		}

		query.SetDefaults()

		assert.Equal(t, "error", query.Status)
		assert.Equal(t, 1, query.Page)

		filter, ok := query.GetStatusFilter()
		require.True(t, ok)
		assert.Equal(t, "IN", filter.FilterType)
		assert.Len(t, filter.Values, 3)
	})

	t.Run("query for pending feeds", func(t *testing.T) {
		query := ListFeedsQuery{
			UserID: "user-999",
			Status: "pending",
			Page:   1,
		}

		query.SetDefaults()

		filter, ok := query.GetStatusFilter()
		require.True(t, ok)
		assert.Equal(t, "IS_NULL", filter.FilterType)
		assert.Equal(t, "last_fetched_at", filter.Column)
	})
}

// TestStatusFilter_EdgeCases tests edge cases for StatusFilter
func TestStatusFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		status string
		check  func(t *testing.T, filter *StatusFilter, ok bool)
	}{
		{
			name:   "case sensitive status",
			status: "Working",
			check: func(t *testing.T, filter *StatusFilter, ok bool) {
				assert.False(t, ok, "status should be case sensitive")
			},
		},
		{
			name:   "status with whitespace",
			status: " working ",
			check: func(t *testing.T, filter *StatusFilter, ok bool) {
				assert.False(t, ok, "status with whitespace should not match")
			},
		},
		{
			name:   "partial status match",
			status: "work",
			check: func(t *testing.T, filter *StatusFilter, ok bool) {
				assert.False(t, ok, "partial status should not match")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ListFeedsQuery{Status: tt.status}
			filter, ok := query.GetStatusFilter()
			tt.check(t, filter, ok)
		})
	}
}

// BenchmarkSetDefaults benchmarks the SetDefaults method
func BenchmarkSetDefaults(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := ListFeedsQuery{}
		query.SetDefaults()
	}
}

// BenchmarkGetStatusFilter benchmarks the GetStatusFilter method
func BenchmarkGetStatusFilter(b *testing.B) {
	tests := []struct {
		name   string
		status string
	}{
		{"working", "working"},
		{"error", "error"},
		{"pending", "pending"},
		{"all", "all"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			query := ListFeedsQuery{Status: tt.status}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = query.GetStatusFilter()
			}
		})
	}
}

// BenchmarkGetStatusFilter_All benchmarks all status types together
func BenchmarkGetStatusFilter_All(b *testing.B) {
	queries := []ListFeedsQuery{
		{Status: "working"},
		{Status: "error"},
		{Status: "pending"},
		{Status: "all"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, query := range queries {
			_, _ = query.GetStatusFilter()
		}
	}
}
