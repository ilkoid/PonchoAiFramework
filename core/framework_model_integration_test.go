package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockDeepSeekModel simulates DeepSeek model behavior for testing
type MockDeepSeekModel struct {
	*base.PonchoBaseModel
	apiKey       string
	baseURL      string
	generateFunc func(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error)
}

func NewMockDeepSeekModel(name, apiKey, baseURL string) *MockDeepSeekModel {
	baseModel := base.NewPonchoBaseModel(name, "deepseek", interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    false,
		System:    true,
		JSONMode:  true,
	})

	return &MockDeepSeekModel{
		PonchoBaseModel: baseModel,
		apiKey:          apiKey,
		baseURL:         baseURL,
	}
}

func (m *MockDeepSeekModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	if apiKey, ok := config["api_key"].(string); ok {
		m.apiKey = apiKey
	}
	if baseURL, ok := config["base_url"].(string); ok {
		m.baseURL = baseURL
	}
	return nil
}

func (m *MockDeepSeekModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}

	// Simulate DeepSeek response
	response := &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: fmt.Sprintf("DeepSeek response to: %s", m.extractTextFromRequest(req)),
				},
			},
		},
		Usage: &interfaces.PonchoUsage{
			PromptTokens:     m.countTokens(req),
			CompletionTokens: 50,
			TotalTokens:      m.countTokens(req) + 50,
		},
		FinishReason: interfaces.PonchoFinishReasonStop,
	}

	return response, nil
}

func (m *MockDeepSeekModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	// Simulate streaming response
	chunks := []string{
		"DeepSeek",
		" streaming",
		" response",
		" to: ",
		m.extractTextFromRequest(req),
	}

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			streamChunk := &interfaces.PonchoStreamChunk{
				Delta: &interfaces.PonchoMessage{
					Role: interfaces.PonchoRoleAssistant,
					Content: []*interfaces.PonchoContentPart{
						{
							Type: interfaces.PonchoContentTypeText,
							Text: chunk,
						},
					},
				},
				Done: i == len(chunks)-1,
			}

			if err := callback(streamChunk); err != nil {
				return err
			}
			time.Sleep(10 * time.Millisecond) // Simulate network delay
		}
	}

	return nil
}

func (m *MockDeepSeekModel) extractTextFromRequest(req *interfaces.PonchoModelRequest) string {
	if len(req.Messages) > 0 && len(req.Messages[0].Content) > 0 {
		for _, content := range req.Messages[0].Content {
			if content.Type == interfaces.PonchoContentTypeText {
				return content.Text
			}
		}
	}
	return "unknown"
}

func (m *MockDeepSeekModel) countTokens(req *interfaces.PonchoModelRequest) int {
	// Simple token counting simulation
	text := m.extractTextFromRequest(req)
	return len(text) / 4 // Rough estimate
}

// MockZAIModel simulates Z.AI GLM model behavior for testing
type MockZAIModel struct {
	*base.PonchoBaseModel
	apiKey         string
	baseURL        string
	supportsVision bool
	generateFunc   func(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error)
}

func NewMockZAIModel(name, apiKey, baseURL string, supportsVision bool) *MockZAIModel {
	baseModel := base.NewPonchoBaseModel(name, "zai", interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    supportsVision,
		System:    true,
		JSONMode:  true,
	})

	return &MockZAIModel{
		PonchoBaseModel: baseModel,
		apiKey:          apiKey,
		baseURL:         baseURL,
		supportsVision:  supportsVision,
	}
}

func (m *MockZAIModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	if apiKey, ok := config["api_key"].(string); ok {
		m.apiKey = apiKey
	}
	if baseURL, ok := config["base_url"].(string); ok {
		m.baseURL = baseURL
	}
	return nil
}

