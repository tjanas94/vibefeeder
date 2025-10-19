package summary

import (
	"errors"

	"github.com/tjanas94/vibefeeder/internal/shared/ai"
)

// extractSummaryContent safely extracts content from AI response.
// Returns error if the response has no choices or empty content.
func extractSummaryContent(response *ai.ChatCompletionResponse) (string, error) {
	if response == nil {
		return "", errors.New("AI response is nil")
	}

	if len(response.Choices) == 0 {
		return "", errors.New("AI response has no choices")
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("AI response content is empty")
	}

	return content, nil
}
