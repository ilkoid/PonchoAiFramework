package deepseek

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/core"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekModel_FrameworkIntegration(t *testing.T) {
	// Create framework with test configuration
	config := &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:    "deepseek",
				ModelName:   "deepseek-chat",
				APIKey:      "test-api-key",
				BaseURL:     "https://api.deepseek.com",
				MaxTokens:   2000,
				Temperature: 0.7,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
				},
				CustomParams: map[string]interface{}{
					"top_p":             0.9,
					"frequency_penalty": 0.0,
					"presence_penalty":  0.0,
				},
			},
		},
	}

	framework := core.NewPonchoFramework(config, nil)

	// Start framework
	err := framework.Start(context.Background())
	require.NoError(t, err)
	defer framework.Stop(context.Background())

	// Create and register DeepSeek model
	model := NewDeepSeekModel()
	err = framework.RegisterModel("deepseek-chat", model)
	require.NoError(t, err)

	// Verify model is registered
	modelRegistry := framework.GetModelRegistry()
	registeredModels := modelRegistry.List()
	assert.Contains(t, registeredModels, "deepseek-chat")

	// Test model retrieval
	retrievedModel, err := modelRegistry.Get("deepseek-chat")
	require.NoError(t, err)
	assert.Equal(t, "deepseek-chat", retrievedModel.Name())
	assert.Equal(t, "deepseek", retrievedModel.Provider())
	assert.True(t, retrievedModel.SupportsStreaming())
	assert.True(t, retrievedModel.SupportsTools())
	assert.False(t, retrievedModel.SupportsVision())
	assert.True(t, retrievedModel.SupportsSystemRole())
}

func TestDeepSeekModel_FrameworkGeneration(t *testing.T) {
	// Mock HTTP server for testing
	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)

		// Send mock response
		response := DeepSeekResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "deepseek-chat",
			Choices: []DeepSeekChoice{
				{
					Index: 0,
					Message: DeepSeekMessage{
						Role:    "assistant",
						Content: "Hello from DeepSeek!",
					},
					FinishReason: "stop",
				},
			},
			Usage: DeepSeekUsage{
				PromptTokens:     10,
				CompletionTokens: 15,
				TotalTokens:      25,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// Create framework with test configuration
	config := &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:    "deepseek",
				ModelName:   "deepseek-chat",
				APIKey:      "test-api-key",
				BaseURL:     server.URL,
				MaxTokens:   2000,
				Temperature: 0.7,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
				},
			},
		},
	}

	framework := core.NewPonchoFramework(config, nil)

	// Start framework
	err := framework.Start(context.Background())
	require.NoError(t, err)
	defer framework.Stop(context.Background())

	// Register DeepSeek model
	model := NewDeepSeekModel()
	err = framework.RegisterModel("deepseek-chat", model)
	require.NoError(t, err)

	// Test generation through framework
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello, framework!"},
				},
			},
		},
		Temperature: func() *float32 { t := float32(0.8); return &t }(),
		MaxTokens:   func() *int { i := 100; return &i }(),
	}

	resp, err := framework.Generate(context.Background(), req)
	require.NoError(t, err)

	// Verify response
	assert.NotNil(t, resp)
	assert.Equal(t, interfaces.PonchoRoleAssistant, resp.Message.Role)
	assert.Len(t, resp.Message.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeText, resp.Message.Content[0].Type)
	assert.Equal(t, "Hello from DeepSeek!", resp.Message.Content[0].Text)
	assert.Equal(t, interfaces.PonchoFinishReasonStop, resp.FinishReason)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 15, resp.Usage.CompletionTokens)
	assert.Equal(t, 25, resp.Usage.TotalTokens)
}

func TestDeepSeekModel_FrameworkStreaming(t *testing.T) {
	// Mock streaming server
	server := createMockStreamingServer(t)
	defer server.Close()

	// Create framework with test configuration
	config := &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:    "deepseek",
				ModelName:   "deepseek-chat",
				APIKey:      "test-api-key",
				BaseURL:     server.URL,
				MaxTokens:   2000,
				Temperature: 0.7,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
				},
			},
		},
	}

	framework := core.NewPonchoFramework(config, nil)

	// Start framework
	err := framework.Start(context.Background())
	require.NoError(t, err)
	defer framework.Stop(context.Background())

	// Register DeepSeek model
	model := NewDeepSeekModel()
	err = framework.RegisterModel("deepseek-chat", model)
	require.NoError(t, err)

	// Test streaming through framework
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello, streaming!"},
				},
			},
		},
		Stream: true,
	}

	var chunks []*interfaces.PonchoStreamChunk
	err = framework.GenerateStreaming(context.Background(), req, func(chunk *interfaces.PonchoStreamChunk) error {
		chunks = append(chunks, chunk)
		return nil
	})

	require.NoError(t, err)
	assert.Len(t, chunks, 3)

	// Verify final chunk
	finalChunk := chunks[2]
	assert.True(t, finalChunk.Done)
	assert.Equal(t, interfaces.PonchoFinishReasonStop, finalChunk.FinishReason)
	assert.Len(t, finalChunk.Delta.Content, 1)
	assert.Equal(t, "!", finalChunk.Delta.Content[0].Text)
}

func TestDeepSeekModel_FrameworkTools(t *testing.T) {
	// Mock HTTP server for tool testing
	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request contains tools
		var req DeepSeekRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Len(t, req.Tools, 1)
		assert.Equal(t, "test_tool", req.Tools[0].Function.Name)

		// Send response with tool call
		response := DeepSeekResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "deepseek-chat",
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
									Arguments: `{"result": "success"}`,
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

	// Create framework with test configuration
	config := &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:    "deepseek",
				ModelName:   "deepseek-chat",
				APIKey:      "test-api-key",
				BaseURL:     server.URL,
				MaxTokens:   2000,
				Temperature: 0.7,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
				},
			},
		},
	}

	framework := core.NewPonchoFramework(config, nil)

	// Start framework
	err := framework.Start(context.Background())
	require.NoError(t, err)
	defer framework.Stop(context.Background())

	// Register DeepSeek model
	model := NewDeepSeekModel()
	err = framework.RegisterModel("deepseek-chat", model)
	require.NoError(t, err)

	// Test generation with tools through framework
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
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
				Description: "A test tool",
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

	resp, err := framework.Generate(context.Background(), req)
	require.NoError(t, err)

	// Verify response contains tool call
	assert.NotNil(t, resp)
	assert.Equal(t, interfaces.PonchoFinishReasonTool, resp.FinishReason)
	assert.Len(t, resp.Message.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeTool, resp.Message.Content[0].Type)
	assert.Equal(t, "call_123", resp.Message.Content[0].Tool.ID)
	assert.Equal(t, "test_tool", resp.Message.Content[0].Tool.Name)
	assert.Equal(t, "success", resp.Message.Content[0].Tool.Args["result"])
}