func (m *MockZAIModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}

	// Check for vision content
	hasVision := m.hasVisionContent(req)
	var responseText string

	if hasVision {
		responseText = m.generateVisionResponse(req)
	} else {
		responseText = fmt.Sprintf("Z.AI GLM response to: %s", m.extractTextFromRequest(req))
	}

	response := &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: responseText,
				},
			},
		},
		Usage: &interfaces.PonchoUsage{
			PromptTokens:     m.countTokens(req),
			CompletionTokens: 60,
			TotalTokens:      m.countTokens(req) + 60,
		},
		FinishReason: interfaces.PonchoFinishReasonStop,
	}

	return response, nil
}

func (m *MockZAIModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	// Simulate streaming response
	hasVision := m.hasVisionContent(req)
	var responseText string

	if hasVision {
		responseText = m.generateVisionResponse(req)
	} else {
		responseText = fmt.Sprintf("Z.AI GLM streaming response to: %s", m.extractTextFromRequest(req))
	}

	chunks := []string{
		"Z.AI",
		" GLM",
		" streaming",
		" response",
		": ",
		responseText,
	}

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			streamChunk := &interfaces.PonchoStreamChunk{
				Delta: &interfaces.PonchoMessage{
					Role: interfaces.PonchoRoleAssistant,
					Content: []*interfaces.PonchoContentPart{
						{
							Type: interfaces.PonchoContentTypeText,
							Text: chunk,
						},
					},
				},
				Done: i == len(chunks)-1,
			}

			if err := callback(streamChunk); err != nil {
				return err
			}
			time.Sleep(15 * time.Millisecond) // Simulate network delay
		}
	}

	return nil
}

func (m *MockZAIModel) hasVisionContent(req *interfaces.PonchoModelRequest) bool {
	for _, message := range req.Messages {
		for _, content := range message.Content {
			if content.Type == interfaces.PonchoContentTypeMedia {
				return true
			}
		}
	}
	return false
}

func (m *MockZAIModel) generateVisionResponse(req *interfaces.PonchoModelRequest) string {
	// Simulate fashion-specific vision analysis
	return "This appears to be a fashion item with the following characteristics: " +
		"Style: Casual, Material: Cotton blend, Season: Spring/Summer, " +
		"Colors: Blue and white pattern, Type: Shirt with collar"
}

func (m *MockZAIModel) extractTextFromRequest(req *interfaces.PonchoModelRequest) string {
	if len(req.Messages) > 0 && len(req.Messages[0].Content) > 0 {
		for _, content := range req.Messages[0].Content {
			if content.Type == interfaces.PonchoContentTypeText {
				return content.Text
			}
		}
	}
	return "unknown"
}

func (m *MockZAIModel) countTokens(req *interfaces.PonchoModelRequest) int {
	// Simple token counting simulation
	text := m.extractTextFromRequest(req)
	return len(text) / 4 // Rough estimate
}

// Test configuration for integration tests
func createTestConfig() *interfaces.PonchoFrameworkConfig {
	return &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:    "deepseek",
				ModelName:   "deepseek-chat",
				APIKey:      "test-deepseek-key",
				BaseURL:     "https://api.deepseek.com/v1",
				MaxTokens:   4000,
				Temperature: 0.7,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
					JSONMode:  true,
				},
			},
			"glm-vision": {
				Provider:    "zai",
				ModelName:   "glm-4.6v",
				APIKey:      "test-zai-key",
				BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
				MaxTokens:   2000,
				Temperature: 0.5,
				Timeout:     "60s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    true,
					System:    true,
					JSONMode:  false,
				},
			},
		},
	}
}

