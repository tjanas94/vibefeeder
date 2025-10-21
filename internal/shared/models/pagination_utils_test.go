package models

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for BuildPageURL

// TestBuildPageURL_SimpleURL tests building a page URL with a simple base URL
func TestBuildPageURL_SimpleURL(t *testing.T) {
	baseURL := "https://example.com/feeds"
	result := BuildPageURL(baseURL, 2)

	u, err := url.Parse(result)
	require.NoError(t, err)
	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "example.com", u.Host)
	assert.Equal(t, "/feeds", u.Path)
	assert.Equal(t, "page=2", u.RawQuery)
}

// TestBuildPageURL_URLWithExistingQuery tests building a page URL when base URL already has query parameters
func TestBuildPageURL_URLWithExistingQuery(t *testing.T) {
	baseURL := "https://example.com/feeds?search=tech&status=active"
	result := BuildPageURL(baseURL, 3)

	u, err := url.Parse(result)
	require.NoError(t, err)
	query := u.Query()
	assert.Equal(t, "3", query.Get("page"))
	assert.Equal(t, "tech", query.Get("search"))
	assert.Equal(t, "active", query.Get("status"))
}

// TestBuildPageURL_ReplacesExistingPageParam tests that existing page parameter is replaced
func TestBuildPageURL_ReplacesExistingPageParam(t *testing.T) {
	baseURL := "https://example.com/feeds?page=1&search=news"
	result := BuildPageURL(baseURL, 5)

	u, err := url.Parse(result)
	require.NoError(t, err)
	query := u.Query()
	assert.Equal(t, "5", query.Get("page"))
	assert.Equal(t, "news", query.Get("search"))
}

// TestBuildPageURL_FirstPage tests building URL for first page
func TestBuildPageURL_FirstPage(t *testing.T) {
	baseURL := "https://example.com/feeds"
	result := BuildPageURL(baseURL, 1)

	u, err := url.Parse(result)
	require.NoError(t, err)
	assert.Equal(t, "page=1", u.RawQuery)
}

// TestBuildPageURL_LargePage tests building URL for a large page number
func TestBuildPageURL_LargePage(t *testing.T) {
	baseURL := "https://example.com/feeds"
	result := BuildPageURL(baseURL, 999)

	u, err := url.Parse(result)
	require.NoError(t, err)
	assert.Equal(t, "page=999", u.RawQuery)
}

// TestBuildPageURL_InvalidBaseURL tests fallback when base URL is invalid
func TestBuildPageURL_InvalidBaseURL(t *testing.T) {
	baseURL := "ht!tp://[invalid"
	result := BuildPageURL(baseURL, 2)

	// Should return the base URL unchanged when parsing fails
	assert.Equal(t, baseURL, result)
}

// TestBuildPageURL_URLWithFragment tests URL with fragment (hash)
func TestBuildPageURL_URLWithFragment(t *testing.T) {
	baseURL := "https://example.com/feeds#section1"
	result := BuildPageURL(baseURL, 2)

	u, err := url.Parse(result)
	require.NoError(t, err)
	assert.Equal(t, "page=2", u.RawQuery)
	assert.Equal(t, "section1", u.Fragment)
}

// TestBuildPageURL_SpecialCharactersInQuery tests URL encoding of special characters
func TestBuildPageURL_SpecialCharactersInQuery(t *testing.T) {
	baseURL := "https://example.com/feeds?search=hello%20world&tag=tech%2Fai"
	result := BuildPageURL(baseURL, 2)

	u, err := url.Parse(result)
	require.NoError(t, err)
	query := u.Query()
	assert.Equal(t, "hello world", query.Get("search"))
	assert.Equal(t, "tech/ai", query.Get("tag"))
	assert.Equal(t, "2", query.Get("page"))
}

// TestBuildPageURL_EmptyBaseURL tests with empty base URL
func TestBuildPageURL_EmptyBaseURL(t *testing.T) {
	baseURL := ""
	result := BuildPageURL(baseURL, 2)

	// URL parsing on empty string results in a relative URL with just query string
	u, err := url.Parse(result)
	require.NoError(t, err)
	assert.Equal(t, "page=2", u.RawQuery)
}

// TestBuildPageURL_RelativeURL tests with relative URL
func TestBuildPageURL_RelativeURL(t *testing.T) {
	baseURL := "/feeds"
	result := BuildPageURL(baseURL, 2)

	u, err := url.Parse(result)
	require.NoError(t, err)
	assert.Equal(t, "/feeds", u.Path)
	assert.Equal(t, "page=2", u.RawQuery)
}

// Tests for GetPaginationRange

// TestGetPaginationRange_LessThanSevenPages tests with fewer than 7 total pages
func TestGetPaginationRange_LessThanSevenPages(t *testing.T) {
	result := GetPaginationRange(2, 5)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

// TestGetPaginationRange_ExactlySevenPages tests with exactly 7 total pages
func TestGetPaginationRange_ExactlySevenPages(t *testing.T) {
	result := GetPaginationRange(4, 7)

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, result)
}

