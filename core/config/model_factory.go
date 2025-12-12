package config

import (
	"fmt"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// DeepSeekModelFactory creates DeepSeek model instances
type DeepSeekModelFactory struct{}

// NewDeepSeekModelFactory creates a new DeepSeek model factory
func NewDeepSeekModelFactory() *DeepSeekModelFactory {
	return &DeepSeekModelFactory{}
}

// CreateModel creates a DeepSeek model instance from configuration
func (f *DeepSeekModelFactory) CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error) {
	if config.Provider != "deepseek" {
		return nil, fmt.Errorf("invalid provider for DeepSeek factory: %s", config.Provider)
	}

	// TODO: This will be fixed when we move factory registration to avoid circular dependency
	// For now, return an error to break the cycle
	return nil, fmt.Errorf("DeepSeek factory temporarily disabled due to circular dependency - this will be fixed in Phase 2")
}

// ValidateConfig validates DeepSeek-specific configuration
func (f *DeepSeekModelFactory) ValidateConfig(config *interfaces.ModelConfig) error {
	if config.Provider != "deepseek" {
		return fmt.Errorf("invalid provider for DeepSeek factory: %s", config.Provider)
	}

	// DeepSeek-specific validations
	if config.Supports != nil && config.Supports.Vision {
		return fmt.Errorf("DeepSeek models do not support vision")
	}

	// Validate custom parameters for DeepSeek
	if config.CustomParams != nil {
		if err := f.validateDeepSeekCustomParams(config.CustomParams); err != nil {
			return fmt.Errorf("invalid custom parameters: %w", err)
		}
	}

	return nil
}

// GetProvider returns the provider this factory supports
func (f *DeepSeekModelFactory) GetProvider() string {
	return "deepseek"
}

// prepareInitConfig prepares initialization configuration for DeepSeek
func (f *DeepSeekModelFactory) prepareInitConfig(config *interfaces.ModelConfig) map[string]interface{} {
	initConfig := map[string]interface{}{
		"provider":      config.Provider,
		"model_name":    config.ModelName,
		"api_key":       config.APIKey,
		"max_tokens":    config.MaxTokens,
		"temperature":   config.Temperature,
		"timeout":       config.Timeout,
		"supports":      config.Supports,
	}

	// Add base URL if provided
	if config.BaseURL != "" {
		initConfig["base_url"] = config.BaseURL
	}

	// Add custom parameters
	if config.CustomParams != nil {
		initConfig["custom_params"] = config.CustomParams
	}

	return initConfig
}

// validateDeepSeekCustomParams validates DeepSeek-specific custom parameters
func (f *DeepSeekModelFactory) validateDeepSeekCustomParams(params map[string]interface{}) error {
	// Validate known DeepSeek parameters
	validParams := map[string]bool{
		"top_p":              true,
		"frequency_penalty":   true,
		"presence_penalty":    true,
		"stop":               true,
		"response_format":     true,
		"thinking":           true,
		"logprobs":           true,
		"top_logprobs":       true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("unknown DeepSeek parameter: %s", param)
		}
	}

	// Validate parameter types and ranges
	if topP, ok := params["top_p"]; ok {
		if err := f.validateTopP(topP); err != nil {
			return fmt.Errorf("invalid top_p: %w", err)
		}
	}

	if freqPenalty, ok := params["frequency_penalty"]; ok {
		if err := f.validatePenalty(freqPenalty, "frequency_penalty"); err != nil {
			return fmt.Errorf("invalid frequency_penalty: %w", err)
		}
	}

	if presPenalty, ok := params["presence_penalty"]; ok {
		if err := f.validatePenalty(presPenalty, "presence_penalty"); err != nil {
			return fmt.Errorf("invalid presence_penalty: %w", err)
		}
	}

	if logprobs, ok := params["logprobs"]; ok {
		if _, ok := logprobs.(bool); !ok {
			return fmt.Errorf("logprobs must be a boolean")
		}
	}

	if topLogprobs, ok := params["top_logprobs"]; ok {
		if err := f.validateTopLogprobs(topLogprobs); err != nil {
			return fmt.Errorf("invalid top_logprobs: %w", err)
		}
	}

	return nil
}

