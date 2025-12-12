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

func TestNewZAIClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *common.CommonModelConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &common.CommonModelConfig{
				Provider:    common.ProviderZAI,
				Model:       "glm-4.6",
				APIKey:      "test-api-key",
				MaxTokens:   1000,
				Temperature: 0.5,
				Timeout:     30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "missing api key",
			config: &common.CommonModelConfig{
				Provider:    common.ProviderZAI,
				Model:       "glm-4.6",
				MaxTokens:   1000,
				Temperature: 0.5,
				Timeout:     30 * time.Second,
			},
			expectError: true,
			errorMsg:    "ZAI_API_KEY environment variable is required",
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "invalid temperature",
			config: &common.CommonModelConfig{
				Provider:    common.ProviderZAI,
				Model:       "glm-4.6",
				APIKey:      "test-api-key",
				MaxTokens:   1000,
				Temperature: 3.0, // Invalid: > 2.0
				Timeout:     30 * time.Second,
			},
			expectError: true,
			errorMsg:    "temperature must be between 0 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable for missing api key test
			if tt.name == "missing api key" {
				os.Unsetenv("ZAI_API_KEY")
			}

			client, err := NewZAIClient(tt.config, nil)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestNewZAIClient_WithEnvVar(t *testing.T) {
	// Set environment variable
	os.Setenv("ZAI_API_KEY", "env-api-key")
	defer os.Unsetenv("ZAI_API_KEY")

	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "env-api-key", client.apiKey)
}

func TestZAIClient_PrepareHeaders(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	headers := client.PrepareHeaders()

	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "application/json", headers["Accept"])
	assert.Equal(t, "Bearer test-api-key", headers["Authorization"])
	assert.Equal(t, "PonchoFramework-ZAI/1.0", headers["User-Agent"])
}

func TestZAIClient_BuildURL(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		BaseURL:     "https://api.example.com",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	url := client.BuildURL("/chat/completions")
	assert.Equal(t, "https://api.example.com/chat/completions", url)
}

func TestZAIClient_BuildURL_DefaultBaseURL(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
		// BaseURL is empty - should use default
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	url := client.BuildURL("/chat/completions")
	assert.Equal(t, "https://api.z.ai/api/paas/v4/chat/completions", url)
}

