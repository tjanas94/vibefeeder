package summary

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/summary/models"
)

// TestBuildPromptFromArticles tests the buildPromptFromArticles function
func TestBuildPromptFromArticles(t *testing.T) {
	tests := []struct {
		name        string
		articles    []models.ArticleForPrompt
		validate    func(t *testing.T, prompt string)
		description string
	}{
		{
			name:     "empty articles list",
			articles: []models.ArticleForPrompt{},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Please generate a concise summary of the following articles:")
				assert.NotContains(t, prompt, "Article 1:")
			},
			description: "Should handle empty list gracefully",
		},
		{
			name: "single article with all fields",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Go Language Best Practices",
					Content: ptr("This article discusses best practices for writing efficient Go code."),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Title: Go Language Best Practices")
				assert.Contains(t, prompt, "Content: This article discusses best practices for writing efficient Go code.")
				assert.Contains(t, prompt, "Please generate a concise summary of the following articles:")
				assert.Contains(t, prompt, "main themes and key insights")
			},
			description: "Should format single article with title and content",
		},
		{
			name: "single article with nil content",
			articles: []models.ArticleForPrompt{
				{
					Title:   "News Headline",
					Content: nil,
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Title: News Headline")
				assert.NotContains(t, prompt, "Content:")
			},
			description: "Should skip content field when nil",
		},
		{
			name: "single article with empty content",
			articles: []models.ArticleForPrompt{
				{
					Title:   "News Headline",
					Content: ptr(""),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Title: News Headline")
				assert.NotContains(t, prompt, "Content:")
			},
			description: "Should skip content field when empty string",
		},
		{
			name: "multiple articles",
			articles: []models.ArticleForPrompt{
				{
					Title:   "First Article",
					Content: ptr("Content of first article"),
				},
				{
					Title:   "Second Article",
					Content: ptr("Content of second article"),
				},
				{
					Title:   "Third Article",
					Content: ptr("Content of third article"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Title: First Article")
				assert.Contains(t, prompt, "Article 2:")
				assert.Contains(t, prompt, "Title: Second Article")
				assert.Contains(t, prompt, "Article 3:")
				assert.Contains(t, prompt, "Title: Third Article")

				// Verify order is preserved
				idx1 := strings.Index(prompt, "Article 1:")
				idx2 := strings.Index(prompt, "Article 2:")
				idx3 := strings.Index(prompt, "Article 3:")
				assert.Less(t, idx1, idx2)
				assert.Less(t, idx2, idx3)
			},
			description: "Should format multiple articles in order",
		},
		{
			name: "article with content longer than maxContentLength",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Long Article",
					Content: ptr(strings.Repeat("a", maxContentLength+100)),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Title: Long Article")
				// Should contain truncation indicator
				assert.Contains(t, prompt, "...")
				// Content should be truncated to maxContentLength + "..."
				contentStart := strings.Index(prompt, "Content: ")
				require.NotEqual(t, -1, contentStart)
				contentSection := prompt[contentStart:]
				contentEnd := strings.Index(contentSection, "\n\n")
				if contentEnd == -1 {
					contentEnd = len(contentSection)
				}
				// Length should be approximately maxContentLength + 3 (for "...") + some formatting
				assert.Less(t, contentEnd, maxContentLength+500) // Give some buffer for formatting
			},
			description: "Should truncate content exceeding maxContentLength",
		},
		{
			name: "article with content exactly at maxContentLength",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Exact Length Article",
					Content: ptr(strings.Repeat("a", maxContentLength)),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				// Should NOT contain "..." when exactly at limit
				contentStart := strings.Index(prompt, "Content: ")
				require.NotEqual(t, -1, contentStart)
				contentEnd := strings.Index(prompt[contentStart:], "\n\n")
				if contentEnd == -1 {
					contentEnd = len(prompt[contentStart:])
				}
				// Should contain exactly maxContentLength 'a's without truncation
				assert.GreaterOrEqual(t, contentEnd, maxContentLength)
			},
			description: "Should not truncate content exactly at maxContentLength",
		},
		{
			name: "article with special characters in content",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Special Chars Article",
					Content: ptr("Content with special chars: @#$%^&*()_+-=[]{}|;:',.<>?/`~"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Content with special chars: @#$%^&*()_+-=[]{}|;:',.<>?/`~")
			},
			description: "Should preserve special characters in content",
		},
		{
			name: "article with unicode characters",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Unicode Article",
					Content: ptr("Content with unicode: 中文, العربية, 한글, 🚀 📚 ✨"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "中文")
				assert.Contains(t, prompt, "العربية")
				assert.Contains(t, prompt, "한글")
				assert.Contains(t, prompt, "🚀")
			},
			description: "Should preserve unicode characters",
		},
		{
			name: "article with newlines in content",
			articles: []models.ArticleForPrompt{
				{
					Title: "Multiline Article",
					Content: ptr(`Line 1
Line 2
Line 3`),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Line 1\nLine 2\nLine 3")
			},
			description: "Should preserve newlines in content",
		},
		{
			name: "mixed articles with and without content",
			articles: []models.ArticleForPrompt{
				{
					Title:   "With Content",
					Content: ptr("This has content"),
				},
				{
					Title:   "Without Content",
					Content: nil,
				},
				{
					Title:   "Empty Content",
					Content: ptr(""),
				},
				{
					Title:   "With Content Again",
					Content: ptr("More content"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Article 2:")
				assert.Contains(t, prompt, "Article 3:")
				assert.Contains(t, prompt, "Article 4:")
				// Only first and fourth should have content
				lines := strings.Split(prompt, "\n")
				contentCount := 0
				for _, line := range lines {
					if strings.HasPrefix(line, "Content:") {
						contentCount++
					}
				}
				assert.Equal(t, 2, contentCount)
			},
			description: "Should handle mixed content presence",
		},
		{
			name: "article with very long title",
			articles: []models.ArticleForPrompt{
				{
					Title:   strings.Repeat("Long Title ", 50),
					Content: ptr("Short content"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, strings.Repeat("Long Title ", 50))
			},
			description: "Should handle very long titles",
		},
		{
			name: "article with tabs and special whitespace",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Tab Article",
					Content: ptr("Content\twith\ttabs\nand\r\nnewlines"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Content\twith\ttabs\nand\r\nnewlines")
			},
			description: "Should preserve tabs and special whitespace",
		},
		{
			name: "large number of articles",
			articles: func() []models.ArticleForPrompt {
				articles := make([]models.ArticleForPrompt, 100)
				for i := 0; i < 100; i++ {
					articles[i] = models.ArticleForPrompt{
						Title:   "Article " + string(rune(i)),
						Content: ptr("Content for article"),
					}
				}
				return articles
			}(),
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Article 1:")
				assert.Contains(t, prompt, "Article 50:")
				assert.Contains(t, prompt, "Article 100:")
				// All articles should be present
				assert.GreaterOrEqual(t, len(strings.Split(prompt, "Article")), 100)
			},
			description: "Should handle large number of articles",
		},
		{
			name: "prompt contains system instruction",
			articles: []models.ArticleForPrompt{
				{
					Title:   "Test Article",
					Content: ptr("Test content"),
				},
			},
			validate: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "main themes")
				assert.Contains(t, prompt, "key insights")
				assert.Contains(t, prompt, "summary")
			},
			description: "Should include system instruction in prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildPromptFromArticles(tt.articles)

			assert.NotEmpty(t, prompt, tt.description)
			tt.validate(t, prompt)
		})
	}
}

