package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildPagination_FirstPage tests pagination for the first page
func TestBuildPagination_FirstPage(t *testing.T) {
	result := BuildPagination(100, 1, 20)

	assert.Equal(t, 1, result.CurrentPage)
	assert.Equal(t, 5, result.TotalPages)
	assert.Equal(t, 100, result.TotalItems)
	assert.False(t, result.HasPrevious)
	assert.True(t, result.HasNext)
}

// TestBuildPagination_MiddlePage tests pagination for a middle page
func TestBuildPagination_MiddlePage(t *testing.T) {
	result := BuildPagination(100, 3, 20)

	assert.Equal(t, 3, result.CurrentPage)
	assert.Equal(t, 5, result.TotalPages)
	assert.Equal(t, 100, result.TotalItems)
	assert.True(t, result.HasPrevious)
	assert.True(t, result.HasNext)
}

// TestBuildPagination_LastPage tests pagination for the last page
func TestBuildPagination_LastPage(t *testing.T) {
	result := BuildPagination(100, 5, 20)

	assert.Equal(t, 5, result.CurrentPage)
	assert.Equal(t, 5, result.TotalPages)
	assert.Equal(t, 100, result.TotalItems)
	assert.True(t, result.HasPrevious)
	assert.False(t, result.HasNext)
}

// TestBuildPagination_SinglePage tests pagination when all items fit on one page
func TestBuildPagination_SinglePage(t *testing.T) {
	result := BuildPagination(15, 1, 20)

	assert.Equal(t, 1, result.CurrentPage)
	assert.Equal(t, 1, result.TotalPages)
	assert.Equal(t, 15, result.TotalItems)
	assert.False(t, result.HasPrevious)
	assert.False(t, result.HasNext)
}

// TestBuildPagination_ZeroItems tests pagination with no items
func TestBuildPagination_ZeroItems(t *testing.T) {
	result := BuildPagination(0, 1, 20)

	assert.Equal(t, 1, result.CurrentPage)
	assert.Equal(t, 1, result.TotalPages) // Should be at least 1
	assert.Equal(t, 0, result.TotalItems)
	assert.False(t, result.HasPrevious)
	assert.False(t, result.HasNext)
}

// TestBuildPagination_ExactPageBoundary tests when total items exactly match page boundary
func TestBuildPagination_ExactPageBoundary(t *testing.T) {
	result := BuildPagination(60, 3, 20)

	assert.Equal(t, 3, result.CurrentPage)
	assert.Equal(t, 3, result.TotalPages)
	assert.Equal(t, 60, result.TotalItems)
	assert.True(t, result.HasPrevious)
	assert.False(t, result.HasNext)
}

// TestBuildPagination_OneItemPerPage tests with minimum page size
func TestBuildPagination_OneItemPerPage(t *testing.T) {
	result := BuildPagination(50, 25, 1)

	assert.Equal(t, 25, result.CurrentPage)
	assert.Equal(t, 50, result.TotalPages)
	assert.Equal(t, 50, result.TotalItems)
	assert.True(t, result.HasPrevious)
	assert.True(t, result.HasNext)
}

// TestBuildPagination_LargePageSize tests with large page size
func TestBuildPagination_LargePageSize(t *testing.T) {
	result := BuildPagination(100, 1, 1000)

	assert.Equal(t, 1, result.CurrentPage)
	assert.Equal(t, 1, result.TotalPages)
	assert.Equal(t, 100, result.TotalItems)
	assert.False(t, result.HasPrevious)
	assert.False(t, result.HasNext)
}

// TestBuildPagination_RoundingUp tests that total pages rounds up correctly
func TestBuildPagination_RoundingUp(t *testing.T) {
	// 25 items with page size 20 should result in 2 pages
	result := BuildPagination(25, 1, 20)

	assert.Equal(t, 2, result.TotalPages)
}

// TestBuildPagination_InvalidPage_BeyondTotal tests querying a page beyond total pages
func TestBuildPagination_InvalidPage_BeyondTotal(t *testing.T) {
	// Page 10 when only 2 pages exist
	result := BuildPagination(30, 10, 20)

	assert.Equal(t, 10, result.CurrentPage)
	assert.Equal(t, 2, result.TotalPages)
	assert.True(t, result.HasPrevious)
	assert.False(t, result.HasNext) // No next page when current > total
}

// TestBuildPagination_StandardPagination tests typical use case
func TestBuildPagination_StandardPagination(t *testing.T) {
	result := BuildPagination(247, 2, 25)

	assert.Equal(t, 2, result.CurrentPage)
	assert.Equal(t, 10, result.TotalPages) // (247 + 25 - 1) / 25 = 10
	assert.Equal(t, 247, result.TotalItems)
	assert.True(t, result.HasPrevious)
	assert.True(t, result.HasNext)
}
