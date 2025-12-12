package zai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// GLMModel represents a Z.AI GLM model implementation
type GLMModel struct {
	*base.PonchoBaseModel
	client            *GLMClient
	visionProcessor   *VisionProcessor
	requestConverter  *RequestConverter
	responseConverter *ResponseConverter
	streamConverter   *StreamConverter
}

// NewGLMModel creates a new GLM model instance
func NewGLMModel() *GLMModel {
	capabilities := interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    true, // GLM-4.6v supports vision
		System:    true,
	}

	baseModel := base.NewPonchoBaseModel("glm-4.6", "zai", capabilities)

	return &GLMModel{
		PonchoBaseModel:   baseModel,
		client:            nil, // Will be initialized in Initialize
		visionProcessor:   nil, // Will be initialized in Initialize
		requestConverter:  nil, // Will be initialized in Initialize
		responseConverter: nil, // Will be initialized in Initialize
		streamConverter:   nil, // Will be initialized in Initialize
	}
}

// Initialize initializes the GLM model with configuration
func (m *GLMModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Initialize base model first
	if err := m.PonchoBaseModel.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize base model: %w", err)
	}

	// Extract GLM-specific configuration
	glmConfig, err := m.extractConfig(config)
	if err != nil {
		return fmt.Errorf("failed to extract GLM config: %w", err)
	}

	// Create GLM client
	m.client = NewGLMClient(glmConfig, m.GetLogger())

	// Initialize converters
	m.requestConverter = NewRequestConverter(m.GetLogger())
	m.responseConverter = NewResponseConverter(m.GetLogger())
	m.streamConverter = NewStreamConverter(m.GetLogger())

	// Initialize vision processor for vision models
	if m.client.IsVisionModel(glmConfig.Model) {
		m.visionProcessor = NewVisionProcessor(m.client, m.GetLogger(), nil)
	}

	m.GetLogger().Info("GLM model initialized",
		"model", glmConfig.Model,
		"base_url", glmConfig.BaseURL,
		"max_tokens", glmConfig.MaxTokens,
		"temperature", glmConfig.Temperature,
		"vision_support", m.client.IsVisionModel(glmConfig.Model))

	return nil
}

// Generate generates a response using GLM API
func (m *GLMModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert Poncho request to GLM request
	glmReq, err := m.requestConverter.ConvertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Add thinking parameter for GLM-4.6 models
	if m.client.SupportsThinking(glmReq.Model) {
		glmReq.Thinking = &GLMThinking{
			Type: GLMThinkingEnabled,
		}
	}

	// Call GLM API
	glmResp, err := m.client.CreateChatCompletion(ctx, glmReq)
	if err != nil {
		return nil, fmt.Errorf("GLM API call failed: %w", err)
	}

	// Convert GLM response to Poncho response
	ponchoResp, err := m.responseConverter.ConvertResponse(glmResp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	m.GetLogger().Debug("Generated response",
		"model", glmResp.Model,
		"finish_reason", glmResp.Choices[0].FinishReason,
		"prompt_tokens", glmResp.Usage.PromptTokens,
		"completion_tokens", glmResp.Usage.CompletionTokens)

	return ponchoResp, nil
}

// GenerateStreaming generates a streaming response using GLM API
func (m *GLMModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Check if streaming is supported
	if !m.SupportsStreaming() {
		return fmt.Errorf("model '%s' does not support streaming", m.Name())
	}

	// Convert Poncho request to GLM request
	glmReq, err := m.requestConverter.ConvertRequest(req)
	if err != nil {
		return fmt.Errorf("failed to convert request: %w", err)
	}

	// Add thinking parameter for GLM-4.6 models
	if m.client.SupportsThinking(glmReq.Model) {
		glmReq.Thinking = &GLMThinking{
			Type: GLMThinkingEnabled,
		}
	}

	// Call GLM streaming API
	return m.client.CreateChatCompletionStream(ctx, glmReq, func(streamResp *GLMStreamResponse) error {
		// Convert streaming response to Poncho format
		chunk, err := m.streamConverter.ConvertStreamChunk(streamResp)
		if err != nil {
			return fmt.Errorf("failed to convert stream chunk: %w", err)
		}

		// Call the callback
		return callback(chunk)
	})
}

// SupportsVision returns true if the model supports vision capabilities
func (m *GLMModel) SupportsVision() bool {
	return m.visionProcessor != nil
}

// AnalyzeImage performs vision analysis on an image
func (m *GLMModel) AnalyzeImage(ctx context.Context, imageURL string, prompt string) (*FashionAnalysis, error) {
	if !m.SupportsVision() {
		return nil, fmt.Errorf("model '%s' does not support vision", m.Name())
	}

	if m.visionProcessor == nil {
		return nil, fmt.Errorf("vision processor not initialized")
	}

	return m.visionProcessor.AnalyzeFashionImage(ctx, imageURL, prompt)
}

// ExtractFeatures extracts specific features from an image
func (m *GLMModel) ExtractFeatures(ctx context.Context, imageURL string, features []string) (map[string]interface{}, error) {
	if !m.SupportsVision() {
		return nil, fmt.Errorf("model '%s' does not support vision", m.Name())
	}

	if m.visionProcessor == nil {
		return nil, fmt.Errorf("vision processor not initialized")
	}

	return m.visionProcessor.ExtractProductFeatures(ctx, imageURL, features)
}

// Shutdown shuts down the GLM model and cleans up resources
func (m *GLMModel) Shutdown(ctx context.Context) error {
	// Close client if it exists
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			m.GetLogger().Warn("Failed to close GLM client", "error", err)
		}
		m.client = nil
	}

	// Shutdown base model
	return m.PonchoBaseModel.Shutdown(ctx)
}

