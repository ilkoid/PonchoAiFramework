package zai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVisionProcessor(t *testing.T) {
	model := NewZAIModel()
	logger := interfaces.NewDefaultLogger()

	// Test with default config
	processor := NewVisionProcessor(model, logger, nil)
	assert.NotNil(t, processor)
	assert.Equal(t, ZAIVisionMaxImageSize, processor.GetConfig().MaxImageSize)
	assert.Equal(t, ZAIVisionQualityAuto, processor.GetConfig().DefaultQuality)
	assert.Equal(t, ZAIVisionDetailAuto, processor.GetConfig().DefaultDetail)

	// Test with custom config
	customConfig := &VisionConfig{
		MaxImageSize:     5 * 1024 * 1024, // 5MB
		SupportedFormats: []string{"image/png"},
		DefaultQuality:   ZAIVisionQualityHigh,
		DefaultDetail:    ZAIVisionDetailHigh,
		Timeout:          45 * time.Second,
		EnableCaching:    true,
		CacheTTL:         2 * time.Hour,
	}

	customProcessor := NewVisionProcessor(model, logger, customConfig)
	assert.NotNil(t, customProcessor)
	assert.Equal(t, customConfig, customProcessor.GetConfig())
}

func TestVisionProcessor_ValidateImageURL(t *testing.T) {
	model := NewZAIModel()
	logger := interfaces.NewDefaultLogger()
	processor := NewVisionProcessor(model, logger, nil)

	tests := []struct {
		name        string
		imageURL    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid image URL",
			imageURL:    "https://example.com/fashion-image.jpg",
			expectError: false,
		},
		{
			name:        "empty image URL",
			imageURL:    "",
			expectError: true,
			errorMsg:    "image URL cannot be empty",
		},
		{
			name:        "invalid URL scheme",
			imageURL:    "ftp://example.com/image.jpg",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.validateImageURL(tt.imageURL)

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

func TestVisionProcessor_ValidateBase64Image(t *testing.T) {
	model := NewZAIModel()
	logger := interfaces.NewDefaultLogger()
	processor := NewVisionProcessor(model, logger, nil)

	// Valid base64 PNG data (minimal 1x1 pixel PNG)
	validPNGBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

	tests := []struct {
		name        string
		base64Data  string
		mimeType    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid PNG base64",
			base64Data:  validPNGBase64,
			mimeType:    "image/png",
			expectError: false,
		},
		{
			name:        "empty base64 data",
			base64Data:  "",
			mimeType:    "image/jpeg",
			expectError: true,
			errorMsg:    "base64 data cannot be empty",
		},
		{
			name:        "unsupported mime type",
			base64Data:  validPNGBase64,
			mimeType:    "image/svg+xml",
			expectError: true,
			errorMsg:    "unsupported mime type",
		},
		{
			name:        "invalid base64",
			base64Data:  "invalid-base64-data",
			mimeType:    "image/jpeg",
			expectError: true,
			errorMsg:    "invalid base64 data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.validateBase64Image(tt.base64Data, tt.mimeType)

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

func TestVisionProcessor_ExtractImageFeatures(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v",
		"max_tokens":  1000,
		"temperature": 0.5,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	processor := NewVisionProcessor(model, model.GetLogger(), nil)
	require.NotNil(t, processor)

	imageURL := "https://example.com/test-image.jpg"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	analysis, err := processor.ExtractImageFeatures(ctx, imageURL)

	// Should fail due to invalid URL, but validation should pass
	assert.Error(t, err)
	assert.Nil(t, analysis)
}

func TestVisionProcessor_GetConfig(t *testing.T) {
	model := NewZAIModel()
	logger := interfaces.NewDefaultLogger()

	config := &VisionConfig{
		MaxImageSize:   2048,
		DefaultQuality: ZAIVisionQualityHigh,
		DefaultDetail:  ZAIVisionDetailLow,
		Timeout:        30 * time.Second,
	}

	processor := NewVisionProcessor(model, logger, config)
	assert.NotNil(t, processor)

	retrievedConfig := processor.GetConfig()
	assert.Equal(t, config, retrievedConfig)
}

func TestVisionProcessor_UpdateConfig(t *testing.T) {
	model := NewZAIModel()
	logger := interfaces.NewDefaultLogger()

	processor := NewVisionProcessor(model, logger, nil)
	assert.NotNil(t, processor)

	// Update configuration
	newConfig := &VisionConfig{
		MaxImageSize:   4096,
		DefaultQuality: ZAIVisionQualityLow,
		DefaultDetail:  ZAIVisionDetailHigh,
		Timeout:        60 * time.Second,
	}

	processor.UpdateConfig(newConfig)

	// Verify updates
	updatedConfig := processor.GetConfig()
	assert.Equal(t, newConfig, updatedConfig)
}

func TestVisionProcessor_ParseFashionAnalysis(t *testing.T) {
	model := NewZAIModel()
	logger := interfaces.NewDefaultLogger()
	processor := NewVisionProcessor(model, logger, nil)
	require.NotNil(t, processor)

	// Create a mock response
	mockResponse := &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "This is a beautiful blue dress made of silk material. The style is elegant and suitable for formal occasions.",
				},
			},
		},
		Usage: &interfaces.PonchoUsage{
			PromptTokens:     50,
			CompletionTokens: 100,
			TotalTokens:      150,
		},
	}

	analysis, err := processor.parseFashionAnalysis(mockResponse)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "This is a beautiful blue dress made of silk material. The style is elegant and suitable for formal occasions.", analysis.Description)
	assert.Equal(t, 0.8, analysis.Confidence)
	assert.NotNil(t, analysis.Metadata)
	assert.Equal(t, "fashion", analysis.Metadata["analysis_type"])
}

func TestVisionProcessor_DownloadAndAnalyzeImage(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	model := NewZAIModel()
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v",
		"max_tokens":  1000,
		"temperature": 0.5,
	}

	err := model.Initialize(context.Background(), config)
	require.NoError(t, err)

	processor := NewVisionProcessor(model, model.GetLogger(), nil)
	require.NotNil(t, processor)

	imageURL := "https://example.com/test-image.jpg"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	analysis, err := processor.DownloadAndAnalyzeImage(ctx, imageURL)

	// Should fail due to invalid URL, but download attempt should be made
	assert.Error(t, err)
	assert.Nil(t, analysis)
}
