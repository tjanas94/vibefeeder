package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
)

// MockHTTPClient is a mock implementation of HTTPClient interface
type MockHTTPClient struct {
	mock.Mock
}

// Do mocks the HTTPClient.Do method
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Helper functions for test data
func newTestConfig(apiKey string) config.OpenRouterConfig {
	return config.OpenRouterConfig{
		APIKey: apiKey,
	}
}

func newTestHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func newTestChatCompletionOptions(model, systemPrompt, userPrompt string) GenerateChatCompletionOptions {
	return GenerateChatCompletionOptions{
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    2000,
	}
}

// Tests for NewOpenRouterService
func TestNewOpenRouterService_Success(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	httpClient := &http.Client{}

	service, err := NewOpenRouterService(cfg, httpClient)

	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "test-api-key", service.apiKey)
	assert.Equal(t, openRouterBaseURL, service.baseURL)
	assert.Equal(t, httpClient, service.httpClient)
}

func TestNewOpenRouterService_NilHTTPClient(t *testing.T) {
	cfg := newTestConfig("test-api-key")

	service, err := NewOpenRouterService(cfg, nil)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "HTTP client is required")
}

// Tests for GenerateChatCompletion - Success Cases
func TestGenerateChatCompletion_Success(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"You are a helpful assistant.",
		"Tell me a joke",
	)

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Why did the programmer quit? Because they didn't get arrays."
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 15,
			"total_tokens": 25
		}
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(newTestHTTPResponse(http.StatusOK, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "response-123", result.ID)
	assert.Equal(t, "openai/gpt-4o-mini", result.Model)
	assert.Len(t, result.Choices, 1)
	assert.Equal(t, "assistant", result.Choices[0].Message.Role)
	assert.Contains(t, result.Choices[0].Message.Content, "arrays")
	assert.Equal(t, "stop", result.Choices[0].FinishReason)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 10, result.Usage.PromptTokens)
	assert.Equal(t, 15, result.Usage.CompletionTokens)
	assert.Equal(t, 25, result.Usage.TotalTokens)
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_SuccessWithoutSystemPrompt(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"", // No system prompt
		"Hello",
	)

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hi there!"
			},
			"finish_reason": "stop"
		}]
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusOK, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hi there!", result.Choices[0].Message.Content)
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_SuccessWithJSONSchema(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "number"},
		},
		"required": []string{"name"},
	}

	options := GenerateChatCompletionOptions{
		Model:       "openai/gpt-4o-mini",
		UserPrompt:  "Generate a person",
		Temperature: 0.7,
		MaxTokens:   1000,
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchema{
				Name:   "person",
				Strict: true,
				Schema: schema,
			},
		},
	}

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "{\"name\":\"John\",\"age\":30}"
			},
			"finish_reason": "stop"
		}]
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusOK, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Choices[0].Message.Content, "John")
	mockHTTP.AssertExpectations(t)
}

// Tests for GenerateChatCompletion - Validation Errors
func TestGenerateChatCompletion_EmptyUserPrompt(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := GenerateChatCompletionOptions{
		Model:      "openai/gpt-4o-mini",
		UserPrompt: "",
	}

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user prompt is required")
	mockHTTP.AssertNotCalled(t, "Do")
}

func TestGenerateChatCompletion_EmptyModel(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := GenerateChatCompletionOptions{
		Model:      "",
		UserPrompt: "Hello",
	}

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "model is required")
	mockHTTP.AssertNotCalled(t, "Do")
}

