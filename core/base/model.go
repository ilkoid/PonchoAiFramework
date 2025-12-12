package base

import (
	"context"
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoBaseModel provides a base implementation of the PonchoModel interface
type PonchoBaseModel struct {
	name               string
	provider           string
	maxTokens          int
	defaultTemperature float32
	capabilities       interfaces.ModelCapabilities
	config             map[string]interface{}
	logger             interfaces.Logger
	mutex              sync.RWMutex
	initialized        bool
}

// NewPonchoBaseModel creates a new base model instance
func NewPonchoBaseModel(name, provider string, capabilities interfaces.ModelCapabilities) *PonchoBaseModel {
	return &PonchoBaseModel{
		name:               name,
		provider:           provider,
		maxTokens:          4000,
		defaultTemperature: 0.7,
		capabilities:       capabilities,
		config:             make(map[string]interface{}),
		logger:             interfaces.NewDefaultLogger(),
	}
}

// Name returns the model name
func (m *PonchoBaseModel) Name() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.name
}

// Provider returns the model provider
func (m *PonchoBaseModel) Provider() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.provider
}

// MaxTokens returns the maximum number of tokens supported by the model
func (m *PonchoBaseModel) MaxTokens() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.maxTokens
}

// DefaultTemperature returns the default temperature for the model
func (m *PonchoBaseModel) DefaultTemperature() float32 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.defaultTemperature
}

// SupportsStreaming returns whether the model supports streaming
func (m *PonchoBaseModel) SupportsStreaming() bool {
	return m.capabilities.Streaming
}

// SupportsTools returns whether the model supports tool calling
func (m *PonchoBaseModel) SupportsTools() bool {
	return m.capabilities.Tools
}

// SupportsVision returns whether the model supports vision capabilities
func (m *PonchoBaseModel) SupportsVision() bool {
	return m.capabilities.Vision
}

// SupportsSystemRole returns whether the model supports system role
func (m *PonchoBaseModel) SupportsSystemRole() bool {
	return m.capabilities.System
}

// Initialize initializes the model with configuration
func (m *PonchoBaseModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.initialized {
		return fmt.Errorf("model '%s' is already initialized", m.name)
	}

	// Validate configuration
	if err := m.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config for model '%s': %w", m.name, err)
	}

	// Set configuration
	if config != nil {
		m.config = make(map[string]interface{})
		for k, v := range config {
			m.config[k] = v
		}
	}

	// Extract model-specific configuration
	if maxTokens, ok := config["max_tokens"].(int); ok {
		m.maxTokens = maxTokens
	}

	if temperature, ok := config["temperature"].(float32); ok {
		m.defaultTemperature = temperature
	}

	// Extract temperature from float64 as well
	if temperature, ok := config["temperature"].(float64); ok {
		m.defaultTemperature = float32(temperature)
	}

	m.initialized = true
	m.logger.Info("Model initialized", "model", m.name, "provider", m.provider)

	return nil
}

// Shutdown shuts down the model and cleans up resources
func (m *PonchoBaseModel) Shutdown(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return nil
	}

	// Clear configuration
	m.config = make(map[string]interface{})
	m.initialized = false

	m.logger.Info("Model shutdown", "model", m.name, "provider", m.provider)

	return nil
}

// Generate generates a response (must be implemented by concrete models)
func (m *PonchoBaseModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if !m.isInitialized() {
		return nil, fmt.Errorf("model '%s' is not initialized", m.name)
	}

	// This is a base implementation - concrete models should override this
	return nil, fmt.Errorf("Generate method must be implemented by concrete model")
}

// GenerateStreaming generates a streaming response (must be implemented by concrete models)
func (m *PonchoBaseModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	if !m.isInitialized() {
		return fmt.Errorf("model '%s' is not initialized", m.name)
	}

	if !m.SupportsStreaming() {
		return fmt.Errorf("model '%s' does not support streaming", m.name)
	}

	// This is a base implementation - concrete models should override this
	return fmt.Errorf("GenerateStreaming method must be implemented by concrete model")
}

// validateConfig validates the model configuration (can be overridden by concrete models)
func (m *PonchoBaseModel) validateConfig(config map[string]interface{}) error {
	// Basic validation - concrete models can override this for specific requirements
	if config == nil {
		return nil
	}

	// Validate required fields if any
	// For example, check for API key if needed
	if apiKey, ok := config["api_key"].(string); ok && apiKey == "" {
		return fmt.Errorf("api_key cannot be empty")
	}

	// Validate max_tokens if provided
	if maxTokens, ok := config["max_tokens"].(int); ok {
		if maxTokens <= 0 {
			return fmt.Errorf("max_tokens must be positive")
		}
		if maxTokens > 32000 { // Reasonable upper limit
			return fmt.Errorf("max_tokens exceeds maximum allowed value")
		}
	}

	// Validate temperature if provided
	if temperature, ok := config["temperature"].(float32); ok {
		if temperature < 0 || temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	if temperature, ok := config["temperature"].(float64); ok {
		if temperature < 0 || temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	return nil
}

// GetConfig gets a configuration value by key
func (m *PonchoBaseModel) GetConfig(key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	value, exists := m.config[key]
	return value, exists
}

// SetConfig sets a configuration value by key
func (m *PonchoBaseModel) SetConfig(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.config == nil {
		m.config = make(map[string]interface{})
	}

	m.config[key] = value
}

// GetAllConfig returns a copy of all configuration
func (m *PonchoBaseModel) GetAllConfig() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config := make(map[string]interface{})
	for k, v := range m.config {
		config[k] = v
	}

	return config
}

// SetLogger sets the logger for the model
func (m *PonchoBaseModel) SetLogger(logger interfaces.Logger) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logger = logger
}