func TestZAIClient_ValidateRequest(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6v", // Vision model
		ModelType:   common.ModelTypeVision,
		APIKey:      "test-api-key",
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

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
			name: "valid text request",
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
		{
			name: "valid vision request",
			request: &interfaces.PonchoModelRequest{
				Messages: []*interfaces.PonchoMessage{
					{
						Role: interfaces.PonchoRoleUser,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeText,
								Text: "What do you see?",
							},
							{
								Type: interfaces.PonchoContentTypeMedia,
								Media: &interfaces.PonchoMediaPart{
									URL:      "https://example.com/image.jpg",
									MimeType: "image/jpeg",
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "media content with non-vision model",
			request: &interfaces.PonchoModelRequest{
				Messages: []*interfaces.PonchoMessage{
					{
						Role: interfaces.PonchoRoleUser,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeMedia,
								Media: &interfaces.PonchoMediaPart{
									URL:      "https://example.com/image.jpg",
									MimeType: "image/jpeg",
								},
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "media content not supported for non-vision model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			// For non-vision model test, create a text model client
			if tt.name == "media content with non-vision model" {
				textConfig := &common.CommonModelConfig{
					Provider:    common.ProviderZAI,
					Model:       "glm-4.6",
					ModelType:   common.ModelTypeText,
					APIKey:      "test-api-key",
					MaxTokens:   1000,
					Temperature: 0.5,
					Timeout:     30 * time.Second,
				}
				textClient, createErr := NewZAIClient(textConfig, nil)
				require.NoError(t, createErr)
				err = textClient.ValidateRequest(tt.request)
			} else {
				err = client.ValidateRequest(tt.request)
			}

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

func TestZAIClient_GetModelCapabilities(t *testing.T) {
	// Test text model
	textConfig := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		ModelType:   common.ModelTypeText,
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	textClient, err := NewZAIClient(textConfig, nil)
	require.NoError(t, err)

	textCaps := textClient.GetModelCapabilities()
	assert.True(t, textCaps.SupportsStreaming)
	assert.True(t, textCaps.SupportsTools)
	assert.False(t, textCaps.SupportsVision)
	assert.True(t, textCaps.SupportsSystem)
	assert.Contains(t, textCaps.SupportedTypes, common.ContentTypeText)
	assert.Equal(t, 1000, textCaps.MaxInputTokens)
	assert.Equal(t, 1000, textCaps.MaxOutputTokens)

	// Test vision model
	visionConfig := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6v",
		ModelType:   common.ModelTypeVision,
		APIKey:      "test-api-key",
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	visionClient, err := NewZAIClient(visionConfig, nil)
	require.NoError(t, err)

	visionCaps := visionClient.GetModelCapabilities()
	assert.True(t, visionCaps.SupportsStreaming)
	assert.True(t, visionCaps.SupportsTools)
	assert.True(t, visionCaps.SupportsVision)
	assert.True(t, visionCaps.SupportsSystem)
	assert.Contains(t, visionCaps.SupportedTypes, common.ContentTypeText)
	assert.Contains(t, visionCaps.SupportedTypes, common.ContentTypeImageURL)
	assert.Contains(t, visionCaps.SupportedTypes, common.ContentTypeMedia)
	assert.Equal(t, 2000, visionCaps.MaxInputTokens)
	assert.Equal(t, 2000, visionCaps.MaxOutputTokens)
}

func TestZAIClient_GetModelMetadata(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		ModelType:   common.ModelTypeText,
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	metadata := client.GetModelMetadata()
	assert.Equal(t, common.ProviderZAI, metadata.Provider)
	assert.Equal(t, "glm-4.6", metadata.Model)
	assert.Equal(t, common.ModelTypeText, metadata.ModelType)
	assert.True(t, metadata.Capabilities.SupportsStreaming)
	assert.True(t, metadata.Capabilities.SupportsTools)
	assert.False(t, metadata.Capabilities.SupportsVision)
	assert.True(t, metadata.Capabilities.SupportsSystem)
	assert.Equal(t, "1.0", metadata.Version)
	assert.Contains(t, metadata.Description, "GLM")
	assert.Equal(t, 0.002, metadata.CostPer1KTokens)
}

func TestZAIClient_IsHealthy(t *testing.T) {
	// Test healthy client
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		BaseURL:     "https://api.example.com",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	err = client.IsHealthy(context.Background())
	assert.NoError(t, err)

	// Test unhealthy client - no API key
	os.Unsetenv("ZAI_API_KEY") // Ensure env var is not set
	unhealthyConfig := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	unhealthyClient, err := NewZAIClient(unhealthyConfig, nil)
	// This should fail because no API key is provided
	if err == nil {
		// If client creation succeeds (unlikely), test health check
		err = unhealthyClient.IsHealthy(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key is not configured")
	} else {
		// Client creation should fail, which is also correct
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ZAI_API_KEY environment variable is required")
	}
}

func TestZAIClient_GetRateLimitInfo(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	rateLimit := client.GetRateLimitInfo()
	assert.Equal(t, 120, rateLimit.RequestsPerMinute)
	assert.Equal(t, 200000, rateLimit.TokensPerMinute)
}

func TestZAIClient_UpdateConfig(t *testing.T) {
	originalConfig := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(originalConfig, nil)
	require.NoError(t, err)

	// Update configuration
	newConfig := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6v",
		APIKey:      "new-api-key",
		MaxTokens:   2000,
		Temperature: 0.3,
		Timeout:     60 * time.Second,
		BaseURL:     "https://new-api.example.com",
	}

	err = client.UpdateConfig(newConfig)
	assert.NoError(t, err)

	// Verify updates
	assert.Equal(t, "glm-4.6v", client.GetConfig().Model)
	assert.Equal(t, "new-api-key", client.apiKey)
	assert.Equal(t, 2000, client.GetConfig().MaxTokens)
	assert.Equal(t, float32(0.3), client.GetConfig().Temperature)
}

func TestZAIClient_UpdateVisionConfig(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6v",
		ModelType:   common.ModelTypeVision,
		APIKey:      "test-api-key",
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	// Update vision config
	newVisionConfig := &ZAIVisionConfig{
		MaxImageSize:   5 * 1024 * 1024, // 5MB
		SupportedTypes: []string{"image/png", "image/jpeg"},
		Quality:        ZAIVisionQualityHigh,
		Detail:         ZAIVisionDetailHigh,
	}

	client.UpdateVisionConfig(newVisionConfig)

	// Verify updates
	updatedConfig := client.GetVisionConfig()
	assert.Equal(t, 5*1024*1024, updatedConfig.MaxImageSize)
	assert.Equal(t, []string{"image/png", "image/jpeg"}, updatedConfig.SupportedTypes)
	assert.Equal(t, ZAIVisionQualityHigh, updatedConfig.Quality)
	assert.Equal(t, ZAIVisionDetailHigh, updatedConfig.Detail)
}

func TestZAIClient_Close(t *testing.T) {
	config := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       "glm-4.6",
		APIKey:      "test-api-key",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     30 * time.Second,
	}

	client, err := NewZAIClient(config, nil)
	require.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}
