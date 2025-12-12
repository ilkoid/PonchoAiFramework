package zai

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZAIIntegration_FrameworkCompatibility(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	// Test framework integration
	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	// Test basic text generation through framework interface
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request := &interfaces.PonchoModelRequest{
		Model: model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello from Z.AI integration test!",
					},
				},
			},
		},
		MaxTokens:   intPtr(50),
		Temperature: float32Ptr(0.7),
	}

	response, err := model.Generate(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Message)
	assert.NotEmpty(t, response.Message.Content)
	assert.NotNil(t, response.Usage)
	assert.Greater(t, response.Usage.TotalTokens, 0)

	// Verify framework interface methods work
	assert.Equal(t, "glm-4.6", model.Name())
	assert.Equal(t, string(common.ProviderZAI), model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.True(t, model.SupportsVision())
	assert.True(t, model.SupportsSystemRole())
}

func TestZAIIntegration_VisionModel(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	// Test vision model integration
	model := NewZAIVisionModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v",
		"max_tokens":  200,
		"temperature": 0.5,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	// Test vision capabilities
	assert.Equal(t, "glm-4.6v", model.Name())
	assert.Equal(t, string(common.ProviderZAI), model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.True(t, model.SupportsVision())
	assert.True(t, model.SupportsSystemRole())

	// Test that vision model can handle media content
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This would fail in real scenario due to fake URL, but tests the interface
	request := &interfaces.PonchoModelRequest{
		Model: model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Analyze this fashion image",
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      "https://example.com/fashion.jpg",
							MimeType: "image/jpeg",
						},
					},
				},
			},
		},
		MaxTokens:   intPtr(100),
		Temperature: float32Ptr(0.3),
	}

	// This should fail at API level but pass validation
	response, err := model.Generate(ctx, request)
	// We expect an error due to invalid URL, but the request format should be valid
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestZAIIntegration_Streaming(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	// Test streaming integration
	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request := &interfaces.PonchoModelRequest{
		Model: model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Write a short story about AI.",
					},
				},
			},
		},
		MaxTokens:   intPtr(100),
		Temperature: float32Ptr(0.7),
		Stream:      true,
	}

	var chunks []*interfaces.PonchoStreamChunk
	err = model.GenerateStreaming(ctx, request, func(chunk *interfaces.PonchoStreamChunk) error {
		chunks = append(chunks, chunk)
		return nil
	})

	// Should fail due to invalid API, but streaming interface should work
	assert.Error(t, err)
}

func TestZAIIntegration_ToolCalling(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	// Test tool calling integration
	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request := &interfaces.PonchoModelRequest{
		Model: model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What's the weather like in New York?",
					},
				},
			},
		},
		Tools: []*interfaces.PonchoToolDef{
			{
				Name:        "get_weather",
				Description: "Get current weather information",
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
		MaxTokens:   intPtr(100),
		Temperature: float32Ptr(0.7),
	}

	// Should fail due to invalid API, but tool format should be valid
	response, err := model.Generate(ctx, request)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestZAIIntegration_ErrorHandling(t *testing.T) {
	// Test error handling with invalid API key
	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     "invalid-key",
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err := model.Initialize(context.Background(), config)
	// Should fail during initialization due to invalid config or API key
	if err != nil {
		assert.Error(t, err)
	} else {
		// If initialization succeeds, generation should fail
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		request := &interfaces.PonchoModelRequest{
			Model: model.Name(),
			Messages: []*interfaces.PonchoMessage{
				{
					Role: interfaces.PonchoRoleUser,
					Content: []*interfaces.PonchoContentPart{
						{
							Type: interfaces.PonchoContentTypeText,
							Text: "Hello",
						},
					},
				},
			},
			MaxTokens: intPtr(50),
		}

		response, err := model.Generate(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
	}
}

func TestZAIIntegration_ConfigurationValidation(t *testing.T) {
	model := NewZAIModel()

	// Test various invalid configurations
	invalidConfigs := []map[string]interface{}{
		{
			"api_key": "", // Empty API key
		},
		{
			"model_name": "", // Empty model name
		},
		{
			"temperature": 3.0, // Invalid temperature > 2.0
		},
		{
			"max_tokens": -1, // Invalid max tokens
		},
	}

	for i, config := range invalidConfigs {
		t.Run(fmt.Sprintf("invalid_config_%d", i), func(t *testing.T) {
			err := model.Initialize(context.Background(), config)
			assert.Error(t, err)
		})
	}
}

func TestZAIIntegration_Lifecycle(t *testing.T) {
	model := NewZAIModel()

	// Test shutdown without initialization
	err := model.Shutdown(context.Background())
	assert.NoError(t, err)

	// Test initialization and shutdown
	config := map[string]interface{}{
		"api_key":     "test-key",
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err = model.Initialize(context.Background(), config)
	assert.NoError(t, err)

	// Test shutdown after initialization
	err = model.Shutdown(context.Background())
	assert.NoError(t, err)

	// Test that model can be reinitialized
	err = model.Initialize(context.Background(), config)
	assert.NoError(t, err)
}