// Test configuration loading with model providers
func TestFrameworkModelConfigurationLoading(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := createTestConfig()

	// Create framework with custom config manager that uses our test config
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Manually register models from test config since we're not loading from file
	for name, modelConfig := range config.Models {
		// Create mock models based on provider
		var model interfaces.PonchoModel
		if modelConfig.Provider == "deepseek" {
			model = NewMockDeepSeekModel(name, modelConfig.APIKey, modelConfig.BaseURL)
		} else if modelConfig.Provider == "zai" {
			model = NewMockZAIModel(name, modelConfig.APIKey, modelConfig.BaseURL, modelConfig.Supports.Vision)
		}

		if model != nil {
			err = framework.RegisterModel(name, model)
			if err != nil {
				t.Fatalf("Expected model registration to succeed for %s, got error: %v", name, err)
			}
		}
	}

	// Verify models are loaded from configuration
	registry := framework.GetModelRegistry()
	models := registry.List()

	expectedModels := []string{"deepseek-chat", "glm-vision"}
	if len(models) != len(expectedModels) {
		t.Errorf("Expected %d models, got %d", len(expectedModels), len(models))
	}

	for _, expectedModel := range expectedModels {
		found := false
		for _, model := range models {
			if model == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s not found in registry", expectedModel)
		}
	}

	// Verify model configurations
	deepseekModel, err := registry.Get("deepseek-chat")
	if err != nil {
		t.Fatalf("Expected to get deepseek-chat model, got error: %v", err)
	}

	if deepseekModel.Provider() != "deepseek" {
		t.Errorf("Expected deepseek provider, got %s", deepseekModel.Provider())
	}

	if deepseekModel.SupportsVision() {
		t.Error("DeepSeek model should not support vision")
	}

	if !deepseekModel.SupportsStreaming() {
		t.Error("DeepSeek model should support streaming")
	}

	glmModel, err := registry.Get("glm-vision")
	if err != nil {
		t.Fatalf("Expected to get glm-vision model, got error: %v", err)
	}

	if glmModel.Provider() != "zai" {
		t.Errorf("Expected zai provider, got %s", glmModel.Provider())
	}

	if !glmModel.SupportsVision() {
		t.Error("Z.AI GLM model should support vision")
	}

	if !glmModel.SupportsStreaming() {
		t.Error("Z.AI GLM model should support streaming")
	}
}

// Test model registration and execution through framework
func TestFrameworkModelRegistrationAndExecution(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Register mock DeepSeek model
	deepseekModel := NewMockDeepSeekModel("deepseek-chat", "test-key", "https://api.deepseek.com/v1")
	err = framework.RegisterModel("deepseek-chat", deepseekModel)
	if err != nil {
		t.Fatalf("Expected DeepSeek model registration to succeed, got error: %v", err)
	}

	// Register mock Z.AI model
	glmModel := NewMockZAIModel("glm-vision", "test-key", "https://open.bigmodel.cn/api/paas/v4", true)
	err = framework.RegisterModel("glm-vision", glmModel)
	if err != nil {
		t.Fatalf("Expected Z.AI model registration to succeed, got error: %v", err)
	}

	// Test DeepSeek model execution
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, DeepSeek!",
					},
				},
			},
		},
	}

	response, err := framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected DeepSeek generation to succeed, got error: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	if response.Message == nil {
		t.Error("Expected message to be non-nil")
	}

	if response.Usage == nil {
		t.Error("Expected usage to be non-nil")
	}

	if response.Usage.TotalTokens <= 0 {
		t.Error("Expected total tokens to be positive")
	}

	// Test Z.AI model execution
	req.Model = "glm-vision"
	req.Messages[0].Content[0].Text = "Hello, Z.AI!"

	response, err = framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected Z.AI generation to succeed, got error: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	if !strings.Contains(response.Message.Content[0].Text, "Z.AI GLM response") {
		t.Error("Expected Z.AI response to contain model identifier")
	}
}