// Tests for GenerateChatCompletion - HTTP Errors
func TestGenerateChatCompletion_HTTPRequestError(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"System",
		"User",
	)

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(nil, errors.New("network error"))

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send request")
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_UnauthorizedError(t *testing.T) {
	cfg := newTestConfig("invalid-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	responseBody := `{
		"error": {
			"message": "Invalid API key",
			"type": "invalid_request_error",
			"code": "invalid_api_key"
		}
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusUnauthorized, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.Contains(t, err.Error(), "invalid API key")
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_RateLimitError(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	responseBody := `{
		"error": {
			"message": "Rate limit exceeded. Please try again later.",
			"type": "rate_limit_error",
			"code": "rate_limit_exceeded"
		}
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusTooManyRequests, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_BadRequestError(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"invalid-model",
		"",
		"Hello",
	)

	responseBody := `{
		"error": {
			"message": "Invalid model specified",
			"type": "invalid_request_error",
			"code": "invalid_model"
		}
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusBadRequest, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bad request")
	assert.Contains(t, err.Error(), "Invalid model specified")
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_InternalServerError(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	responseBody := `{
		"error": {
			"message": "Internal server error",
			"type": "server_error",
			"code": "internal_error"
		}
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusInternalServerError, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "API error")
	assert.Contains(t, err.Error(), "500")
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_InvalidErrorResponse(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	// Non-JSON error response
	responseBody := `This is not a JSON response`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusBadRequest, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "API returned status 400")
	mockHTTP.AssertExpectations(t)
}

// Tests for GenerateChatCompletion - Response Parsing Errors
func TestGenerateChatCompletion_InvalidJSONResponse(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	// Invalid JSON
	responseBody := `{"id": "response-123", "choices": [incomplete`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusOK, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse response")
	mockHTTP.AssertExpectations(t)
}

func TestGenerateChatCompletion_EmptyChoicesArray(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": []
	}`

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(newTestHTTPResponse(http.StatusOK, responseBody), nil)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "empty choices array")
	mockHTTP.AssertExpectations(t)
}

// Tests for buildRequest
func TestBuildRequest_WithSystemPrompt(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"You are helpful",
		"Hello",
	)

	req, err := service.buildRequest(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, openRouterBaseURL+"/chat/completions", req.URL.String())
	assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

	// Verify request body contains both system and user messages
	bodyBytes, _ := io.ReadAll(req.Body)
	bodyStr := string(bodyBytes)
	assert.Contains(t, bodyStr, "system")
	assert.Contains(t, bodyStr, "You are helpful")
	assert.Contains(t, bodyStr, "user")
	assert.Contains(t, bodyStr, "Hello")
}

func TestBuildRequest_WithoutSystemPrompt(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"", // No system prompt
		"Hello",
	)

	req, err := service.buildRequest(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, req)

	// Verify request body contains only user message
	bodyBytes, _ := io.ReadAll(req.Body)
	bodyStr := string(bodyBytes)
	assert.NotContains(t, bodyStr, "system")
	assert.Contains(t, bodyStr, "user")
	assert.Contains(t, bodyStr, "Hello")
}

func TestBuildRequest_WithJSONSchema(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}

	options := GenerateChatCompletionOptions{
		Model:      "openai/gpt-4o-mini",
		UserPrompt: "Test",
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchema{
				Name:   "test_schema",
				Strict: true,
				Schema: schema,
			},
		},
	}

	req, err := service.buildRequest(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, req)

	// Verify request body contains response format
	bodyBytes, _ := io.ReadAll(req.Body)
	bodyStr := string(bodyBytes)
	assert.Contains(t, bodyStr, "response_format")
	assert.Contains(t, bodyStr, "json_schema")
	assert.Contains(t, bodyStr, "test_schema")
}

func TestBuildRequest_InvalidJSONSchema(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()

	// Create a schema with a channel (which cannot be marshaled to JSON)
	invalidSchema := make(chan int)

	options := GenerateChatCompletionOptions{
		Model:      "openai/gpt-4o-mini",
		UserPrompt: "Test",
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchema{
				Name:   "test",
				Strict: true,
				Schema: invalidSchema,
			},
		},
	}

	req, err := service.buildRequest(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "failed to marshal json schema")
}

