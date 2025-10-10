package summary

import (
	"fmt"
	"strings"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// buildPromptFromArticles creates a prompt for the AI from article data.
// This is a pure function with no side effects.
func buildPromptFromArticles(articles []database.PublicArticlesSelect) string {
	var sb strings.Builder

	sb.WriteString("Please generate a concise summary of the following articles:\n\n")

	for i, article := range articles {
		sb.WriteString(fmt.Sprintf("Article %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("Title: %s\n", article.Title))

		if article.Content != nil && *article.Content != "" {
			sb.WriteString(fmt.Sprintf("Content: %s\n", *article.Content))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("\nProvide a summary that highlights the main themes and key insights from these articles.")

	return sb.String()
}