// validateTopP validates top_p parameter
func (f *DeepSeekModelFactory) validateTopP(value interface{}) error {
	var topP float32
	
	switch v := value.(type) {
	case float32:
		topP = v
	case float64:
		topP = float32(v)
	case int:
		topP = float32(v)
	default:
		return fmt.Errorf("top_p must be a number")
	}

	if topP < 0.0 || topP > 1.0 {
		return fmt.Errorf("top_p must be between 0.0 and 1.0")
	}

	return nil
}

// validatePenalty validates penalty parameters
func (f *DeepSeekModelFactory) validatePenalty(value interface{}, paramName string) error {
	var penalty float32

	switch v := value.(type) {
	case float32:
		penalty = v
	case float64:
		penalty = float32(v)
	case int:
		penalty = float32(v)
	default:
		return fmt.Errorf("%s must be a number", paramName)
	}

	if penalty < -2.0 || penalty > 2.0 {
		return fmt.Errorf("%s must be between -2.0 and 2.0", paramName)
	}

	return nil
}

// validateTopLogprobs validates top_logprobs parameter
func (f *DeepSeekModelFactory) validateTopLogprobs(value interface{}) error {
	switch v := value.(type) {
	case int:
		if v < 0 || v > 20 {
			return fmt.Errorf("top_logprobs must be between 0 and 20")
		}
	default:
		return fmt.Errorf("top_logprobs must be an integer")
	}

	return nil
}

// ZAIModelFactory creates Z.AI model instances
type ZAIModelFactory struct{}

// NewZAIModelFactory creates a new Z.AI model factory
func NewZAIModelFactory() *ZAIModelFactory {
	return &ZAIModelFactory{}
}

// CreateModel creates a Z.AI model instance from configuration
func (f *ZAIModelFactory) CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error) {
	if config.Provider != "zai" {
		return nil, fmt.Errorf("invalid provider for Z.AI factory: %s", config.Provider)
	}

	// TODO: This will be fixed when we move factory registration to avoid circular dependency
	// For now, return an error to break the cycle
	return nil, fmt.Errorf("Z.AI factory temporarily disabled due to circular dependency - this will be fixed in Phase 2")
}

// ValidateConfig validates Z.AI-specific configuration
func (f *ZAIModelFactory) ValidateConfig(config *interfaces.ModelConfig) error {
	if config.Provider != "zai" {
		return fmt.Errorf("invalid provider for Z.AI factory: %s", config.Provider)
	}

	// Z.AI-specific validations
	if config.Supports != nil && config.Supports.Vision {
		// Check if model name supports vision
		visionModels := []string{"glm-4.6v", "glm-4.5v", "glm-4v"}
		found := false
		for _, model := range visionModels {
			if config.ModelName == model {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("model %s does not support vision for Z.AI provider", config.ModelName)
		}
	}

	// Validate custom parameters for Z.AI
	if config.CustomParams != nil {
		if err := f.validateZAICustomParams(config.CustomParams); err != nil {
			return fmt.Errorf("invalid custom parameters: %w", err)
		}
	}

	return nil
}

// GetProvider returns the provider this factory supports
func (f *ZAIModelFactory) GetProvider() string {
	return "zai"
}

// prepareInitConfig prepares initialization configuration for Z.AI
func (f *ZAIModelFactory) prepareInitConfig(config *interfaces.ModelConfig) map[string]interface{} {
	initConfig := map[string]interface{}{
		"provider":      config.Provider,
		"model_name":    config.ModelName,
		"api_key":       config.APIKey,
		"max_tokens":    config.MaxTokens,
		"temperature":   config.Temperature,
		"timeout":       config.Timeout,
		"supports":      config.Supports,
	}

	// Add base URL if provided
	if config.BaseURL != "" {
		initConfig["base_url"] = config.BaseURL
	}

	// Add custom parameters
	if config.CustomParams != nil {
		initConfig["custom_params"] = config.CustomParams
	}

	return initConfig
}

// validateZAICustomParams validates Z.AI-specific custom parameters
func (f *ZAIModelFactory) validateZAICustomParams(params map[string]interface{}) error {
	// Validate known Z.AI parameters
	validParams := map[string]bool{
		"top_p":              true,
		"frequency_penalty":   true,
		"presence_penalty":    true,
		"stop":               true,
		"thinking":           true,
		"logprobs":           true,
		"top_logprobs":       true,
		"max_image_size":     true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("unknown Z.AI parameter: %s", param)
		}
	}

	// Validate parameter types and ranges
	if topP, ok := params["top_p"]; ok {
		if err := f.validateTopP(topP); err != nil {
			return fmt.Errorf("invalid top_p: %w", err)
		}
	}

	if freqPenalty, ok := params["frequency_penalty"]; ok {
		if err := f.validatePenalty(freqPenalty, "frequency_penalty"); err != nil {
			return fmt.Errorf("invalid frequency_penalty: %w", err)
		}
	}

	if presPenalty, ok := params["presence_penalty"]; ok {
		if err := f.validatePenalty(presPenalty, "presence_penalty"); err != nil {
			return fmt.Errorf("invalid presence_penalty: %w", err)
		}
	}

	if maxImageSize, ok := params["max_image_size"]; ok {
		if err := f.validateMaxImageSize(maxImageSize); err != nil {
			return fmt.Errorf("invalid max_image_size: %w", err)
		}
	}

	if thinking, ok := params["thinking"]; ok {
		if err := f.validateThinking(thinking); err != nil {
			return fmt.Errorf("invalid thinking: %w", err)
		}
	}

	return nil
}