// Tests for parseResponse
func TestParseResponse_Success(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Test response"
			},
			"finish_reason": "stop"
		}]
	}`

	resp := newTestHTTPResponse(http.StatusOK, responseBody)

	result, err := service.parseResponse(resp)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "response-123", result.ID)
	assert.Equal(t, "openai/gpt-4o-mini", result.Model)
	assert.Len(t, result.Choices, 1)
	assert.Equal(t, "Test response", result.Choices[0].Message.Content)
}

func TestParseResponse_ReadBodyError(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	// Create a response with a body that errors on read
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&errorReader{}),
		Header:     make(http.Header),
	}

	result, err := service.parseResponse(resp)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to read response body")
}

func TestParseResponse_WithUsageInfo(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Test"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 100,
			"completion_tokens": 50,
			"total_tokens": 150
		}
	}`

	resp := newTestHTTPResponse(http.StatusOK, responseBody)

	result, err := service.parseResponse(resp)

	require.NoError(t, err)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 100, result.Usage.PromptTokens)
	assert.Equal(t, 50, result.Usage.CompletionTokens)
	assert.Equal(t, 150, result.Usage.TotalTokens)
}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// Tests for context cancellation
func TestGenerateChatCompletion_ContextCanceled(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	options := newTestChatCompletionOptions(
		"openai/gpt-4o-mini",
		"",
		"Hello",
	)

	// The request building should succeed, but the HTTP call should fail
	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).
		Return(nil, context.Canceled)

	result, err := service.GenerateChatCompletion(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send request")
}

// Tests for temperature and max_tokens parameters
func TestBuildRequest_WithTemperatureAndMaxTokens(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	service, _ := NewOpenRouterService(cfg, &http.Client{})

	ctx := context.Background()
	options := GenerateChatCompletionOptions{
		Model:       "openai/gpt-4o-mini",
		UserPrompt:  "Test",
		Temperature: 0.9,
		MaxTokens:   500,
	}

	req, err := service.buildRequest(ctx, options)

	require.NoError(t, err)
	assert.NotNil(t, req)

	// Verify request body contains temperature and max_tokens
	bodyBytes, _ := io.ReadAll(req.Body)
	bodyStr := string(bodyBytes)
	assert.Contains(t, bodyStr, "0.9")
	assert.Contains(t, bodyStr, "500")
}

// Tests for multiple choices in response
func TestParseResponse_MultipleChoices(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	responseBody := `{
		"id": "response-123",
		"model": "openai/gpt-4o-mini",
		"choices": [
			{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "First choice"
				},
				"finish_reason": "stop"
			},
			{
				"index": 1,
				"message": {
					"role": "assistant",
					"content": "Second choice"
				},
				"finish_reason": "stop"
			}
		]
	}`

	resp := newTestHTTPResponse(http.StatusOK, responseBody)

	result, err := service.parseResponse(resp)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Choices, 2)
	assert.Equal(t, "First choice", result.Choices[0].Message.Content)
	assert.Equal(t, "Second choice", result.Choices[1].Message.Content)
}

// Test request body structure
func TestBuildRequest_BodyStructure(t *testing.T) {
	cfg := newTestConfig("test-api-key")
	mockHTTP := new(MockHTTPClient)
	service, _ := NewOpenRouterService(cfg, mockHTTP)

	ctx := context.Background()
	options := GenerateChatCompletionOptions{
		Model:        "openai/gpt-4o-mini",
		SystemPrompt: "System",
		UserPrompt:   "User",
		Temperature:  0.5,
		MaxTokens:    1000,
	}

	req, err := service.buildRequest(ctx, options)

	require.NoError(t, err)

	// Read and parse the body
	bodyBytes, _ := io.ReadAll(req.Body)

	// Create a new reader for the body since we consumed it
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Verify the body can be parsed as chatCompletionRequest
	var reqBody chatCompletionRequest
	err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&reqBody)

	require.NoError(t, err)
	assert.Equal(t, "openai/gpt-4o-mini", reqBody.Model)
	assert.Equal(t, 0.5, reqBody.Temperature)
	assert.Equal(t, 1000, reqBody.MaxTokens)
	assert.Len(t, reqBody.Messages, 2)
	assert.Equal(t, "system", reqBody.Messages[0].Role)
	assert.Equal(t, "System", reqBody.Messages[0].Content)
	assert.Equal(t, "user", reqBody.Messages[1].Role)
	assert.Equal(t, "User", reqBody.Messages[1].Content)
}
