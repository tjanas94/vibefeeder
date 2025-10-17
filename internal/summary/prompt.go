package summary

import (
	"fmt"
	"strings"

	"github.com/tjanas94/vibefeeder/internal/summary/models"
)

const (
	systemPrompt = "You are a helpful assistant that creates concise and insightful summaries of news articles and blog posts. Focus on extracting key themes, main points, and actionable insights."
	// maxContentLength limits the number of characters from each article's content
	maxContentLength = 1000
)

// buildPromptFromArticles creates a prompt for the AI from article data.
// This is a pure function with no side effects.
// Article content is truncated to maxContentLength to prevent excessive token usage.
func buildPromptFromArticles(articles []models.ArticleForPrompt) string {
	var sb strings.Builder

	sb.WriteString("Please generate a concise summary of the following articles:\n\n")

	for i, article := range articles {
		sb.WriteString(fmt.Sprintf("Article %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("Title: %s\n", article.Title))

		if article.Content != nil && *article.Content != "" {
			content := *article.Content
			// Truncate content if it exceeds maxContentLength
			if len(content) > maxContentLength {
				content = content[:maxContentLength] + "..."
			}
			sb.WriteString(fmt.Sprintf("Content: %s\n", content))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("\nProvide a summary that highlights the main themes and key insights from these articles.")

	return sb.String()
}
