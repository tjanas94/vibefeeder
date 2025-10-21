package feed

import (
	"testing"

	"github.com/tjanas94/vibefeeder/internal/feed/models"
)

func TestBuildDashboardURL(t *testing.T) {
	tests := []struct {
		name     string
		query    models.ListFeedsQuery
		expected string
	}{
		{
			name: "empty query returns base URL",
			query: models.ListFeedsQuery{
				Status: "",
				Search: "",
				Page:   0,
			},
			expected: "/dashboard",
		},
		{
			name: "status 'all' is omitted from URL",
			query: models.ListFeedsQuery{
				Status: "all",
				Search: "",
				Page:   0,
			},
			expected: "/dashboard",
		},
		{
			name: "page 1 is omitted from URL",
			query: models.ListFeedsQuery{
				Status: "",
				Search: "",
				Page:   1,
			},
			expected: "/dashboard",
		},
		{
			name: "search parameter is included",
			query: models.ListFeedsQuery{
				Search: "test feed",
				Status: "",
				Page:   0,
			},
			expected: "/dashboard?search=test+feed",
		},
		{
			name: "status 'working' is included",
			query: models.ListFeedsQuery{
				Status: "working",
				Search: "",
				Page:   0,
			},
			expected: "/dashboard?status=working",
		},
		{
			name: "status 'error' is included",
			query: models.ListFeedsQuery{
				Status: "error",
				Search: "",
				Page:   0,
			},
			expected: "/dashboard?status=error",
		},
		{
			name: "status 'pending' is included",
			query: models.ListFeedsQuery{
				Status: "pending",
				Search: "",
				Page:   0,
			},
			expected: "/dashboard?status=pending",
		},
		{
			name: "page 2 is included",
			query: models.ListFeedsQuery{
				Status: "",
				Search: "",
				Page:   2,
			},
			expected: "/dashboard?page=2",
		},
		{
			name: "page 10 is included",
			query: models.ListFeedsQuery{
				Status: "",
				Search: "",
				Page:   10,
			},
			expected: "/dashboard?page=10",
		},
		{
			name: "search and status are combined",
			query: models.ListFeedsQuery{
				Search: "news",
				Status: "working",
				Page:   0,
			},
			expected: "/dashboard?search=news&status=working",
		},
		{
			name: "search and page are combined",
			query: models.ListFeedsQuery{
				Search: "tech",
				Status: "",
				Page:   3,
			},
			expected: "/dashboard?page=3&search=tech",
		},
		{
			name: "status and page are combined",
			query: models.ListFeedsQuery{
				Search: "",
				Status: "error",
				Page:   5,
			},
			expected: "/dashboard?page=5&status=error",
		},
		{
			name: "all parameters are combined",
			query: models.ListFeedsQuery{
				Search: "golang",
				Status: "pending",
				Page:   2,
			},
			expected: "/dashboard?page=2&search=golang&status=pending",
		},
		{
			name: "search with special characters is URL encoded",
			query: models.ListFeedsQuery{
				Search: "test&feed=true",
				Status: "",
				Page:   0,
			},
			expected: "/dashboard?search=test%26feed%3Dtrue",
		},
		{
			name: "search with spaces is URL encoded",
			query: models.ListFeedsQuery{
				Search: "my awesome feed",
				Status: "",
				Page:   0,
			},
			expected: "/dashboard?search=my+awesome+feed",
		},
		{
			name: "empty search is omitted",
			query: models.ListFeedsQuery{
				Search: "",
				Status: "working",
				Page:   2,
			},
			expected: "/dashboard?page=2&status=working",
		},
		{
			name: "all defaults (status=all, page=1, empty search)",
			query: models.ListFeedsQuery{
				Search: "",
				Status: "all",
				Page:   1,
			},
			expected: "/dashboard",
		},
		{
			name: "UserID is ignored in URL building",
			query: models.ListFeedsQuery{
				UserID: "user-123",
				Search: "test",
				Status: "working",
				Page:   2,
			},
			expected: "/dashboard?page=2&search=test&status=working",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDashboardURL(tt.query)
			if result != tt.expected {
				t.Errorf("buildDashboardURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildDashboardURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		query    models.ListFeedsQuery
		expected string
	}{
		{
			name: "negative page is included (edge case - should be validated elsewhere)",
			query: models.ListFeedsQuery{
				Page: -1,
			},
			expected: "/dashboard",
		},
		{
			name: "very large page number",
			query: models.ListFeedsQuery{
				Page: 999999,
			},
			expected: "/dashboard?page=999999",
		},
		{
			name: "search with unicode characters",
			query: models.ListFeedsQuery{
				Search: "测试",
			},
			expected: "/dashboard?search=%E6%B5%8B%E8%AF%95",
		},
		{
			name: "search with emoji",
			query: models.ListFeedsQuery{
				Search: "feed 🚀",
			},
			expected: "/dashboard?search=feed+%F0%9F%9A%80",
		},
		{
			name: "very long search query",
			query: models.ListFeedsQuery{
				Search: "this is a very long search query that contains many words and should still be properly encoded in the URL without any issues",
			},
			expected: "/dashboard?search=this+is+a+very+long+search+query+that+contains+many+words+and+should+still+be+properly+encoded+in+the+URL+without+any+issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDashboardURL(tt.query)
			if result != tt.expected {
				t.Errorf("buildDashboardURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkBuildDashboardURL(b *testing.B) {
	query := models.ListFeedsQuery{
		Search: "test feed",
		Status: "working",
		Page:   5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildDashboardURL(query)
	}
}

func BenchmarkBuildDashboardURL_Empty(b *testing.B) {
	query := models.ListFeedsQuery{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildDashboardURL(query)
	}
}