// Test streaming functionality
func TestFrameworkModelStreaming(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Register mock models
	deepseekModel := NewMockDeepSeekModel("deepseek-chat", "test-key", "https://api.deepseek.com/v1")
	err = framework.RegisterModel("deepseek-chat", deepseekModel)
	if err != nil {
		t.Fatalf("Expected DeepSeek model registration to succeed, got error: %v", err)
	}

	glmModel := NewMockZAIModel("glm-vision", "test-key", "https://open.bigmodel.cn/api/paas/v4", true)
	err = framework.RegisterModel("glm-vision", glmModel)
	if err != nil {
		t.Fatalf("Expected Z.AI model registration to succeed, got error: %v", err)
	}

	// Test DeepSeek streaming
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Stream response, DeepSeek!",
					},
				},
			},
		},
	}

	var chunks []string
	callback := func(chunk *interfaces.PonchoStreamChunk) error {
		if chunk == nil {
			return fmt.Errorf("chunk is nil")
		}

		if chunk.Delta == nil || len(chunk.Delta.Content) == 0 {
			return fmt.Errorf("chunk delta is empty")
		}

		chunks = append(chunks, chunk.Delta.Content[0].Text)
		return nil
	}

	err = framework.GenerateStreaming(ctx, req, callback)
	if err != nil {
		t.Fatalf("Expected DeepSeek streaming to succeed, got error: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one streaming chunk")
	}

	// Verify final chunk is marked as done
	if len(chunks) > 0 && !strings.Contains(chunks[len(chunks)-1], "Stream response, DeepSeek!") {
		t.Error("Expected final chunk to contain complete response")
	}

	// Test Z.AI streaming
	req.Model = "glm-vision"
	req.Messages[0].Content[0].Text = "Stream response, Z.AI!"
	chunks = []string{}

	err = framework.GenerateStreaming(ctx, req, callback)
	if err != nil {
		t.Fatalf("Expected Z.AI streaming to succeed, got error: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one streaming chunk")
	}

	// Verify Z.AI specific content
	var fullResponse string
	for _, chunk := range chunks {
		fullResponse += chunk
	}

	if !contains(fullResponse, "Z.AI GLM") {
		t.Error("Expected Z.AI streaming response to contain model identifier")
	}
}

// Test fashion-specific vision scenarios
func TestFrameworkFashionVisionScenarios(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Register Z.AI vision model
	glmModel := NewMockZAIModel("glm-vision", "test-key", "https://open.bigmodel.cn/api/paas/v4", true)
	err = framework.RegisterModel("glm-vision", glmModel)
	if err != nil {
		t.Fatalf("Expected Z.AI model registration to succeed, got error: %v", err)
	}

	// Create test image data (base64 encoded)
	testImageData := base64.StdEncoding.EncodeToString([]byte("fake-image-data"))

	// Test fashion image analysis
	req := &interfaces.PonchoModelRequest{
		Model: "glm-vision",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Analyze this fashion item and describe its characteristics",
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      "data:image/jpeg;base64," + testImageData,
							MimeType: "image/jpeg",
						},
					},
				},
			},
		},
	}

	t.Logf("Request has vision content: %v", glmModel.hasVisionContent(req))
	response, err := framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected fashion vision analysis to succeed, got error: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	// Verify fashion-specific analysis
	responseText := response.Message.Content[0].Text
	t.Logf("Fashion vision response: %s", responseText)
	t.Logf("Response length: %d", len(responseText))

	if !strings.Contains(responseText, "fashion") {
		t.Errorf("Expected response to contain fashion-related analysis, got: %s", responseText)
	}

	if !strings.Contains(responseText, "Style") {
		t.Errorf("Expected response to contain style analysis, got: %s", responseText)
	}

	if !strings.Contains(responseText, "Material") {
		t.Errorf("Expected response to contain material analysis, got: %s", responseText)
	}

	if !strings.Contains(responseText, "Season") {
		t.Errorf("Expected response to contain season analysis, got: %s", responseText)
	}

	// Test streaming fashion analysis
	var streamingChunks []string
	streamingCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		if chunk != nil && chunk.Delta != nil && len(chunk.Delta.Content) > 0 {
			streamingChunks = append(streamingChunks, chunk.Delta.Content[0].Text)
		}
		return nil
	}

	err = framework.GenerateStreaming(ctx, req, streamingCallback)
	if err != nil {
		t.Fatalf("Expected fashion vision streaming to succeed, got error: %v", err)
	}

	if len(streamingChunks) == 0 {
		t.Error("Expected at least one streaming chunk for fashion analysis")
	}

	// Verify streaming contains fashion content
	var fullStreamingResponse string
	for _, chunk := range streamingChunks {
		fullStreamingResponse += chunk
	}

	if !strings.Contains(fullStreamingResponse, "fashion") {
		t.Error("Expected streaming response to contain fashion-related analysis")
	}
}