// GetLogger returns the current logger
func (m *PonchoBaseModel) GetLogger() interfaces.Logger {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.logger
}

// SetMaxTokens sets the maximum number of tokens
func (m *PonchoBaseModel) SetMaxTokens(maxTokens int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.maxTokens = maxTokens
}

// SetDefaultTemperature sets the default temperature
func (m *PonchoBaseModel) SetDefaultTemperature(temperature float32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.defaultTemperature = temperature
}

// SetCapabilities sets the model capabilities
func (m *PonchoBaseModel) SetCapabilities(capabilities interfaces.ModelCapabilities) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.capabilities = capabilities
}

// GetCapabilities returns the current capabilities
func (m *PonchoBaseModel) GetCapabilities() interfaces.ModelCapabilities {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.capabilities
}

// isInitialized checks if the model is initialized (internal method)
func (m *PonchoBaseModel) isInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.initialized
}

// ValidateRequest validates a model request
func (m *PonchoBaseModel) ValidateRequest(req *interfaces.PonchoModelRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Messages == nil || len(req.Messages) == 0 {
		return fmt.Errorf("request must contain at least one message")
	}

	// Validate messages
	for i, msg := range req.Messages {
		if msg == nil {
			return fmt.Errorf("message %d cannot be nil", i)
		}

		if msg.Role == "" {
			return fmt.Errorf("message %d must have a role", i)
		}

		if msg.Content == nil || len(msg.Content) == 0 {
			return fmt.Errorf("message %d must have content", i)
		}

		// Validate content parts
		for j, part := range msg.Content {
			if part == nil {
				return fmt.Errorf("content part %d in message %d cannot be nil", j, i)
			}

			if part.Type == "" {
				return fmt.Errorf("content part %d in message %d must have a type", j, i)
			}

			// Validate specific content types
			switch part.Type {
			case interfaces.PonchoContentTypeText:
				if part.Text == "" {
					return fmt.Errorf("text content part %d in message %d cannot be empty", j, i)
				}
			case interfaces.PonchoContentTypeMedia:
				if part.Media == nil || part.Media.URL == "" {
					return fmt.Errorf("media content part %d in message %d must have a valid URL", j, i)
				}
				if !m.SupportsVision() {
					return fmt.Errorf("model '%s' does not support media content", m.name)
				}
			case interfaces.PonchoContentTypeTool:
				if part.Tool == nil || part.Tool.Name == "" {
					return fmt.Errorf("tool content part %d in message %d must have a valid tool definition", j, i)
				}
				if !m.SupportsTools() {
					return fmt.Errorf("model '%s' does not support tool calling", m.name)
				}
			}
		}
	}

	// Validate max_tokens if provided
	if req.MaxTokens != nil {
		if *req.MaxTokens <= 0 {
			return fmt.Errorf("max_tokens must be positive")
		}
		if *req.MaxTokens > m.MaxTokens() {
			return fmt.Errorf("max_tokens (%d) exceeds model maximum (%d)", *req.MaxTokens, m.MaxTokens())
		}
	}

	// Validate temperature if provided
	if req.Temperature != nil {
		if *req.Temperature < 0 || *req.Temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	// Validate tools if provided
	if req.Tools != nil && len(req.Tools) > 0 {
		if !m.SupportsTools() {
			return fmt.Errorf("model '%s' does not support tools", m.name)
		}

		for i, tool := range req.Tools {
			if tool == nil {
				return fmt.Errorf("tool %d cannot be nil", i)
			}
			if tool.Name == "" {
				return fmt.Errorf("tool %d must have a name", i)
			}
		}
	}

	return nil
}

// PrepareResponse prepares a basic response structure
func (m *PonchoBaseModel) PrepareResponse(message *interfaces.PonchoMessage, usage *interfaces.PonchoUsage, finishReason interfaces.PonchoFinishReason) *interfaces.PonchoModelResponse {
	return &interfaces.PonchoModelResponse{
		Message:      message,
		Usage:        usage,
		FinishReason: finishReason,
		Metadata:     make(map[string]interface{}),
	}
}

// PrepareUsage creates a usage structure
func (m *PonchoBaseModel) PrepareUsage(promptTokens, completionTokens int) *interfaces.PonchoUsage {
	return &interfaces.PonchoUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
}

// PrepareStreamChunk creates a streaming chunk
func (m *PonchoBaseModel) PrepareStreamChunk(delta *interfaces.PonchoMessage, usage *interfaces.PonchoUsage, finishReason interfaces.PonchoFinishReason, done bool) *interfaces.PonchoStreamChunk {
	return &interfaces.PonchoStreamChunk{
		Delta:        delta,
		Usage:        usage,
		FinishReason: finishReason,
		Done:         done,
		Metadata:     make(map[string]interface{}),
	}
}