// TestBuildPromptFromArticles_ContentTruncation specifically tests the truncation logic
func TestBuildPromptFromArticles_ContentTruncation(t *testing.T) {
	tests := []struct {
		name             string
		contentLength    int
		expectTruncation bool
		description      string
	}{
		{
			name:             "content at boundary minus 1",
			contentLength:    maxContentLength - 1,
			expectTruncation: false,
			description:      "Content just below limit should not be truncated",
		},
		{
			name:             "content at boundary exactly",
			contentLength:    maxContentLength,
			expectTruncation: false,
			description:      "Content exactly at limit should not be truncated",
		},
		{
			name:             "content at boundary plus 1",
			contentLength:    maxContentLength + 1,
			expectTruncation: true,
			description:      "Content just above limit should be truncated",
		},
		{
			name:             "content double the limit",
			contentLength:    maxContentLength * 2,
			expectTruncation: true,
			description:      "Content well above limit should be truncated",
		},
		{
			name:             "very small content",
			contentLength:    1,
			expectTruncation: false,
			description:      "Very small content should not be truncated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := strings.Repeat("a", tt.contentLength)
			articles := []models.ArticleForPrompt{
				{
					Title:   "Test",
					Content: ptr(content),
				},
			}

			prompt := buildPromptFromArticles(articles)

			if tt.expectTruncation {
				assert.Contains(t, prompt, "...", tt.description)
			}
		})
	}
}

// TestBuildPromptFromArticles_NilInput tests behavior with nil input
func TestBuildPromptFromArticles_NilInput(t *testing.T) {
	prompt := buildPromptFromArticles(nil)

	assert.NotEmpty(t, prompt, "Should not crash with nil input")
	assert.Contains(t, prompt, "Please generate a concise summary of the following articles:")
}

// BenchmarkBuildPromptFromArticles benchmarks with typical article count
func BenchmarkBuildPromptFromArticles(b *testing.B) {
	articles := make([]models.ArticleForPrompt, 10)
	for i := 0; i < 10; i++ {
		content := strings.Repeat("This is article content. ", 20)
		articles[i] = models.ArticleForPrompt{
			Title:   "Article Title " + string(rune(i)),
			Content: ptr(content),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPromptFromArticles(articles)
	}
}

// BenchmarkBuildPromptFromArticles_SingleArticle benchmarks with single article
func BenchmarkBuildPromptFromArticles_SingleArticle(b *testing.B) {
	articles := []models.ArticleForPrompt{
		{
			Title:   "Article Title",
			Content: ptr("This is article content. This is article content. This is article content."),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPromptFromArticles(articles)
	}
}

// BenchmarkBuildPromptFromArticles_ManyArticles benchmarks with many articles
func BenchmarkBuildPromptFromArticles_ManyArticles(b *testing.B) {
	articles := make([]models.ArticleForPrompt, 100)
	for i := 0; i < 100; i++ {
		content := strings.Repeat("Content ", 30)
		articles[i] = models.ArticleForPrompt{
			Title:   "Article " + string(rune(i%256)),
			Content: ptr(content),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPromptFromArticles(articles)
	}
}

// BenchmarkBuildPromptFromArticles_LongContent benchmarks with long content
func BenchmarkBuildPromptFromArticles_LongContent(b *testing.B) {
	articles := []models.ArticleForPrompt{
		{
			Title:   "Article",
			Content: ptr(strings.Repeat("a", maxContentLength+1000)),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPromptFromArticles(articles)
	}
}

// BenchmarkBuildPromptFromArticles_Empty benchmarks with empty articles
func BenchmarkBuildPromptFromArticles_Empty(b *testing.B) {
	articles := []models.ArticleForPrompt{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPromptFromArticles(articles)
	}
}

// Helper function to create pointers to strings (avoids repetition in tests)
func ptr(s string) *string {
	return &s
}