// validateMaxImageSize validates max_image_size parameter
func (f *ZAIModelFactory) validateMaxImageSize(value interface{}) error {
	switch v := value.(type) {
	case string:
		// Could be "10MB", "5MB", etc.
		return nil // Basic validation - could be enhanced
	case int:
		if v <= 0 {
			return fmt.Errorf("max_image_size must be positive")
		}
		if v > 50*1024*1024 { // 50MB
			return fmt.Errorf("max_image_size must be at most 50MB")
		}
		return nil
	default:
		return fmt.Errorf("max_image_size must be a string or integer")
	}
}

// validateThinking validates thinking parameter
func (f *ZAIModelFactory) validateThinking(value interface{}) error {
	switch v := value.(type) {
	case map[string]interface{}:
		// Validate thinking configuration
		if thinkingType, ok := v["type"].(string); ok {
			validTypes := []string{"enabled", "disabled"}
			found := false
			for _, validType := range validTypes {
				if thinkingType == validType {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid thinking type: %s", thinkingType)
			}
		} else {
			return fmt.Errorf("thinking configuration must have a 'type' field")
		}
	case string:
		// Simple string format
		validTypes := []string{"enabled", "disabled"}
		found := false
		for _, validType := range validTypes {
			if v == validType {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid thinking type: %s", v)
		}
	default:
		return fmt.Errorf("thinking must be a map or string")
	}

	return nil
}

// validateTopP validates top_p parameter (shared between factories)
func (f *ZAIModelFactory) validateTopP(value interface{}) error {
	var topP float32

	switch v := value.(type) {
	case float32:
		topP = v
	case float64:
		topP = float32(v)
	case int:
		topP = float32(v)
	default:
		return fmt.Errorf("top_p must be a number")
	}

	if topP < 0.0 || topP > 1.0 {
		return fmt.Errorf("top_p must be between 0.0 and 1.0")
	}

	return nil
}

// validatePenalty validates penalty parameters (shared between factories)
func (f *ZAIModelFactory) validatePenalty(value interface{}, paramName string) error {
	var penalty float32

	switch v := value.(type) {
	case float32:
		penalty = v
	case float64:
		penalty = float32(v)
	case int:
		penalty = float32(v)
	default:
		return fmt.Errorf("%s must be a number", paramName)
	}

	if penalty < -2.0 || penalty > 2.0 {
		return fmt.Errorf("%s must be between -2.0 and 2.0", paramName)
	}

	return nil
}

// OpenAIModelFactory creates OpenAI model instances (placeholder for future)
type OpenAIModelFactory struct{}

// NewOpenAIModelFactory creates a new OpenAI model factory
func NewOpenAIModelFactory() *OpenAIModelFactory {
	return &OpenAIModelFactory{}
}

// CreateModel creates an OpenAI model instance from configuration
func (f *OpenAIModelFactory) CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error) {
	if config.Provider != "openai" {
		return nil, fmt.Errorf("invalid provider for OpenAI factory: %s", config.Provider)
	}

	// TODO: Implement OpenAI model when available
	return nil, fmt.Errorf("OpenAI model factory not yet implemented")
}

// GetProvider returns the provider this factory supports
func (f *OpenAIModelFactory) GetProvider() string {
	return "openai"
}

// ValidateConfig validates OpenAI-specific configuration
func (f *OpenAIModelFactory) ValidateConfig(config *interfaces.ModelConfig) error {
	if config.Provider != "openai" {
		return fmt.Errorf("invalid provider for OpenAI factory: %s", config.Provider)
	}

	// TODO: Implement OpenAI validation when available
	return fmt.Errorf("OpenAI model factory not yet implemented")
}

// ModelFactoryManager manages all model factories
type ModelFactoryManager struct {
	factories map[string]ModelFactory
	logger    interfaces.Logger
}

// NewModelFactoryManager creates a new model factory manager
func NewModelFactoryManager(logger interfaces.Logger) *ModelFactoryManager {
	manager := &ModelFactoryManager{
		factories: make(map[string]ModelFactory),
		logger:    logger,
	}

	// Register default factories
	manager.registerDefaultFactories()

	return manager
}

// registerDefaultFactories registers all default model factories
func (m *ModelFactoryManager) registerDefaultFactories() {
	m.RegisterFactory("deepseek", NewDeepSeekModelFactory())
	m.RegisterFactory("zai", NewZAIModelFactory())
	m.RegisterFactory("openai", NewOpenAIModelFactory()) // Placeholder for future

	m.logger.Info("Default model factories registered",
		"providers", []string{"deepseek", "zai", "openai"})
}

// RegisterFactory registers a model factory
func (m *ModelFactoryManager) RegisterFactory(provider string, factory ModelFactory) error {
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	m.logger.Debug("Registering model factory", "provider", provider)
	m.factories[provider] = factory

	m.logger.Info("Model factory registered", "provider", provider)
	return nil
}

// GetFactory returns a factory for the specified provider
func (m *ModelFactoryManager) GetFactory(provider string) (ModelFactory, error) {
	factory, exists := m.factories[provider]
	if !exists {
		return nil, fmt.Errorf("no factory registered for provider: %s", provider)
	}

	return factory, nil
}

// GetSupportedProviders returns list of supported providers
func (m *ModelFactoryManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(m.factories))
	for provider := range m.factories {
		providers = append(providers, provider)
	}

	return providers
}

// CreateModel creates a model instance using the appropriate factory
func (m *ModelFactoryManager) CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error) {
	if config == nil {
		return nil, fmt.Errorf("model config cannot be nil")
	}

	if config.Provider == "" {
		return nil, fmt.Errorf("model provider is required")
	}

	factory, err := m.GetFactory(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get factory for provider %s: %w", config.Provider, err)
	}

	m.logger.Debug("Creating model using factory",
		"provider", config.Provider,
		"model_name", config.ModelName)

	model, err := factory.CreateModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create model with factory %s: %w", config.Provider, err)
	}

	m.logger.Info("Model created successfully",
		"provider", config.Provider,
		"model_name", config.ModelName,
		"model", model.Name())

	return model, nil
}

// ValidateConfig validates model configuration using the appropriate factory
func (m *ModelFactoryManager) ValidateConfig(config *interfaces.ModelConfig) error {
	if config == nil {
		return fmt.Errorf("model config cannot be nil")
	}

	if config.Provider == "" {
		return fmt.Errorf("model provider is required")
	}

	factory, err := m.GetFactory(config.Provider)
	if err != nil {
		return fmt.Errorf("failed to get factory for provider %s: %w", config.Provider, err)
	}

	return factory.ValidateConfig(config)
}