package models

// CheckURLTakenQuery represents the input parameters for checking if a URL is already in use.
// Used by: Feed creation and update validation
type CheckURLTakenQuery struct {
	UserID        string // Required: User ID from authenticated session
	URL           string // Required: URL to check for duplicates
	ExcludeFeedID string // Optional: Feed ID to exclude from check (used during updates)
}

// ListFeedsQuery represents the input parameters for listing feeds.
// Used by: GET /feeds
type ListFeedsQuery struct {
	UserID string `query:"-"`      // Required: User ID from authenticated session (set by handler)
	Search string `query:"search"` // Optional: Search phrase for feed names (case-insensitive)
	Status string `query:"status"` // Optional: Filter by last fetch status (all, working, error, pending)
	Page   int    `query:"page"`   // Optional: Page number (1-indexed), default: 1
}

// SetDefaults sets default values for optional query parameters
// and sanitizes invalid values
func (q *ListFeedsQuery) SetDefaults() {
	// Sanitize status - only allow valid values
	switch q.Status {
	case "all", "working", "error", "pending":
		// Valid - keep it
	default:
		// Invalid or empty - default to "all"
		q.Status = "all"
	}

	// Sanitize page - must be >= 1
	if q.Page < 1 {
		q.Page = 1
	}
}

// StatusFilter describes a filter operation for feed statuses
type StatusFilter struct {
	FilterType string   // e.g., "IN", "IS_NULL"
	Column     string   // The database column to filter on
	Values     []string // Values for the "IN" filter type
}

// GetStatusFilter returns a StatusFilter based on the query's status parameter.
// This allows for more complex filtering logic than a simple IN clause.
func (q *ListFeedsQuery) GetStatusFilter() (*StatusFilter, bool) {
	switch q.Status {
	case "working":
		return &StatusFilter{
			FilterType: "IN",
			Column:     "last_fetch_status",
			Values:     []string{"success"},
		}, true
	case "error":
		return &StatusFilter{
			FilterType: "IN",
			Column:     "last_fetch_status",
			Values:     []string{"temporary_error", "permanent_error", "unauthorized"},
		}, true
	case "pending":
		return &StatusFilter{
			FilterType: "IS_NULL",
			Column:     "last_fetched_at",
		}, true
	default: // "all"
		return nil, false
	}
}
