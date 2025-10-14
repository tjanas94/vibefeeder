package models

import (
	"net/url"
	"strconv"
)

// BuildPageURL constructs a URL with the page query parameter
// Uses net/url for proper URL encoding
func BuildPageURL(baseURL string, page int) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		// Fallback to base URL if parsing fails
		return baseURL
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

// GetPaginationRange calculates which page numbers to display
// Shows current page with context, and uses ellipsis for gaps
// Example for page 5 of 20: [1, -1, 4, 5, 6, -1, 20]
// Where -1 represents ellipsis
func GetPaginationRange(currentPage, totalPages int) []int {
	if totalPages <= 7 {
		// Show all pages if 7 or fewer
		pages := make([]int, totalPages)
		for i := 0; i < totalPages; i++ {
			pages[i] = i + 1
		}
		return pages
	}

	// Always show first page, last page, current page, and pages around current
	var pages []int

	// First page
	pages = append(pages, 1)

	// Pages around current
	start := currentPage - 1
	end := currentPage + 1

	if start > 2 {
		// Add ellipsis before current range
		pages = append(pages, -1)
	}

	// Add pages around current (but not before page 2 or after totalPages-1)
	for i := max(2, start); i <= min(totalPages-1, end); i++ {
		pages = append(pages, i)
	}

	if end < totalPages-1 {
		// Add ellipsis after current range
		pages = append(pages, -1)
	}

	// Last page (only if not already added)
	if totalPages > 1 && pages[len(pages)-1] != totalPages {
		pages = append(pages, totalPages)
	}

	return pages
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