// extractConfig extracts GLM configuration from generic config map
func (m *GLMModel) extractConfig(config map[string]interface{}) (*GLMConfig, error) {
	glmConfig := &GLMConfig{
		BaseURL:     GLMDefaultBaseURL,
		Model:       GLMDefaultModel,
		MaxTokens:   m.MaxTokens(),
		Temperature: m.DefaultTemperature(),
		Timeout:     60 * time.Second,
	}

	// Extract API key (required)
	if apiKey, ok := config["api_key"].(string); ok && apiKey != "" {
		glmConfig.APIKey = apiKey
	} else {
		return nil, fmt.Errorf("api_key is required for GLM model")
	}

	// Extract optional fields
	if baseURL, ok := config["base_url"].(string); ok && baseURL != "" {
		glmConfig.BaseURL = baseURL
	}

	if modelName, ok := config["model_name"].(string); ok && modelName != "" {
		glmConfig.Model = modelName
	}

	if maxTokens, ok := config["max_tokens"].(int); ok && maxTokens > 0 {
		glmConfig.MaxTokens = maxTokens
	}

	if temperature, ok := config["temperature"].(float32); ok {
		glmConfig.Temperature = temperature
	} else if temperature, ok := config["temperature"].(float64); ok {
		glmConfig.Temperature = float32(temperature)
	}

	if timeoutStr, ok := config["timeout"].(string); ok && timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			glmConfig.Timeout = timeout
		}
	}

	// Extract custom parameters
	if customParams, ok := config["custom_params"].(map[string]interface{}); ok {
		if topP, ok := customParams["top_p"].(float32); ok {
			glmConfig.TopP = &topP
		} else if topP, ok := customParams["top_p"].(float64); ok {
			tp := float32(topP)
			glmConfig.TopP = &tp
		}

		if freqPenalty, ok := customParams["frequency_penalty"].(float32); ok {
			glmConfig.FrequencyPenalty = &freqPenalty
		} else if freqPenalty, ok := customParams["frequency_penalty"].(float64); ok {
			fp := float32(freqPenalty)
			glmConfig.FrequencyPenalty = &fp
		}

		if presPenalty, ok := customParams["presence_penalty"].(float32); ok {
			glmConfig.PresencePenalty = &presPenalty
		} else if presPenalty, ok := customParams["presence_penalty"].(float64); ok {
			pp := float32(presPenalty)
			glmConfig.PresencePenalty = &pp
		}

		if stop, ok := customParams["stop"]; ok {
			glmConfig.Stop = stop
		}

		// Thinking mode
		if thinking, ok := customParams["thinking"].(map[string]interface{}); ok {
			if thinkingType, ok := thinking["type"].(string); ok {
				glmConfig.Thinking = &GLMThinking{
					Type: thinkingType,
				}
			}
		}
	}

	return glmConfig, nil
}

// Helper methods for JSON conversion
func (m *GLMModel) mapToJSONString(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	// Simple JSON marshaling - in production, you might want better error handling
	if bytes, err := json.Marshal(data); err == nil {
		return string(bytes)
	}
	return "{}"
}

func (m *GLMModel) jsonStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}
