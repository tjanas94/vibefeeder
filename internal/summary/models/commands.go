package models

import "github.com/tjanas94/vibefeeder/internal/shared/database"

// GenerateSummaryCommand represents the input for generating a new summary.
// No input fields required - summary is generated from user's articles from last 24h.
// Used by: POST /summaries
type GenerateSummaryCommand struct {
	// Empty command - all data comes from database queries
	// UserID will be provided by authenticated session
}

// ToInsert converts the generated summary content to database.PublicSummariesInsert.
// UserID must be set from authenticated session, Content from AI generation.
func ToInsert(userID string, content string) database.PublicSummariesInsert {
	return database.PublicSummariesInsert{
		UserId:  userID,
		Content: content,
		// CreatedAt, Id will be set by database
	}
}
