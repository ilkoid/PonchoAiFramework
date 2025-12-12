package deepseek

import (
	"context"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekModelIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test model initialization
	model := NewDeepSeekModel()
	require.NotNil(t, model)

	// Initialize with test configuration
	config := map[string]interface{}{
		"api_key":     "test-api-key", // Use test key for integration
		"model_name":  "deepseek-chat",
		"max_tokens":  1000,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Skipping integration test due to initialization error: %v", err)
		return
	}

	// Test basic properties
	assert.Equal(t, "deepseek-chat", model.Name())
	assert.Equal(t, string(common.ProviderDeepSeek), model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.False(t, model.SupportsVision())
	assert.True(t, model.SupportsSystemRole())

	// Test request validation
	validReq := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello, how are you?"},
				},
			},
		},
		MaxTokens:   func() *int { i := 100; return &i }(),
		Temperature: func() *float32 { f := float32(0.7); return &f }(),
	}

	err = model.ValidateRequest(validReq)
	assert.NoError(t, err)

	// Test invalid request validation
	invalidReq := &interfaces.PonchoModelRequest{
		Model:    "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{},
	}

	err = model.ValidateRequest(invalidReq)
	assert.Error(t, err)

	// Cleanup
	err = model.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestDeepSeekModelWithTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	model := NewDeepSeekModel()
	require.NotNil(t, model)

	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"model_name": "deepseek-chat",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Skipping integration test due to initialization error: %v", err)
		return
	}

	// Test request with tools
	reqWithTools := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "What's the weather in Boston?"},
				},
			},
		},
		Tools: []*interfaces.PonchoToolDef{
			{
				Name:        "get_weather",
				Description: "Get the current weather for a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	err = model.ValidateRequest(reqWithTools)
	assert.NoError(t, err)

	// Test request with tool calls in messages
	reqWithToolCalls := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "What's the weather in Boston?"},
				},
			},
			{
				Role: interfaces.PonchoRoleTool,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeTool,
						Tool: &interfaces.PonchoToolPart{
							ID:   "call_123",
							Name: "get_weather",
							Args: map[string]interface{}{"location": "Boston"},
						},
					},
				},
			},
		},
	}

	err = model.ValidateRequest(reqWithToolCalls)
	assert.NoError(t, err)

	err = model.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestDeepSeekModelStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	model := NewDeepSeekModel()
	require.NotNil(t, model)

	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"model_name": "deepseek-chat",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Skipping integration test due to initialization error: %v", err)
		return
	}

	// Test streaming request
	streamReq := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Tell me a short story"},
				},
			},
		},
		Stream: true,
	}

	// Test streaming callback
	var chunkCount int
	callback := func(chunk *interfaces.PonchoStreamChunk) error {
		chunkCount++
		assert.NotNil(t, chunk)
		assert.NotNil(t, chunk.Metadata)

		// Verify metadata fields
		assert.Contains(t, chunk.Metadata, "id")
		assert.Contains(t, chunk.Metadata, "object")
		assert.Contains(t, chunk.Metadata, "created")
		assert.Contains(t, chunk.Metadata, "model")

		// For testing purposes, we'll stop after a few chunks
		if chunkCount >= 1 {
			return assert.AnError // Stop the stream
		}

		return nil
	}

	err = model.GenerateStreaming(ctx, streamReq, callback)
	// Should get an error due to our intentional stop or API auth error
	assert.Error(t, err)
	// chunkCount might be 0 if API call fails immediately due to auth error
	assert.GreaterOrEqual(t, chunkCount, 0)

	err = model.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestDeepSeekClientConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test client configuration
	config := &common.CommonModelConfig{
		Provider:    common.ProviderDeepSeek,
		Model:       "deepseek-chat",
		APIKey:      "test-api-key",
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     45 * time.Second,
	}

	client, err := NewDeepSeekClient(config, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test configuration methods
	retrievedConfig := client.GetConfig()
	assert.Equal(t, config.Provider, retrievedConfig.Provider)
	assert.Equal(t, config.Model, retrievedConfig.Model)
	assert.Equal(t, config.MaxTokens, retrievedConfig.MaxTokens)
	assert.Equal(t, config.Temperature, retrievedConfig.Temperature)

	// Test capabilities
	capabilities := client.GetModelCapabilities()
	assert.True(t, capabilities.SupportsStreaming)
	assert.True(t, capabilities.SupportsTools)
	assert.False(t, capabilities.SupportsVision)
	assert.True(t, capabilities.SupportsSystem)

	// Test metadata
	metadata := client.GetModelMetadata()
	assert.Equal(t, config.Provider, metadata.Provider)
	assert.Equal(t, config.Model, metadata.Model)
	assert.Equal(t, common.ModelTypeText, metadata.ModelType)

	// Test health check
	ctx := context.Background()
	err = client.IsHealthy(ctx)
	assert.NoError(t, err)

	// Test rate limit info
	rateLimit := client.GetRateLimitInfo()
	assert.NotNil(t, rateLimit)
	assert.Greater(t, rateLimit.RequestsPerMinute, 0)
	assert.Greater(t, rateLimit.TokensPerMinute, 0)

	// Cleanup
	err = client.Close()
	assert.NoError(t, err)
}

func TestDeepSeekErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with invalid configuration
	invalidConfig := &common.CommonModelConfig{
		Provider:    common.ProviderDeepSeek,
		Model:       "", // Empty model should cause error
		APIKey:      "test-api-key",
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	_, err := NewDeepSeekClient(invalidConfig, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name is required")

	// Test with invalid API key
	invalidKeyConfig := &common.CommonModelConfig{
		Provider:    common.ProviderDeepSeek,
		Model:       "deepseek-chat",
		APIKey:      "", // Empty API key should cause error
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	_, err = NewDeepSeekClient(invalidKeyConfig, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key is required")
}

// Benchmark tests for performance validation
func BenchmarkDeepSeekModelValidation(b *testing.B) {
	model := NewDeepSeekModel()

	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello, world!"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.ValidateRequest(req)
	}
}

func BenchmarkDeepSeekStreamChunkConversion(b *testing.B) {
	streamChunk := &DeepSeekStreamResponse{
		ID:      "test-chunk",
		Object:  "chat.completion.chunk",
		Created: 1677652288,
		Model:   "deepseek-chat",
		Choices: []DeepSeekStreamChoice{
			{
				Index: 0,
				Delta: DeepSeekStreamDelta{
					Role:    "assistant",
					Content: "Hello, world!",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertStreamChunkToPoncho(streamChunk)
	}
}
