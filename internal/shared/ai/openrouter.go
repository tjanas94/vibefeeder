package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tjanas94/vibefeeder/internal/shared/config"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// OpenRouterService handles communication with OpenRouter API
type OpenRouterService struct {
	apiKey     string
	httpClient HTTPClient
	baseURL    string
}

// GenerateChatCompletionOptions defines parameters for chat completion request
type GenerateChatCompletionOptions struct {
	Model          string
	SystemPrompt   string
	UserPrompt     string
	ResponseFormat *ResponseFormat
	Temperature    float64
	MaxTokens      int
}

// ResponseFormat defines the format of the response, e.g., JSON schema
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema defines the details of JSON schema
type JSONSchema struct {
	Name   string `json:"name"`
	Strict bool   `json:"strict"`
	Schema any    `json:"schema"` // Can be map[string]any or struct
}

// ChatCompletionResponse represents the full response from the API
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Model   string   `json:"model"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Choice represents a single choice/response in the API response
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// chatCompletionRequest represents the request body for chat completion
type chatCompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// errorResponse represents an error response from the API
type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenRouterService creates a new instance of OpenRouter service.
// Requires OpenRouter configuration and HTTP client.
func NewOpenRouterService(cfg config.OpenRouterConfig, httpClient HTTPClient) (*OpenRouterService, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("HTTP client is required")
	}

	return &OpenRouterService{
		apiKey:     cfg.APIKey,
		httpClient: httpClient,
		baseURL:    openRouterBaseURL,
	}, nil
}

// GenerateChatCompletion sends a chat completion request to OpenRouter API.
// Returns the parsed response or an error.
func (s *OpenRouterService) GenerateChatCompletion(ctx context.Context, options GenerateChatCompletionOptions) (*ChatCompletionResponse, error) {
	// Validate input
	if options.UserPrompt == "" {
		return nil, fmt.Errorf("user prompt is required")
	}

	if options.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	// Build request
	req, err := s.buildRequest(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Parse response
	return s.parseResponse(resp)
}

// buildRequest creates an HTTP request from the given options
func (s *OpenRouterService) buildRequest(ctx context.Context, options GenerateChatCompletionOptions) (*http.Request, error) {
	// Build messages array
	messages := make([]ChatMessage, 0, 2)

	// Add system prompt if provided
	if options.SystemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: options.SystemPrompt,
		})
	}

	// Add user prompt
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: options.UserPrompt,
	})

	// Create request body
	requestBody := chatCompletionRequest{
		Model:       options.Model,
		Messages:    messages,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
	}

	// Handle response format if provided
	if options.ResponseFormat != nil && options.ResponseFormat.JSONSchema != nil {
		// Convert schema to map if it's a struct
		var schemaPayload any
		if options.ResponseFormat.JSONSchema.Schema != nil {
			schemaBytes, err := json.Marshal(options.ResponseFormat.JSONSchema.Schema)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json schema: %w", err)
			}

			var schemaMap map[string]any
			if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal json schema to map: %w", err)
			}
			schemaPayload = schemaMap
		}

		requestBody.ResponseFormat = &ResponseFormat{
			Type: options.ResponseFormat.Type,
			JSONSchema: &JSONSchema{
				Name:   options.ResponseFormat.JSONSchema.Name,
				Strict: options.ResponseFormat.JSONSchema.Strict,
				Schema: schemaPayload,
			},
		}
	}

	// Serialize request body to JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := s.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// parseResponse parses the HTTP response into ChatCompletionResponse
func (s *OpenRouterService) parseResponse(resp *http.Response) (*ChatCompletionResponse, error) {
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
			// If we can't parse the error response, return the raw body
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// Return specific error based on status code
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("authentication failed: invalid API key")
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("rate limit exceeded: %s", errResp.Error.Message)
		case http.StatusBadRequest:
			return nil, fmt.Errorf("bad request: %s", errResp.Error.Message)
		default:
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Error.Message)
		}
	}

	// Parse successful response
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(bodyBytes))
	}

	// Validate response has choices
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("API returned empty choices array")
	}

	return &chatResp, nil
}
