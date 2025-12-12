package zai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZAIModel(t *testing.T) {
	model := NewZAIModel()

	assert.NotNil(t, model)
	assert.Equal(t, "glm-4.6", model.Name())
	assert.Equal(t, string(common.ProviderZAI), model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.True(t, model.SupportsVision())
	assert.True(t, model.SupportsSystemRole())
}

func TestNewZAIVisionModel(t *testing.T) {
	model := NewZAIVisionModel()

	assert.NotNil(t, model)
	assert.Equal(t, "glm-4.6v", model.Name())
	assert.Equal(t, string(common.ProviderZAI), model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.True(t, model.SupportsVision())
	assert.True(t, model.SupportsSystemRole())
}

func TestZAIModel_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"api_key":     "test-api-key",
				"model_name":  "glm-4.6",
				"max_tokens":  1000,
				"temperature": 0.5,
			},
			expectError: false,
		},
		{
			name: "missing api key",
			config: map[string]interface{}{
				"model_name": "glm-4.6",
				"max_tokens": 1000,
			},
			expectError: true,
			errorMsg:    "api_key is required",
		},
		{
			name: "invalid temperature",
			config: map[string]interface{}{
				"api_key":     "test-api-key",
				"model_name":  "glm-4.6",
				"temperature": 3.0, // Invalid: > 2.0
			},
			expectError: true,
		},
		{
			name: "vision model config",
			config: map[string]interface{}{
				"api_key":    "test-api-key",
				"model_name": "glm-4.6v",
				"max_tokens": 2000,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewZAIModel()

			err := model.Initialize(context.Background(), tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, model.isInitialized())
			}
		})
	}
}

func TestZAIModel_InitializeWithEnvVar(t *testing.T) {
	// Set environment variable
	os.Setenv("ZAI_API_KEY", "env-api-key")
	defer os.Unsetenv("ZAI_API_KEY")

	model := NewZAIModel()
	config := map[string]interface{}{
		"model_name": "glm-4.6",
		"max_tokens": 1000,
	}

	err := model.Initialize(context.Background(), config)
	assert.NoError(t, err)
	assert.True(t, model.isInitialized())
}

func TestZAIModel_Generate(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		request     *interfaces.PonchoModelRequest
		expectError bool
	}{
		{
			name: "simple text generation",
			request: &interfaces.PonchoModelRequest{
				Model: "glm-4.6",
				Messages: []*interfaces.PonchoMessage{
					{
						Role: interfaces.PonchoRoleUser,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeText,
								Text: "Hello, how are you?",
							},
						},
					},
				},
				MaxTokens:   intPtr(50),
				Temperature: float32Ptr(0.7),
			},
			expectError: false,
		},
		{
			name: "system message",
			request: &interfaces.PonchoModelRequest{
				Model: "glm-4.6",
				Messages: []*interfaces.PonchoMessage{
					{
						Role: interfaces.PonchoRoleSystem,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeText,
								Text: "You are a helpful assistant.",
							},
						},
					},
					{
						Role: interfaces.PonchoRoleUser,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeText,
								Text: "What can you do?",
							},
						},
					},
				},
				MaxTokens: intPtr(50),
			},
			expectError: false,
		},
		{
			name: "tool call",
			request: &interfaces.PonchoModelRequest{
				Model: "glm-4.6",
				Messages: []*interfaces.PonchoMessage{
					{
						Role: interfaces.PonchoRoleUser,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeText,
								Text: "What's the weather like?",
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
				MaxTokens: intPtr(50),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := model.Generate(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Message)
				assert.NotEmpty(t, resp.Message.Content)
				assert.NotNil(t, resp.Usage)
				assert.Greater(t, resp.Usage.TotalTokens, 0)
			}
		})
	}
}

func TestZAIModel_GenerateStreaming(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  100,
		"temperature": 0.7,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	request := &interfaces.PonchoModelRequest{
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Write a short story about a robot.",
					},
				},
			},
		},
		MaxTokens:   intPtr(100),
		Temperature: float32Ptr(0.7),
		Stream:      true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var chunks []*interfaces.PonchoStreamChunk
	err = model.GenerateStreaming(ctx, request, func(chunk *interfaces.PonchoStreamChunk) error {
		chunks = append(chunks, chunk)
		return nil
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, chunks)

	// Check that we got a completion
	lastChunk := chunks[len(chunks)-1]
	assert.True(t, lastChunk.Done)
	assert.NotEqual(t, interfaces.PonchoFinishReasonStop, lastChunk.FinishReason)
}

func TestZAIModel_ValidateRequest(t *testing.T) {
	model := NewZAIModel()

	tests := []struct {
		name        string
		request     *interfaces.PonchoModelRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil request",
			request:     nil,
			expectError: true,
			errorMsg:    "request cannot be nil",
		},
		{
			name: "empty messages",
			request: &interfaces.PonchoModelRequest{
				Messages: []*interfaces.PonchoMessage{},
			},
			expectError: true,
			errorMsg:    "request must contain at least one message",
		},
		{
			name: "valid request",
			request: &interfaces.PonchoModelRequest{
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
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.ValidateRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestZAIModel_Shutdown(t *testing.T) {
	model := NewZAIModel()

	// Test shutdown without initialization
	err := model.Shutdown(context.Background())
	assert.NoError(t, err)

	// Test shutdown after initialization
	config := map[string]interface{}{
		"api_key":    "test-key",
		"model_name": "glm-4.6",
	}
	err = model.Initialize(context.Background(), config)
	require.NoError(t, err)

	err = model.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestZAIModel_convertConfig(t *testing.T) {
	model := NewZAIModel()

	tests := []struct {
		name        string
		config      map[string]interface{}
		expected    *common.CommonModelConfig
		expectError bool
	}{
		{
			name: "minimal config",
			config: map[string]interface{}{
				"model_name": "glm-4.6",
			},
			expected: &common.CommonModelConfig{
				Provider:    common.ProviderZAI,
				Model:       "glm-4.6",
				ModelType:   common.ModelTypeMultimodal,
				MaxTokens:   2000,
				Temperature: 0.5,
				Timeout:     60 * time.Second,
			},
			expectError: true, // API key is required
		},
		{
			name: "vision model",
			config: map[string]interface{}{
				"model_name":  "glm-4.6v",
				"max_tokens":  1500,
				"temperature": 0.3,
			},
			expected: &common.CommonModelConfig{
				Provider:    common.ProviderZAI,
				Model:       "glm-4.6v",
				ModelType:   common.ModelTypeVision,
				MaxTokens:   1500,
				Temperature: 0.3,
				Timeout:     60 * time.Second,
			},
			expectError: true, // API key is required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.convertConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Provider, result.Provider)
				assert.Equal(t, tt.expected.Model, result.Model)
				assert.Equal(t, tt.expected.ModelType, result.ModelType)
				assert.Equal(t, tt.expected.MaxTokens, result.MaxTokens)
				assert.Equal(t, tt.expected.Temperature, result.Temperature)
				assert.Equal(t, tt.expected.Timeout, result.Timeout)
			}
		})
	}
}
