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

func TestNewDeepSeekClient(t *testing.T) {
	logger := interfaces.NewDefaultLogger()

	// Test with nil config
	client := NewDeepSeekClient(nil, logger)
	assert.NotNil(t, client)
	assert.Equal(t, DeepSeekDefaultBaseURL, client.config.BaseURL)
	assert.Equal(t, DeepSeekDefaultModel, client.config.Model)
	assert.Equal(t, 4000, client.config.MaxTokens)
	assert.Equal(t, float32(0.7), client.config.Temperature)
	assert.Equal(t, 30*time.Second, client.config.Timeout)

	// Test with custom config
	config := &DeepSeekConfig{
		BaseURL:     "https://custom.deepseek.com",
		Model:       "custom-model",
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     60 * time.Second,
		APIKey:      "test-key",
	}
	client = NewDeepSeekClient(config, logger)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.config)
}

func TestDeepSeekClient_ValidateRequest(t *testing.T) {
	client := NewDeepSeekClient(nil, interfaces.NewDefaultLogger())

	// Test nil request
	err := client.ValidateRequest(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")

	// Test empty messages
	req := &DeepSeekRequest{}
	err = client.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must contain at least one message")

	// Test valid request
	req.Messages = []DeepSeekMessage{
		{Role: "user", Content: "Hello"},
	}
	err = client.ValidateRequest(req)
	assert.NoError(t, err)

	// Test invalid temperature
	req.Temperature = func() *float32 { t := float32(-1); return &t }()
	err = client.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temperature must be between 0 and 2")

	// Test invalid max_tokens
	req.Temperature = nil
	req.MaxTokens = func() *int { i := -1; return &i }()
	err = client.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_tokens must be positive")
}

func TestDeepSeekClient_PrepareRequest(t *testing.T) {
	config := &DeepSeekConfig{
		Model:            "test-model",
		MaxTokens:        1000,
		Temperature:      0.8,
		TopP:             func() *float32 { p := float32(0.9); return &p }(),
		FrequencyPenalty: func() *float32 { f := float32(0.1); return &f }(),
		PresencePenalty:  func() *float32 { p := float32(0.2); return &p }(),
		Stop:             []string{"stop1", "stop2"},
		ResponseFormat:   &DeepSeekResponseFormat{Type: "json_object"},
		Thinking:         &DeepSeekThinking{Type: "disabled"},
		LogProbs:         true,
		TopLogProbs:      func() *int { i := 5; return &i }(),
	}
	client := NewDeepSeekClient(config, interfaces.NewDefaultLogger())

	req := client.PrepareRequest()
	assert.Equal(t, config.Model, req.Model)
	assert.Equal(t, &config.MaxTokens, req.MaxTokens)
	assert.Equal(t, &config.Temperature, req.Temperature)
	assert.Equal(t, config.TopP, req.TopP)
	assert.Equal(t, config.FrequencyPenalty, req.FrequencyPenalty)
	assert.Equal(t, config.PresencePenalty, req.PresencePenalty)
	assert.Equal(t, config.Stop, req.Stop)
	assert.Equal(t, config.ResponseFormat, req.ResponseFormat)
	assert.Equal(t, config.Thinking, req.Thinking)
	assert.Equal(t, config.LogProbs, req.LogProbs)
	assert.Equal(t, config.TopLogProbs, req.TopLogProbs)
}

func TestDeepSeekClient_CreateChatCompletion(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Verify request body
		var req DeepSeekRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "deepseek-chat", req.Model)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)
		assert.Equal(t, "Hello", req.Messages[0].Content)

		// Send mock response
		response := DeepSeekResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "deepseek-chat",
			Choices: []DeepSeekChoice{
				{
					Index: 0,
					Message: DeepSeekMessage{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
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
	}))
	defer server.Close()

	// Create client with mock server URL
	config := &DeepSeekConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Model:   "deepseek-chat",
	}
	client := NewDeepSeekClient(config, interfaces.NewDefaultLogger())

	// Test successful request
	req := &DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Equal(t, "deepseek-chat", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.Equal(t, "Hello! How can I help you?", resp.Choices[0].Message.Content)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 15, resp.Usage.CompletionTokens)
	assert.Equal(t, 25, resp.Usage.TotalTokens)
}

func TestDeepSeekClient_CreateChatCompletion_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := DeepSeekError{
			Error: DeepSeekErrorDetail{
				Message: "Invalid request",
				Type:    "invalid_request_error",
				Code:    "invalid_request",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &DeepSeekConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	}
	client := NewDeepSeekClient(config, interfaces.NewDefaultLogger())

	req := &DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid request")
	assert.Contains(t, err.Error(), "invalid_request_error")
}

func TestDeepSeekClient_CreateChatCompletionStream(t *testing.T) {
	// Create mock server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))

		// Send streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		// Send chunks
		chunks := []string{
			`data: {"id":"chunk1","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{"role":"assistant"}}]}`,
			`data: {"id":"chunk2","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"Hello"}}]}`,
			`data: {"id":"chunk3","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()

	config := &DeepSeekConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	}
	client := NewDeepSeekClient(config, interfaces.NewDefaultLogger())

	req := &DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	var chunks []*DeepSeekStreamResponse
	err := client.CreateChatCompletionStream(context.Background(), req, func(chunk *DeepSeekStreamResponse) error {
		chunks = append(chunks, chunk)
		return nil
	})

	require.NoError(t, err)
	assert.Len(t, chunks, 3) // Should skip [DONE] chunk

	// Verify first chunk
	assert.Equal(t, "chunk1", chunks[0].ID)
	assert.Equal(t, "chat.completion.chunk", chunks[0].Object)
	assert.Len(t, chunks[0].Choices, 1)
	assert.Equal(t, "assistant", chunks[0].Choices[0].Delta.Role)

	// Verify second chunk
	assert.Equal(t, "chunk2", chunks[1].ID)
	assert.Equal(t, "Hello", chunks[1].Choices[0].Delta.Content)

	// Verify third chunk
	assert.Equal(t, "chunk3", chunks[2].ID)
	assert.Equal(t, "!", chunks[2].Choices[0].Delta.Content)
	assert.NotNil(t, chunks[2].Choices[0].FinishReason)
	assert.Equal(t, "stop", *chunks[2].Choices[0].FinishReason)
}

func TestDeepSeekClient_SetConfig(t *testing.T) {
	client := NewDeepSeekClient(nil, interfaces.NewDefaultLogger())

	newConfig := &DeepSeekConfig{
		BaseURL: "https://new.deepseek.com",
		Model:   "new-model",
		Timeout: 45 * time.Second,
	}

	client.SetConfig(newConfig)
	assert.Equal(t, newConfig, client.GetConfig())
	assert.Equal(t, 45*time.Second, client.httpClient.Timeout)
}

func TestDeepSeekClient_Close(t *testing.T) {
	client := NewDeepSeekClient(nil, interfaces.NewDefaultLogger())

	// Should not panic
	err := client.Close()
	assert.NoError(t, err)
}
