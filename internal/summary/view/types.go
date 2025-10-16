package view

import "github.com/tjanas94/vibefeeder/internal/summary/models"

// GenerateSummaryActionProps contains props for the GenerateSummaryAction component.
type GenerateSummaryActionProps struct {
	// ButtonText is the text displayed on the submit button
	ButtonText string

	// AriaLabel provides an accessible label for screen readers
	AriaLabel string

	// ButtonSize is an optional CSS class for button sizing (e.g., "btn-sm", "btn-lg")
	// Leave empty for default size
	ButtonSize string
}

// ContentProps contains props for the Content component.
type ContentProps struct {
	// Summary is the summary data to display
	Summary models.SummaryViewModel

	// CanGenerate indicates whether the user can generate a new summary
	// (true if user has at least one working feed)
	CanGenerate bool
}

// EmptyStateProps contains props for the EmptyState component.
type EmptyStateProps struct {
	// CanGenerate indicates whether the user can generate a summary
	// (true if user has at least one working feed)
	CanGenerate bool
}