// TestGetPaginationRange_FirstPageLargeTotals tests first page with many total pages
func TestGetPaginationRange_FirstPageLargeTotals(t *testing.T) {
	result := GetPaginationRange(1, 20)

	// Should show: 1, -1, 1-2 context around 1, -1, 20
	// But since we're on page 1, it should be: 1, 2, -1, 20
	expected := []int{1, 2, -1, 20}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_MiddlePageLargeTotals tests middle page with many total pages
func TestGetPaginationRange_MiddlePageLargeTotals(t *testing.T) {
	result := GetPaginationRange(10, 20)

	// Should show: 1, -1, 9-11, -1, 20
	expected := []int{1, -1, 9, 10, 11, -1, 20}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_LastPageLargeTotals tests last page with many total pages
func TestGetPaginationRange_LastPageLargeTotals(t *testing.T) {
	result := GetPaginationRange(20, 20)

	// Should show: 1, -1, 19-20, (no final ellipsis needed)
	expected := []int{1, -1, 19, 20}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_SecondPageLargeTotals tests second page with many total pages
func TestGetPaginationRange_SecondPageLargeTotals(t *testing.T) {
	result := GetPaginationRange(2, 20)

	// Should show: 1, 2, 3, -1, 20 (no ellipsis before 2 since we start at page 2)
	expected := []int{1, 2, 3, -1, 20}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_SecondToLastPageLargeTotals tests second-to-last page with many total pages
func TestGetPaginationRange_SecondToLastPageLargeTotals(t *testing.T) {
	result := GetPaginationRange(19, 20)

	// Should show: 1, -1, 18, 19, 20 (no ellipsis after 19 since 20 is last page)
	expected := []int{1, -1, 18, 19, 20}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_SinglePage tests with only one page
func TestGetPaginationRange_SinglePage(t *testing.T) {
	result := GetPaginationRange(1, 1)

	assert.Equal(t, []int{1}, result)
}

// TestGetPaginationRange_TwoPages tests with two pages
func TestGetPaginationRange_TwoPages(t *testing.T) {
	result := GetPaginationRange(1, 2)

	assert.Equal(t, []int{1, 2}, result)
}

// TestGetPaginationRange_EightPages tests with eight pages (boundary condition)
func TestGetPaginationRange_EightPages(t *testing.T) {
	result := GetPaginationRange(4, 8)

	// With 8 pages, we show all since > 7
	// Current is 4, so range is 3-5
	// Pattern: 1, -1, 3, 4, 5, -1, 8
	expected := []int{1, -1, 3, 4, 5, -1, 8}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_EdgeCaseStartBoundary tests edge case at start boundary
func TestGetPaginationRange_EdgeCaseStartBoundary(t *testing.T) {
	result := GetPaginationRange(2, 10)

	// Page 2 of 10
	// Pages around current (2): 1, 2, 3
	// No ellipsis before since next range starts at 2
	// Pattern: 1, 2, 3, -1, 10
	expected := []int{1, 2, 3, -1, 10}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_EdgeCaseEndBoundary tests edge case at end boundary
func TestGetPaginationRange_EdgeCaseEndBoundary(t *testing.T) {
	result := GetPaginationRange(9, 10)

	// Page 9 of 10
	// Pages around current (9): 8, 9, 10
	// No ellipsis after since last page is 10
	// Pattern: 1, -1, 8, 9, 10
	expected := []int{1, -1, 8, 9, 10}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_PageFiveOfTwenty tests example from documentation
func TestGetPaginationRange_PageFiveOfTwenty(t *testing.T) {
	result := GetPaginationRange(5, 20)

	// Page 5 of 20
	// Pages around current (5): 4, 5, 6
	// Pattern: 1, -1, 4, 5, 6, -1, 20
	expected := []int{1, -1, 4, 5, 6, -1, 20}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_LargePageNumber tests with a very large page number
func TestGetPaginationRange_LargePageNumber(t *testing.T) {
	result := GetPaginationRange(50, 100)

	// Page 50 of 100
	// Pages around current (50): 49, 50, 51
	// Pattern: 1, -1, 49, 50, 51, -1, 100
	expected := []int{1, -1, 49, 50, 51, -1, 100}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_ThreePages tests with three pages
func TestGetPaginationRange_ThreePages(t *testing.T) {
	result := GetPaginationRange(2, 3)

	assert.Equal(t, []int{1, 2, 3}, result)
}

// TestGetPaginationRange_CurrentPageThree tests page 3 of 10
func TestGetPaginationRange_CurrentPageThree(t *testing.T) {
	result := GetPaginationRange(3, 10)

	// Page 3 of 10
	// Pages around current (3): 2, 3, 4
	// Pattern: 1, 2, 3, 4, -1, 10
	expected := []int{1, 2, 3, 4, -1, 10}
	assert.Equal(t, expected, result)
}

// TestGetPaginationRange_InvalidCurrentPageZero tests with current page = 0 (edge case)
func TestGetPaginationRange_InvalidCurrentPageZero(t *testing.T) {
	result := GetPaginationRange(0, 10)

	// Even with invalid current page, the function should handle it gracefully
	// Pages around 0: -1, 0, 1
	// But filtered to range 2 to 9: 2 onwards
	// Pattern: 1, 2, 3, -1, 10
	assert.NotEmpty(t, result)
	assert.Contains(t, result, 1)
	assert.Contains(t, result, 10)
}

// TestGetPaginationRange_InvalidCurrentPageNegative tests with negative current page
func TestGetPaginationRange_InvalidCurrentPageNegative(t *testing.T) {
	result := GetPaginationRange(-5, 10)

	// Should still produce a valid range
	assert.NotEmpty(t, result)
	assert.Contains(t, result, 1)
	assert.Contains(t, result, 10)
}
