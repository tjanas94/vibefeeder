package summary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/shared/ai"
)

// TestExtractSummaryContent tests the extractSummaryContent helper function
func TestExtractSummaryContent(t *testing.T) {
	tests := []struct {
		name        string
		response    *ai.ChatCompletionResponse
		wantContent string
		wantErr     bool
		errMsg      string
		description string
	}{
		{
			name: "valid response with content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "This is a summary of the articles.",
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "This is a summary of the articles.",
			wantErr:     false,
			description: "Should extract content from valid response",
		},
		{
			name: "response with multiple choices returns first",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "First summary",
						},
						FinishReason: "stop",
					},
					{
						Index: 1,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "Second summary",
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "First summary",
			wantErr:     false,
			description: "Should return content from first choice only",
		},
		{
			name:        "nil response returns error",
			response:    nil,
			wantContent: "",
			wantErr:     true,
			errMsg:      "AI response is nil",
			description: "Should error when response is nil",
		},
		{
			name: "response with empty choices array",
			response: &ai.ChatCompletionResponse{
				ID:      "resp-123",
				Choices: []ai.Choice{},
				Model:   "gpt-4",
			},
			wantContent: "",
			wantErr:     true,
			errMsg:      "AI response has no choices",
			description: "Should error when choices array is empty",
		},
		{
			name: "response with empty content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "",
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "",
			wantErr:     true,
			errMsg:      "AI response content is empty",
			description: "Should error when content is empty string",
		},
		{
			name: "response with whitespace-only content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "   ",
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "   ",
			wantErr:     false,
			description: "Should accept whitespace-only content (not our responsibility to validate)",
		},
		{
			name: "response with very long content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "This is a very long summary. " + string(make([]byte, 10000)),
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "This is a very long summary. " + string(make([]byte, 10000)),
			wantErr:     false,
			description: "Should handle very long content",
		},
		{
			name: "response with special characters in content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "Summary with special chars: @#$%^&*()_+-=[]{}|;:',.<>?/`~\n\t",
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "Summary with special chars: @#$%^&*()_+-=[]{}|;:',.<>?/`~\n\t",
			wantErr:     false,
			description: "Should handle special characters and escape sequences",
		},
		{
			name: "response with unicode content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "Summary with emojis: 🚀 📚 ✨ and Chinese: 中文",
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "Summary with emojis: 🚀 📚 ✨ and Chinese: 中文",
			wantErr:     false,
			description: "Should handle unicode characters properly",
		},
		{
			name: "response with multiline content",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role: "assistant",
							Content: `Summary:
- Point 1
- Point 2
- Point 3`,
						},
						FinishReason: "stop",
					},
				},
				Model: "gpt-4",
			},
			wantContent: `Summary:
- Point 1
- Point 2
- Point 3`,
			wantErr:     false,
			description: "Should preserve multiline content with newlines",
		},
		{
			name: "response choice with different finish reason",
			response: &ai.ChatCompletionResponse{
				ID: "resp-123",
				Choices: []ai.Choice{
					{
						Index: 0,
						Message: ai.ChatMessage{
							Role:    "assistant",
							Content: "Incomplete summary due to token limit",
						},
						FinishReason: "length",
					},
				},
				Model: "gpt-4",
			},
			wantContent: "Incomplete summary due to token limit",
			wantErr:     false,
			description: "Should extract content regardless of finish reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := extractSummaryContent(tt.response)

			if tt.wantErr {
				require.Error(t, err, tt.description)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantContent, content)
			}
		})
	}
}

// TestExtractSummaryContent_EdgeCases focuses on edge cases and boundary conditions
func TestExtractSummaryContent_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		response    *ai.ChatCompletionResponse
		wantContent string
		wantErr     bool
		description string
	}{
		{
			name: "response with single character content",
			response: &ai.ChatCompletionResponse{
				Choices: []ai.Choice{
					{
						Message: ai.ChatMessage{Content: "A"},
					},
				},
			},
			wantContent: "A",
			wantErr:     false,
			description: "Should handle single character content",
		},
		{
			name: "response with newline as content",
			response: &ai.ChatCompletionResponse{
				Choices: []ai.Choice{
					{
						Message: ai.ChatMessage{Content: "\n"},
					},
				},
			},
			wantContent: "\n",
			wantErr:     false,
			description: "Should accept newline as valid content",
		},
		{
			name: "response with tab characters",
			response: &ai.ChatCompletionResponse{
				Choices: []ai.Choice{
					{
						Message: ai.ChatMessage{Content: "\t\t\tindented content"},
					},
				},
			},
			wantContent: "\t\t\tindented content",
			wantErr:     false,
			description: "Should preserve tab characters",
		},
		{
			name: "response with zero-width characters",
			response: &ai.ChatCompletionResponse{
				Choices: []ai.Choice{
					{
						Message: ai.ChatMessage{Content: "content\u200bwith\u200czero\u200bwidth"},
					},
				},
			},
			wantContent: "content\u200bwith\u200czero\u200bwidth",
			wantErr:     false,
			description: "Should preserve zero-width characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := extractSummaryContent(tt.response)

			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantContent, content)
			}
		})
	}
}

// BenchmarkExtractSummaryContent benchmarks the extractSummaryContent function
func BenchmarkExtractSummaryContent(b *testing.B) {
	response := &ai.ChatCompletionResponse{
		ID: "resp-123",
		Choices: []ai.Choice{
			{
				Index: 0,
				Message: ai.ChatMessage{
					Role:    "assistant",
					Content: "This is a sample summary that would be extracted from an AI response.",
				},
				FinishReason: "stop",
			},
		},
		Model: "gpt-4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractSummaryContent(response)
	}
}

// BenchmarkExtractSummaryContent_LongContent benchmarks with longer content
func BenchmarkExtractSummaryContent_LongContent(b *testing.B) {
	longContent := "This is a sample summary. " + string(make([]byte, 50000))

	response := &ai.ChatCompletionResponse{
		ID: "resp-123",
		Choices: []ai.Choice{
			{
				Index: 0,
				Message: ai.ChatMessage{
					Role:    "assistant",
					Content: longContent,
				},
				FinishReason: "stop",
			},
		},
		Model: "gpt-4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractSummaryContent(response)
	}
}

// BenchmarkExtractSummaryContent_Error benchmarks error case (nil response)
func BenchmarkExtractSummaryContent_Error(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractSummaryContent(nil)
	}
}
