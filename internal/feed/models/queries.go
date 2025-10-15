package models

// ListFeedsQuery represents the input parameters for listing feeds.
// Used by: GET /feeds
type ListFeedsQuery struct {
	UserID string `query:"-"`                                                           // Required: User ID from authenticated session (set by handler)
	Search string `query:"search"`                                                      // Optional: Search phrase for feed names (case-insensitive)
	Status string `query:"status" validate:"omitempty,oneof=all working error pending"` // Optional: Filter by last fetch status
	Page   int    `query:"page" validate:"omitempty,gte=1"`                             // Optional: Page number (1-indexed), default: 1
}

// SetDefaults sets default values for optional query parameters
func (q *ListFeedsQuery) SetDefaults() {
	if q.Status == "" {
		q.Status = "all"
	}
	if q.Page == 0 {
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
