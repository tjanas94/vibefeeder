package models

// ListFeedsQuery represents the input parameters for listing feeds.
// Used by: GET /feeds
type ListFeedsQuery struct {
	UserID string `query:"-"`                                                   // Required: User ID from authenticated session (set by handler)
	Search string `query:"search"`                                              // Optional: Search phrase for feed names (case-insensitive)
	Status string `query:"status" validate:"omitempty,oneof=all working error"` // Optional: Filter by last fetch status
	Page   int    `query:"page" validate:"omitempty,gte=1"`                     // Optional: Page number (1-indexed), default: 1
	Limit  int    `query:"limit" validate:"omitempty,gte=1,lte=100"`            // Optional: Number of items per page, default: 20, max: 100
}

// SetDefaults sets default values for optional query parameters
func (q *ListFeedsQuery) SetDefaults() {
	if q.Status == "" {
		q.Status = "all"
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
}

// GetStatusFilter returns database status values for filtering
// Returns (statuses, true) when filter should be applied, or (nil, false) for no filter
func (q *ListFeedsQuery) GetStatusFilter() ([]string, bool) {
	switch q.Status {
	case "working":
		return []string{"success"}, true
	case "error":
		return []string{"temporary_error", "permanent_error", "unauthorized"}, true
	default: // "all"
		return nil, false
	}
}
