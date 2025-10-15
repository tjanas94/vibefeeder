package components

import (
	"github.com/a-h/templ"
	"github.com/tjanas94/vibefeeder/internal/shared/models"
)

// PaginationProps defines configuration for the Pagination component.
type PaginationProps struct {
	// Pagination contains the current page state and metadata
	Pagination models.PaginationViewModel

	// BaseURL is the endpoint for pagination requests (e.g., "/feeds")
	BaseURL string

	// FormID is the HTML id of the form to include in htmx requests
	// Example: "#feed-filter-form"
	FormID string
}

// EmptyStateProps defines configuration for the EmptyState component.
// EmptyState is used to display a centered message when there is no content to show.
type EmptyStateProps struct {
	// Icon is an emoji or icon to display above the title
	// Example: "📡", "🔍", "📝"
	Icon string

	// Title is the main heading text
	Title string

	// Description is optional explanatory text displayed below the title
	Description string

	// ActionText is the optional button text for a call-to-action
	ActionText string

	// ActionAttrs are optional attributes for the action button
	// Used for htmx, Alpine.js, or other interactive behaviors
	ActionAttrs templ.Attributes
}

// LoadingSpinnerProps defines configuration for the LoadingSpinner component.
type LoadingSpinnerProps struct {
	// Size of the spinner: "sm", "md", "lg" (default: "lg")
	Size string
}

// ModalProps defines configuration for the Modal component.
type ModalProps struct {
	// ID is the unique HTML id for the modal element
	ID string

	// ContentID is the unique HTML id for the content container inside the modal
	ContentID string

	// AlpineStateVar is the Alpine.js variable name controlling modal visibility
	// Example: "isSummaryModalOpen"
	AlpineStateVar string

	// MaxWidth defines the modal width: "sm", "md", "lg", "xl", "2xl", "4xl" (default: "2xl")
	MaxWidth string
}

// AlertProps defines configuration for the Alert component.
type AlertProps struct {
	// Type of alert: "error", "success", "warning", "info" (default: "info")
	Type string

	// ShowIcon determines whether to show an icon (default: true)
	ShowIcon bool

	// UseOOB determines whether to use hx-swap-oob for out-of-band updates
	// When true, alert will be injected into #global-errors container
	UseOOB bool
}

// NavbarProps defines configuration for the Navbar component.
// Navbar is only rendered for authenticated users, so IsAuthenticated is not needed.
// The layout decides whether to show the navbar based on authentication state.
type NavbarProps struct {
	// UserEmail is the email address of the authenticated user
	// Displayed in the navbar for user identification
	UserEmail string
}

// ToastProps defines configuration for the Toast notification component.
type ToastProps struct {
	// Type of toast: "error", "success", "warning", "info" (default: "info")
	Type string

	// Message is the toast text to display
	Message string

	// Duration is the time in milliseconds before auto-dismiss (default: 3000)
	// Set to 0 to use default duration
	Duration int

	// UseOOB determines whether to use hx-swap-oob for out-of-band updates
	// When true, toast will be injected into #toast-container
	UseOOB bool
}

// BadgeProps defines configuration for the Badge component.
// Badge renders a small label with colored background to indicate status or category.
type BadgeProps struct {
	// Text is the badge text to display
	Text string

	// Type of badge: "success", "error", "warning", "info", "neutral" (default: "neutral")
	Type string

	// Size of badge: "sm", "md", "lg" (default: "md")
	Size string

	// Icon is optional icon (emoji or SVG) to display before text
	Icon string

	// Tooltip is optional tooltip text for the badge
	Tooltip string
}

// FormFieldProps defines configuration for the FormField component.
// FormField renders a simple form input field with label and error display.
type FormFieldProps struct {
	// Label is the text for the input label
	Label string

	// ID is the HTML id attribute for the input (used in label's for attribute)
	ID string

	// Name is the HTML name attribute for the input
	Name string

	// Type is the input type: "text", "url", "email", "password" (default: "text")
	Type string

	// Value is the current value of the input
	Value string

	// Placeholder is the placeholder text for the input
	Placeholder string

	// Error is the validation error message to display
	Error string

	// Required indicates if the field is required (shows * in label and adds required attribute)
	Required bool
}
