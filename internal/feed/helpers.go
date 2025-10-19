package feed

import (
	"fmt"
	"net/url"

	"github.com/tjanas94/vibefeeder/internal/feed/models"
)

// buildDashboardURL builds the dashboard URL with query parameters based on the feed list query.
// This is a pure function that constructs a URL string from the query parameters.
func buildDashboardURL(query models.ListFeedsQuery) string {
	pushURL := "/dashboard"
	params := make(url.Values)

	if query.Search != "" {
		params.Set("search", query.Search)
	}
	if query.Status != "" && query.Status != "all" {
		params.Set("status", query.Status)
	}
	if query.Page > 1 {
		params.Set("page", fmt.Sprintf("%d", query.Page))
	}

	if len(params) > 0 {
		pushURL += "?" + params.Encode()
	}

	return pushURL
}
