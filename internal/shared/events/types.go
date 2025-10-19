package events

const (
	// Authentication events
	EventUserRegistered    = "user_registered"
	EventUserLogin         = "user_login"
	EventUserPasswordReset = "user_password_reset"

	// Feed events
	EventFeedAdded = "feed_added"

	// Summary events
	EventSummaryGenerated = "summary_generated"
)
