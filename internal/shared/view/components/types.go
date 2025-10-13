package components

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

	// Title is the modal header text
	Title string

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

	// Message is the alert text to display
	Message string

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
