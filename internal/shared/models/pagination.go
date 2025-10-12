package models

// PaginationViewModel represents pagination information for paginated lists.
// Can be used across different features (feeds, summaries, users, etc.)
type PaginationViewModel struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasPrevious bool `json:"has_previous"`
	HasNext     bool `json:"has_next"`
}

// BuildPagination creates pagination view model from total count and query parameters.
// Calculates total pages, determines if there are previous/next pages.
//
// Parameters:
//   - totalCount: Total number of items across all pages
//   - currentPage: Current page number (1-indexed)
//   - pageSize: Number of items per page
//
// Returns:
//   - PaginationViewModel with calculated pagination metadata
func BuildPagination(totalCount, currentPage, pageSize int) PaginationViewModel {
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationViewModel{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalItems:  totalCount,
		HasPrevious: currentPage > 1,
		HasNext:     currentPage < totalPages,
	}
}
