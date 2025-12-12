package deepseek

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeepSeekModel(t *testing.T) {
	model := NewDeepSeekModel()

	assert.NotNil(t, model)
	assert.Equal(t, "deepseek-chat", model.Name())
	assert.Equal(t, "deepseek", model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.False(t, model.SupportsVision())
	assert.True(t, model.SupportsSystemRole())
	assert.Equal(t, 4000, model.MaxTokens())
	assert.Equal(t, float32(0.7), model.DefaultTemperature())
}

func TestDeepSeekModel_Initialize(t *testing.T) {
	model := NewDeepSeekModel()

	// Test initialization with valid config
	config := map[string]interface{}{
		"api_key":     "test-api-key",
		"model_name":  "test-model",
		"max_tokens":  2000,
		"temperature": float32(0.5),
		"timeout":     "60s",
		"custom_params": map[string]interface{}{
			"top_p":             float32(0.9),
			"frequency_penalty": float32(0.1),
			"presence_penalty":  float32(0.2),
			"response_format": map[string]interface{}{
				"type": "json_object",
			},
			"thinking": map[string]interface{}{
				"type": "disabled",
			},
			"logprobs":     true,
			"top_logprobs": 5,
		},
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)
	assert.NotNil(t, model.client)
	assert.Equal(t, "test-api-key", model.client.GetConfig().APIKey)
	assert.Equal(t, "test-model", model.client.GetConfig().Model)
	assert.Equal(t, 2000, model.client.GetConfig().MaxTokens)
	assert.Equal(t, float32(0.5), model.client.GetConfig().Temperature)
	assert.Equal(t, 60*time.Second, model.client.GetConfig().Timeout)
}

func TestDeepSeekModel_Initialize_Error(t *testing.T) {
	model := NewDeepSeekModel()

	// Test initialization without API key
	config := map[string]interface{}{
		"model_name": "test-model",
	}

	err := model.Initialize(context.Background(), config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")
}

func TestDeepSeekModel_Generate(t *testing.T) {
	model := NewDeepSeekModel()

	// Mock HTTP server
	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		var req DeepSeekRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-model", req.Model)
		assert.Len(t, req.Messages, 2)
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "You are helpful assistant", req.Messages[0].Content)
		assert.Equal(t, "user", req.Messages[1].Role)
		assert.Equal(t, "Hello", req.Messages[1].Content)
		assert.Equal(t, float32(0.8), *req.Temperature)
		assert.Equal(t, 100, *req.MaxTokens)

		// Send response
		response := DeepSeekResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []DeepSeekChoice{
				{
					Index: 0,
					Message: DeepSeekMessage{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
			Usage: DeepSeekUsage{
				PromptTokens:     15,
				CompletionTokens: 20,
				TotalTokens:      35,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// Initialize model
	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"base_url":   server.URL,
		"model_name": "test-model",
	}
	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	// Test generation
	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleSystem,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "You are helpful assistant"},
				},
			},
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
				},
			},
		},
		Temperature: func() *float32 { t := float32(0.8); return &t }(),
		MaxTokens:   func() *int { i := 100; return &i }(),
	}

	resp, err := model.Generate(context.Background(), req)
	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.Equal(t, interfaces.PonchoRoleAssistant, resp.Message.Role)
	assert.Len(t, resp.Message.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeText, resp.Message.Content[0].Type)
	assert.Equal(t, "Hello! How can I help you today?", resp.Message.Content[0].Text)
	assert.Equal(t, interfaces.PonchoFinishReasonStop, resp.FinishReason)
	assert.Equal(t, 15, resp.Usage.PromptTokens)
	assert.Equal(t, 20, resp.Usage.CompletionTokens)
	assert.Equal(t, 35, resp.Usage.TotalTokens)
	assert.Equal(t, "test-id", resp.Metadata["id"])
	assert.Equal(t, "chat.completion", resp.Metadata["object"])
	assert.Equal(t, "test-model", resp.Metadata["model"])
}

