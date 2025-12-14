package models

// ModelFactory implements the model factory for the PonchoFramework.
// It provides factory methods for creating AI model instances from configuration.
// It supports multiple model providers (DeepSeek, Z.AI, etc.).
// It handles model initialization with proper configuration and credentials.
// It provides model capability detection and validation.
// It serves as the central mechanism for model instantiation.
// It includes error handling for unsupported model types.
// It enables dynamic model loading from configuration files.
// It abstracts model creation complexity from the main framework.

import (
	"context"
	"fmt"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/deepseek"
	"github.com/ilkoid/PonchoAiFramework/models/zai"
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

	// Import and create actual DeepSeek model
	model := deepseek.NewDeepSeekModel()
	
	// Convert ModelConfig to map[string]interface{} for Initialize
	configMap := map[string]interface{}{
		"api_key":     config.APIKey,
		"model_name":   config.ModelName,
		"max_tokens":   config.MaxTokens,
		"temperature":  config.Temperature,
		"timeout":      config.Timeout,
		"base_url":     config.BaseURL,
		"supports":     config.Supports,
	}
	
	// Add custom parameters if any
	if config.CustomParams != nil {
		for k, v := range config.CustomParams {
			configMap[k] = v
		}
	}
	
	// Initialize model with the provided config
	if err := model.Initialize(context.Background(), configMap); err != nil {
		return nil, fmt.Errorf("failed to initialize DeepSeek model: %w", err)
	}
	
	return model, nil
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
		"provider":    config.Provider,
		"model_name":  config.ModelName,
		"api_key":     config.APIKey,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"timeout":     config.Timeout,
		"supports":    config.Supports,
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
		"top_p":             true,
		"frequency_penalty": true,
		"presence_penalty":  true,
		"stop":              true,
		"response_format":   true,
		"thinking":          true,
		"logprobs":          true,
		"top_logprobs":      true,
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

	// Create appropriate Z.AI model based on model name
	var model interfaces.PonchoModel
	if strings.Contains(strings.ToLower(config.ModelName), "vision") || strings.Contains(strings.ToLower(config.ModelName), "v") {
		model = zai.NewZAIVisionModel()
	} else {
		model = zai.NewZAIModel()
	}
	
	// Convert ModelConfig to map[string]interface{} for Initialize
	configMap := map[string]interface{}{
		"api_key":     config.APIKey,
		"model_name":   config.ModelName,
		"max_tokens":   config.MaxTokens,
		"temperature":  config.Temperature,
		"timeout":      config.Timeout,
		"base_url":     config.BaseURL,
		"supports":     config.Supports,
	}
	
	// Add custom parameters if any
	if config.CustomParams != nil {
		for k, v := range config.CustomParams {
			configMap[k] = v
		}
	}
	
	// Initialize model with the provided config
	if err := model.Initialize(context.Background(), configMap); err != nil {
		return nil, fmt.Errorf("failed to initialize Z.AI model: %w", err)
	}
	
	return model, nil
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
		"provider":    config.Provider,
		"model_name":  config.ModelName,
		"api_key":     config.APIKey,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"timeout":     config.Timeout,
		"supports":    config.Supports,
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
		"top_p":             true,
		"frequency_penalty": true,
		"presence_penalty":  true,
		"stop":              true,
		"thinking":          true,
		"logprobs":          true,
		"top_logprobs":      true,
		"max_image_size":    true,
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

	// TODO: Implement actual OpenAI model creation when available
	// For now, return a placeholder to indicate configuration is valid
	return &PlaceholderModel{
		name:     config.ModelName,
		provider: config.Provider,
		capabilities: &interfaces.ModelCapabilities{
			Streaming: config.Supports != nil && config.Supports.Streaming,
			Tools:     config.Supports != nil && config.Supports.Tools,
			Vision:    config.Supports != nil && config.Supports.Vision,
			System:    config.Supports != nil && config.Supports.System,
			JSONMode:  config.Supports != nil && config.Supports.JSONMode,
		},
	}, nil
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

	// Basic OpenAI validation
	if config.Supports != nil && config.Supports.Vision {
		visionModels := []string{"gpt-4-vision-preview", "gpt-4o", "gpt-4o-mini"}
		found := false
		for _, model := range visionModels {
			if config.ModelName == model {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("model %s does not support vision for OpenAI provider", config.ModelName)
		}
	}

	return nil
}

// ModelFactoryManager manages all model factories
type ModelFactoryManager struct {
	factories map[string]interfaces.ModelFactory
	logger    interfaces.Logger
}

// NewModelFactoryManager creates a new model factory manager
func NewModelFactoryManager(logger interfaces.Logger) *ModelFactoryManager {
	manager := &ModelFactoryManager{
		factories: make(map[string]interfaces.ModelFactory),
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
func (m *ModelFactoryManager) RegisterFactory(provider string, factory interfaces.ModelFactory) error {
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
func (m *ModelFactoryManager) GetFactory(provider string) (interfaces.ModelFactory, error) {
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

// PlaceholderModel is a temporary placeholder for model implementations
// This will be replaced with actual model implementations in Phase 2
type PlaceholderModel struct {
	name         string
	provider     string
	capabilities *interfaces.ModelCapabilities
}

// Name returns model name
func (p *PlaceholderModel) Name() string {
	return p.name
}

// Provider returns model provider
func (p *PlaceholderModel) Provider() string {
	return p.provider
}

// MaxTokens returns maximum tokens supported
func (p *PlaceholderModel) MaxTokens() int {
	return 4000 // Default placeholder value
}

// DefaultTemperature returns default temperature setting
func (p *PlaceholderModel) DefaultTemperature() float32 {
	return 0.7 // Default placeholder value
}

// Initialize initializes model (placeholder implementation)
func (p *PlaceholderModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Placeholder implementation - does nothing
	return nil
}

// Shutdown shuts down model (placeholder implementation)
func (p *PlaceholderModel) Shutdown(ctx context.Context) error {
	// Placeholder implementation - does nothing
	return nil
}

// Temperature returns temperature setting (deprecated, use DefaultTemperature)
func (p *PlaceholderModel) Temperature() float32 {
	return p.DefaultTemperature()
}

// SupportsStreaming returns whether model supports streaming
func (p *PlaceholderModel) SupportsStreaming() bool {
	return p.capabilities != nil && p.capabilities.Streaming
}

// SupportsTools returns whether model supports tools
func (p *PlaceholderModel) SupportsTools() bool {
	return p.capabilities != nil && p.capabilities.Tools
}

// SupportsVision returns whether model supports vision
func (p *PlaceholderModel) SupportsVision() bool {
	return p.capabilities != nil && p.capabilities.Vision
}

// SupportsSystemRole returns whether model supports system role
func (p *PlaceholderModel) SupportsSystemRole() bool {
	return p.capabilities != nil && p.capabilities.System
}

// SupportsJSONMode returns whether model supports JSON mode
func (p *PlaceholderModel) SupportsJSONMode() bool {
	return p.capabilities != nil && p.capabilities.JSONMode
}

// Generate generates a response (placeholder implementation)
func (p *PlaceholderModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	return nil, fmt.Errorf("placeholder model %s does not support generation - implementation pending in Phase 2", p.name)
}

// GenerateStreaming generates a streaming response (placeholder implementation)
func (p *PlaceholderModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	return fmt.Errorf("placeholder model %s does not support streaming - implementation pending in Phase 2", p.name)
}