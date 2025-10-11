package models

import (
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// ArticleForPrompt contains only the fields needed for AI prompt generation.
// Used by: fetchRecentArticles, buildPromptFromArticles
type ArticleForPrompt struct {
	Title   string  `json:"title"`
	Content *string `json:"content"`
}

// SummaryViewModel represents a single summary for display.
// Derived from database.PublicSummariesSelect.
// Used by: GET /summaries/latest, POST /summaries, GET /dashboard
type SummaryViewModel struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// SummaryDisplayViewModel represents the summary section display with empty state support.
// Used by: GET /summaries/latest, POST /summaries
type SummaryDisplayViewModel struct {
	Summary        *SummaryViewModel `json:"summary,omitempty"`
	ShowEmptyState bool              `json:"show_empty_state"`
	CanGenerate    bool              `json:"can_generate"` // true if user has at least one working feed
}

// SummaryErrorViewModel represents errors during summary generation.
// Used by: POST /summaries
type SummaryErrorViewModel struct {
	ErrorMessage string `json:"error_message"`
}

// NewSummaryFromDB creates a SummaryViewModel from database.PublicSummariesSelect.
// Parses timestamp from database string format.
func NewSummaryFromDB(dbSummary database.PublicSummariesSelect) SummaryViewModel {
	vm := SummaryViewModel{
		ID:      dbSummary.Id,
		Content: dbSummary.Content,
	}

	// Parse created_at timestamp
	if createdAt, err := time.Parse(time.RFC3339, dbSummary.CreatedAt); err == nil {
		vm.CreatedAt = createdAt
	}

	return vm
}
