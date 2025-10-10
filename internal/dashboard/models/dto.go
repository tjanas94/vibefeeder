package models

import (
	feedModels "github.com/tjanas94/vibefeeder/internal/feed/models"
	summaryModels "github.com/tjanas94/vibefeeder/internal/summary/models"
)

// DashboardViewModel represents the main dashboard view.
// Aggregates data from multiple features (auth, feeds, summaries).
// Used by: GET /dashboard
type DashboardViewModel struct {
	User           UserInfo                        `json:"user"`
	Feeds          []feedModels.FeedItemViewModel  `json:"feeds"`
	LatestSummary  *summaryModels.SummaryViewModel `json:"latest_summary,omitempty"`
	ShowEmptyState bool                            `json:"show_empty_state"` // true when no feeds exist
}

// UserInfo represents authenticated user information.
// Derived from auth.users (via Supabase Auth session).
// Used by: GET /dashboard
type UserInfo struct {
	Email string `json:"email"`
}
