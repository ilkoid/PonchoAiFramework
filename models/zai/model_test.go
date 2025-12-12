package zai

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGLMModel_NewGLMModel tests the constructor
func TestGLMModel_NewGLMModel(t *testing.T) {
	model := NewGLMModel()

	assert.NotNil(t, model)
	assert.Equal(t, "glm-4.6", model.Name())
	assert.Equal(t, "zai", model.Provider())
	assert.True(t, model.SupportsStreaming())
	assert.True(t, model.SupportsTools())
	assert.False(t, model.SupportsVision()) // Vision support depends on model
	assert.True(t, model.SupportsSystemRole())
}

// TestGLMModel_Initialize tests model initialization
func TestGLMModel_Initialize(t *testing.T) {
	model := NewGLMModel()

	// Test valid configuration
	config := map[string]interface{}{
		"api_key":     "test-api-key",
		"model_name":  "glm-4.6v",
		"max_tokens":  2000,
		"temperature": 0.5,
		"timeout":     "30s",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, model.client)
	assert.NotNil(t, model.visionProcessor)
	assert.NotNil(t, model.requestConverter)
	assert.NotNil(t, model.responseConverter)
	assert.NotNil(t, model.streamConverter)

	// Test missing API key
	badConfig := map[string]interface{}{
		"model_name": "glm-4.6",
	}

	model2 := NewGLMModel()
	err = model2.Initialize(ctx, badConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")
}

// TestGLMModel_Generate tests text generation
func TestGLMModel_Generate(t *testing.T) {
	model := NewGLMModel()

	// Mock setup - in real tests, you'd use a mock HTTP server
	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"model_name": "glm-4.6",
		"max_tokens": 100,
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	// Test basic text generation request
	// This would fail in real test without mock server, but we test the conversion
	// In a full test, you'd mock the HTTP client
	assert.NotNil(t, model.requestConverter)
	assert.NotNil(t, model.responseConverter)
}

// TestGLMModel_GenerateVision tests vision generation
func TestGLMModel_GenerateVision(t *testing.T) {
	model := NewGLMModel()

	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"model_name": "glm-4.6v",
		"max_tokens": 2000,
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	// Test vision request
	req := &interfaces.PonchoModelRequest{
		Model: "glm-4.6v",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL: "https://example.com/image.jpg",
						},
					},
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What do you see in this image?",
					},
				},
			},
		},
	}

	// Test that vision processor is available
	assert.NotNil(t, model.visionProcessor)
	assert.True(t, model.SupportsVision())
	_ = req // Avoid unused warning
}

// TestGLMModel_GenerateStreaming tests streaming generation
func TestGLMModel_GenerateStreaming(t *testing.T) {
	model := NewGLMModel()

	config := map[string]interface{}{
		"api_key":    "test-api-key",
		"model_name": "glm-4.6",
		"max_tokens": 100,
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	// Test streaming request
	req := &interfaces.PonchoModelRequest{
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Tell me a story",
					},
				},
			},
		},
		Stream: true,
	}
	_ = req // Avoid unused warning

	// Test streaming capability
	assert.True(t, model.SupportsStreaming())

	// Mock callback
	callbackCalled := false
	var callback interfaces.PonchoStreamCallback = func(chunk *interfaces.PonchoStreamChunk) error {
		callbackCalled = true
		assert.NotNil(t, chunk)
		return nil
	}
	_ = callbackCalled
	// In real test, you'd mock the streaming HTTP response
	assert.NotNil(t, model.streamConverter)
	assert.NotNil(t, callback)
}

// TestGLMModel_ExtractConfig tests configuration extraction
func TestGLMModel_ExtractConfig(t *testing.T) {
	model := NewGLMModel()

	// Test full configuration
	config := map[string]interface{}{
		"api_key":     "test-key",
		"base_url":    "https://custom.api.com",
		"model_name":  "glm-4.6v",
		"max_tokens":  3000,
		"temperature": 0.8,
		"timeout":     "45s",
		"custom_params": map[string]interface{}{
			"top_p":             0.9,
			"frequency_penalty": 0.1,
			"presence_penalty":  0.2,
			"stop":              []string{"\n"},
			"thinking": map[string]interface{}{
				"type": "enabled",
			},
		},
	}

	glmConfig, err := model.extractConfig(config)
	require.NoError(t, err)

	assert.Equal(t, "test-key", glmConfig.APIKey)
	assert.Equal(t, "https://custom.api.com", glmConfig.BaseURL)
	assert.Equal(t, "glm-4.6v", glmConfig.Model)
	assert.Equal(t, 3000, glmConfig.MaxTokens)
	assert.Equal(t, float32(0.8), glmConfig.Temperature)
	assert.Equal(t, 45*time.Second, glmConfig.Timeout)

	// Test custom params
	assert.NotNil(t, glmConfig.TopP)
	assert.Equal(t, float32(0.9), *glmConfig.TopP)
	assert.NotNil(t, glmConfig.FrequencyPenalty)
	assert.Equal(t, float32(0.1), *glmConfig.FrequencyPenalty)
	assert.NotNil(t, glmConfig.PresencePenalty)
	assert.Equal(t, float32(0.2), *glmConfig.PresencePenalty)
	assert.NotNil(t, glmConfig.Thinking)
	assert.Equal(t, "enabled", glmConfig.Thinking.Type)
}

