package models

// DashboardViewModel represents the main dashboard view.
// Renders only layout and filter forms. All content (user info, feeds, summary)
// loaded via htmx partials from dedicated endpoints.
// Used by: GET /dashboard
type DashboardViewModel struct {
	// Empty - all data loaded via htmx:
	// - User info: GET /auth/me
	// - Feeds: GET /feeds
	// - Summary: GET /summaries/latest
}