// Test error handling and validation
func TestFrameworkModelErrorHandling(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Test generation with non-existent model
	req := &interfaces.PonchoModelRequest{
		Model: "non-existent-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello!",
					},
				},
			},
		},
	}

	_, err = framework.Generate(ctx, req)
	if err == nil {
		t.Error("Expected generation to fail for non-existent model")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Error("Expected error to mention model not found")
	}

	// Test streaming with non-existent model
	err = framework.GenerateStreaming(ctx, req, func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	})
	if err == nil {
		t.Error("Expected streaming to fail for non-existent model")
	}

	// Test streaming with model that doesn't support streaming
	nonStreamingModel := NewMockNonStreamingModel()
	err = framework.RegisterModel("non-streaming", nonStreamingModel)
	if err != nil {
		t.Fatalf("Expected non-streaming model registration to succeed, got error: %v", err)
	}

	req.Model = "non-streaming"
	err = framework.GenerateStreaming(ctx, req, func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	})
	if err == nil {
		t.Error("Expected streaming to fail for non-streaming model")
	}

	if !strings.Contains(err.Error(), "does not support streaming") {
		t.Error("Expected error to mention streaming not supported")
	}

	// Test framework not started
	stoppedFramework := NewPonchoFramework(config, logger)
	_, err = stoppedFramework.Generate(ctx, req)
	if err == nil {
		t.Error("Expected generation to fail when framework not started")
	}

	// Verify error metrics
	metrics, err := framework.Metrics(ctx)
	if err != nil {
		t.Fatalf("Expected metrics retrieval to succeed, got error: %v", err)
	}

	if metrics.Errors == nil {
		t.Error("Expected error metrics to be recorded")
	}

	if metrics.Errors.TotalErrors == 0 {
		t.Error("Expected at least one error to be recorded")
	}
}