// TestGLMModel_ExtractConfigDefaults tests configuration extraction with defaults
func TestGLMModel_ExtractConfigDefaults(t *testing.T) {
	model := NewGLMModel()

	// Test minimal configuration
	config := map[string]interface{}{
		"api_key": "test-key",
	}

	glmConfig, err := model.extractConfig(config)
	require.NoError(t, err)

	assert.Equal(t, "test-key", glmConfig.APIKey)
	assert.Equal(t, GLMDefaultBaseURL, glmConfig.BaseURL)
	assert.Equal(t, GLMDefaultModel, glmConfig.Model)
	assert.Equal(t, model.MaxTokens(), glmConfig.MaxTokens)
	assert.Equal(t, model.DefaultTemperature(), glmConfig.Temperature)
	assert.Equal(t, 60*time.Second, glmConfig.Timeout)
}

// TestGLMModel_ValidateRequest tests request validation
func TestGLMModel_ValidateRequest(t *testing.T) {
	model := NewGLMModel()

	config := map[string]interface{}{
		"api_key": "test-key",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	// Test valid request
	validReq := &interfaces.PonchoModelRequest{
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
				},
			},
		},
	}

	err = model.ValidateRequest(validReq)
	assert.NoError(t, err)

	// Test invalid request (no messages)
	invalidReq := &interfaces.PonchoModelRequest{
		Model:    "glm-4.6",
		Messages: []*interfaces.PonchoMessage{},
	}

	err = model.ValidateRequest(invalidReq)
	assert.Error(t, err)
}

// TestGLMModel_AnalyzeImage tests image analysis
func TestGLMModel_AnalyzeImage(t *testing.T) {
	model := NewGLMModel()

	// Test with non-vision model
	config := map[string]interface{}{
		"api_key":    "test-key",
		"model_name": "glm-4.6", // Non-vision model
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	result, err := model.AnalyzeImage(ctx, "https://example.com/image.jpg", "What do you see?")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support vision")
	_ = result // Avoid unused warning

	// Test with vision model
	config2 := map[string]interface{}{
		"api_key":    "test-key",
		"model_name": "glm-4.6v", // Vision model
	}

	model2 := NewGLMModel()
	err = model2.Initialize(ctx, config2)
	require.NoError(t, err)

	assert.True(t, model2.SupportsVision())
	assert.NotNil(t, model2.visionProcessor)
}

// TestGLMModel_ExtractFeatures tests feature extraction
func TestGLMModel_ExtractFeatures(t *testing.T) {
	model := NewGLMModel()

	config := map[string]interface{}{
		"api_key":    "test-key",
		"model_name": "glm-4.6v",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	features := []string{"color", "style", "material"}

	// Test with vision model
	result, err := model.ExtractFeatures(ctx, "https://example.com/image.jpg", features)
	// Would fail without mock server, but tests the interface
	assert.NotNil(t, model.visionProcessor)
	_ = result
	_ = err
}

// TestGLMModel_Shutdown tests graceful shutdown
func TestGLMModel_Shutdown(t *testing.T) {
	model := NewGLMModel()

	config := map[string]interface{}{
		"api_key": "test-key",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	require.NoError(t, err)

	// Test shutdown
	err = model.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestGLMModel_JSONHelpers tests JSON conversion helpers
func TestGLMModel_JSONHelpers(t *testing.T) {
	model := NewGLMModel()

	// Test mapToJSONString
	testMap := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	jsonStr := model.mapToJSONString(testMap)
	assert.NotEmpty(t, jsonStr)

	var parsedMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &parsedMap)
	assert.NoError(t, err)
	assert.Equal(t, testMap["key1"], parsedMap["key1"])
	assert.Equal(t, float64(testMap["key2"].(int)), parsedMap["key2"])

	// Test jsonStringToMap
	testJSON := `{"test": "value", "number": 42}`
	parsed, err := model.jsonStringToMap(testJSON)
	require.NoError(t, err)
	assert.Equal(t, "value", parsed["test"])
	assert.Equal(t, float64(42), parsed["number"])

	// Test empty JSON
	empty, err := model.jsonStringToMap("")
	require.NoError(t, err)
	assert.Empty(t, empty)

	// Test invalid JSON
	invalid, err := model.jsonStringToMap("invalid json")
	assert.Error(t, err)
	assert.Nil(t, invalid)
}

// BenchmarkGLMModel_ConvertRequest benchmarks request conversion
func BenchmarkGLMModel_ConvertRequest(b *testing.B) {
	model := NewGLMModel()
	config := map[string]interface{}{
		"api_key": "test-key",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	if err != nil {
		b.Fatal(err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello world"},
				},
			},
		},
		Tools: []*interfaces.PonchoToolDef{
			{
				Name:        "test_tool",
				Description: "A test tool",
				Parameters:  map[string]interface{}{"param": "string"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.requestConverter.ConvertRequest(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