func TestDeepSeekModel_Generate_WithTools(t *testing.T) {
	model := NewDeepSeekModel()

	// Mock HTTP server
	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req DeepSeekRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify tools
		assert.Len(t, req.Tools, 1)
		assert.Equal(t, "function", req.Tools[0].Type)
		assert.Equal(t, "test_tool", req.Tools[0].Function.Name)
		assert.Equal(t, "Test tool", req.Tools[0].Function.Description)
		assert.Equal(t, DeepSeekToolChoiceAuto, req.ToolChoice)

		// Send response with tool call
		response := DeepSeekResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []DeepSeekChoice{
				{
					Index: 0,
					Message: DeepSeekMessage{
						Role: "assistant",
						ToolCalls: []DeepSeekToolCall{
							{
								ID:   "call_123",
								Type: "function",
								Function: DeepSeekFunctionCall{
									Name:      "test_tool",
									Arguments: `{"param1": "value1"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: DeepSeekUsage{
				PromptTokens:     20,
				CompletionTokens: 10,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// Initialize model
	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"base_url":   server.URL,
		"model_name": "test-model",
	}
	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	// Test generation with tools
	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Use the test tool"},
				},
			},
		},
		Tools: []*interfaces.PonchoToolDef{
			{
				Name:        "test_tool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param1": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	}

	resp, err := model.Generate(context.Background(), req)
	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.Equal(t, interfaces.PonchoFinishReasonTool, resp.FinishReason)
	assert.Len(t, resp.Message.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeTool, resp.Message.Content[0].Type)
	assert.Equal(t, "call_123", resp.Message.Content[0].Tool.ID)
	assert.Equal(t, "test_tool", resp.Message.Content[0].Tool.Name)
	assert.Equal(t, "value1", resp.Message.Content[0].Tool.Args["param1"])
}

func TestDeepSeekModel_GenerateStreaming(t *testing.T) {
	model := NewDeepSeekModel()

	// Mock streaming server
	server := createMockStreamingServer(t)
	defer server.Close()

	// Initialize model
	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"base_url":   server.URL,
		"model_name": "test-model",
	}
	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	// Test streaming generation
	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
				},
			},
		},
		Stream: true,
	}

	var chunks []*interfaces.PonchoStreamChunk
	err = model.GenerateStreaming(context.Background(), req, func(chunk *interfaces.PonchoStreamChunk) error {
		chunks = append(chunks, chunk)
		return nil
	})

	require.NoError(t, err)
	assert.Len(t, chunks, 3)

	// Verify first chunk (role)
	assert.Equal(t, interfaces.PonchoRoleAssistant, chunks[0].Delta.Role)
	assert.Len(t, chunks[0].Delta.Content, 0) // First chunk only has role, no content
	assert.False(t, chunks[0].Done)

	// Verify second chunk (content)
	assert.Equal(t, "Hello", chunks[1].Delta.Content[0].Text)
	assert.False(t, chunks[1].Done)

	// Verify final chunk
	assert.Equal(t, "!", chunks[2].Delta.Content[0].Text)
	assert.True(t, chunks[2].Done)
	assert.Equal(t, interfaces.PonchoFinishReasonStop, chunks[2].FinishReason)
}

func TestDeepSeekModel_Shutdown(t *testing.T) {
	model := NewDeepSeekModel()

	// Initialize model first
	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"model_name": "test-model",
	}
	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	// Test shutdown
	err = model.Shutdown(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, model.client)
}

// Helper functions

func createMockServer(t *testing.T, handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func createMockStreamingServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		chunks := []string{
			`data: {"id":"chunk1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"role":"assistant"}}]}`,
			`data: {"id":"chunk2","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"content":"Hello"}}]}`,
			`data: {"id":"chunk3","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			w.(http.Flusher).Flush()
		}
	}))
}