// Test concurrent execution and model switching
func TestFrameworkConcurrentExecution(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Register mock models
	deepseekModel := NewMockDeepSeekModel("deepseek-chat", "test-key", "https://api.deepseek.com/v1")
	err = framework.RegisterModel("deepseek-chat", deepseekModel)
	if err != nil {
		t.Fatalf("Expected DeepSeek model registration to succeed, got error: %v", err)
	}

	glmModel := NewMockZAIModel("glm-vision", "test-key", "https://open.bigmodel.cn/api/paas/v4", true)
	err = framework.RegisterModel("glm-vision", glmModel)
	if err != nil {
		t.Fatalf("Expected Z.AI model registration to succeed, got error: %v", err)
	}

	// Test concurrent executions
	numGoroutines := 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]string, 0, numGoroutines)
	errors := make([]error, 0, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Alternate between models
			modelName := "deepseek-chat"
			if id%2 == 0 {
				modelName = "glm-vision"
			}

			req := &interfaces.PonchoModelRequest{
				Model: modelName,
				Messages: []*interfaces.PonchoMessage{
					{
						Role: interfaces.PonchoRoleUser,
						Content: []*interfaces.PonchoContentPart{
							{
								Type: interfaces.PonchoContentTypeText,
								Text: fmt.Sprintf("Concurrent request %d", id),
							},
						},
					},
				},
			}

			response, err := framework.Generate(ctx, req)
			mu.Lock()
			if err != nil {
				errors = append(errors, err)
			} else {
				results = append(results, response.Message.Content[0].Text)
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify results
	if len(errors) > 0 {
		t.Errorf("Expected no errors in concurrent execution, got %d errors", len(errors))
		for _, err := range errors {
			t.Logf("Concurrent execution error: %v", err)
		}
	}

	if len(results) != numGoroutines {
		t.Errorf("Expected %d results, got %d", numGoroutines, len(results))
	}

	// Verify model switching worked correctly
	deepseekCount := 0
	glmCount := 0
	for _, result := range results {
		if contains(result, "DeepSeek") {
			deepseekCount++
		} else if contains(result, "Z.AI GLM") {
			glmCount++
		}
	}

	expectedDeepseek := numGoroutines / 2
	if numGoroutines%2 == 1 {
		expectedDeepseek++ // Odd number, DeepSeek gets one more
	}

	if deepseekCount != expectedDeepseek {
		t.Errorf("Expected %d DeepSeek responses, got %d", expectedDeepseek, deepseekCount)
	}

	if glmCount != numGoroutines/2 {
		t.Errorf("Expected %d Z.AI responses, got %d", numGoroutines/2, glmCount)
	}
}

// Test resource cleanup and lifecycle management
func TestFrameworkResourceCleanup(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Start framework
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework start to succeed, got error: %v", err)
	}

	// Register models
	deepseekModel := NewMockDeepSeekModel("deepseek-chat", "test-key", "https://api.deepseek.com/v1")
	err = framework.RegisterModel("deepseek-chat", deepseekModel)
	if err != nil {
		t.Fatalf("Expected DeepSeek model registration to succeed, got error: %v", err)
	}

	glmModel := NewMockZAIModel("glm-vision", "test-key", "https://open.bigmodel.cn/api/paas/v4", true)
	err = framework.RegisterModel("glm-vision", glmModel)
	if err != nil {
		t.Fatalf("Expected Z.AI model registration to succeed, got error: %v", err)
	}

	// Execute some requests to ensure models are active
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Test before shutdown",
					},
				},
			},
		},
	}

	_, err = framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected generation to succeed before shutdown, got error: %v", err)
	}

	// Test framework shutdown
	err = framework.Stop(ctx)
	if err != nil {
		t.Fatalf("Expected framework stop to succeed, got error: %v", err)
	}

	// Verify framework is stopped
	health, err := framework.Health(ctx)
	if err != nil {
		t.Fatalf("Expected health check to succeed, got error: %v", err)
	}

	if health.Status != "unhealthy" {
		t.Errorf("Expected unhealthy status after stop, got %s", health.Status)
	}

	// Test operations after shutdown
	_, err = framework.Generate(ctx, req)
	if err == nil {
		t.Error("Expected generation to fail after framework shutdown")
	}

	// Test restart functionality
	err = framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected framework restart to succeed, got error: %v", err)
	}

	// Verify models are still available after restart
	registry := framework.GetModelRegistry()
	models := registry.List()
	if len(models) != 2 {
		t.Errorf("Expected 2 models after restart, got %d", len(models))
	}

	// Test generation after restart
	_, err = framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected generation to succeed after restart, got error: %v", err)
	}
}

// MockNonStreamingModel for testing streaming limitations
type MockNonStreamingModel struct {
	*base.PonchoBaseModel
}

func NewMockNonStreamingModel() *MockNonStreamingModel {
	baseModel := base.NewPonchoBaseModel("non-streaming", "test", interfaces.ModelCapabilities{
		Streaming: false,
		Tools:     false,
		Vision:    false,
		System:    false,
	})

	return &MockNonStreamingModel{
		PonchoBaseModel: baseModel,
	}
}

func (m *MockNonStreamingModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	return &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Non-streaming response",
				},
			},
		},
		Usage: &interfaces.PonchoUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
		FinishReason: interfaces.PonchoFinishReasonStop,
	}, nil
}

func (m *MockNonStreamingModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	return fmt.Errorf("model does not support streaming")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
