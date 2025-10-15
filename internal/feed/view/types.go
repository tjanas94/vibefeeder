package view

// FeedSearchFilterProps defines the properties required to render the feed
// search + status filter form (see FeedSearchFilter templ component).
// These values typically originate from query parameters so that the form
// reflects the current filter state after HTMX-driven updates.
type FeedSearchFilterProps struct {
	// Search is the initial search query (feed name substring)
	Search string

	// Status is the selected status filter:
	// "all", "working", "pending", "error"
	Status string
}
