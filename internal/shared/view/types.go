package view

// LayoutProps contains the data needed to render the base layout
type LayoutProps struct {
	Title string
}

// ErrorPageProps contains the data needed to render a standalone error page
type ErrorPageProps struct {
	Code    int
	Title   string
	Message string
}

// ErrorFragmentProps contains the data needed to render an error fragment for HTMX requests
type ErrorFragmentProps struct {
	Message string
}
