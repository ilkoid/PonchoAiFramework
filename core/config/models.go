package config

// This file defines the configuration data structures for the PonchoFramework.
// It provides comprehensive type definitions for all configuration components
// including model configuration structures (DeepSeek, Z.AI, etc.),
// tool configuration structures (S3, Wildberries, Vision, etc.),
// and flow configuration structures with dependency management.
// The file includes system-wide configuration settings (logging, metrics, etc.),
// supports environment variable substitution and validation tags,
// and serves as the central type system for framework configuration.
// It enables type-safe configuration management throughout the framework.

import (
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ModelFactory interface for creating model instances
type ModelFactory interface {
	// CreateModel creates a model instance from configuration
	CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error)
	
	// ValidateConfig validates model configuration
	ValidateConfig(config *interfaces.ModelConfig) error
	
	// GetProvider returns the provider this factory supports
	GetProvider() string
}

// ModelRegistry manages model factories and instances
type ModelRegistry struct {
	factories map[string]ModelFactory
	logger    interfaces.Logger
}

// NewModelRegistry creates a new model registry
func NewModelRegistry(logger interfaces.Logger) *ModelRegistry {
	return &ModelRegistry{
		factories: make(map[string]ModelFactory),
		logger:    logger,
	}
}

// RegisterFactory registers a model factory for a provider
func (mr *ModelRegistry) RegisterFactory(provider string, factory ModelFactory) error {
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}
	
	mr.logger.Debug("Registering model factory", "provider", provider)
	
	mr.factories[provider] = factory
	
	mr.logger.Info("Model factory registered", "provider", provider)
	return nil
}

// CreateModel creates a model instance from configuration
func (mr *ModelRegistry) CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error) {
	if config == nil {
		return nil, fmt.Errorf("model config cannot be nil")
	}
	
	if config.Provider == "" {
		return nil, fmt.Errorf("model provider is required")
	}
	
	factory, exists := mr.factories[config.Provider]
	if !exists {
		return nil, fmt.Errorf("no factory registered for provider: %s", config.Provider)
	}
	
	mr.logger.Debug("Creating model instance", 
		"provider", config.Provider,
		"model_name", config.ModelName)
	
	// Validate configuration first
	if err := factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid model config for provider %s: %w", config.Provider, err)
	}
	
	// Create model instance
	model, err := factory.CreateModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create model for provider %s: %w", config.Provider, err)
	}
	
	mr.logger.Info("Model instance created",
		"provider", config.Provider,
		"model_name", config.ModelName,
		"model", model.Name())
	
	return model, nil
}

// GetSupportedProviders returns list of supported providers
func (mr *ModelRegistry) GetSupportedProviders() []string {
	providers := make([]string, 0, len(mr.factories))
	for provider := range mr.factories {
		providers = append(providers, provider)
	}
	return providers
}

// ModelConfigValidator validates model configurations
type ModelConfigValidator struct {
	rules []ModelValidationRule
	logger interfaces.Logger
}

// NewModelConfigValidator creates a new model config validator
func NewModelConfigValidator(logger interfaces.Logger) *ModelConfigValidator {
	validator := &ModelConfigValidator{
		rules: make([]ModelValidationRule, 0),
		logger: logger,
	}
	
	// Add default validation rules
	validator.addDefaultRules()
	
	return validator
}

// ValidateConfig validates a model configuration
func (mcv *ModelConfigValidator) ValidateConfig(config *interfaces.ModelConfig) error {
	if config == nil {
		return fmt.Errorf("model config cannot be nil")
	}
	
	var errors []error
	
	// Apply all validation rules
	for _, rule := range mcv.rules {
		if err := mcv.applyRule(config, rule); err != nil {
			errors = append(errors, err)
		}
	}
	
	// Provider-specific validation
	if err := mcv.validateProviderSpecific(config); err != nil {
		errors = append(errors, err)
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("model config validation failed: %v", errors)
	}
	
	return nil
}

// AddRule adds a validation rule
func (mcv *ModelConfigValidator) AddRule(rule ModelValidationRule) {
	mcv.rules = append(mcv.rules, rule)
}

// applyRule applies a single validation rule
func (mcv *ModelConfigValidator) applyRule(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	switch rule.Field {
	case "provider":
		return mcv.validateProvider(config, rule)
	case "model_name":
		return mcv.validateModelName(config, rule)
	case "api_key":
		return mcv.validateAPIKey(config, rule)
	case "max_tokens":
		return mcv.validateMaxTokens(config, rule)
	case "temperature":
		return mcv.validateTemperature(config, rule)
	case "timeout":
		return mcv.validateTimeout(config, rule)
	case "supports":
		return mcv.validateSupports(config, rule)
	default:
		// Unknown field, skip
		return nil
	}
}

// validateProvider validates the provider field
func (mcv *ModelConfigValidator) validateProvider(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.Provider == "" {
		if rule.Required {
			return fmt.Errorf("provider is required")
		}
		return nil
	}
	
	// Check if provider is supported
	supportedProviders := []string{"deepseek", "zai", "openai"}
	for _, provider := range supportedProviders {
		if config.Provider == provider {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported provider: %s", config.Provider)
}

// validateModelName validates the model_name field
func (mcv *ModelConfigValidator) validateModelName(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.ModelName == "" {
		if rule.Required {
			return fmt.Errorf("model_name is required")
		}
		return nil
	}
	
	if len(config.ModelName) < 1 {
		return fmt.Errorf("model_name must be at least 1 character")
	}
	
	if len(config.ModelName) > 100 {
		return fmt.Errorf("model_name must be at most 100 characters")
	}
	
	return nil
}

// validateAPIKey validates the api_key field
func (mcv *ModelConfigValidator) validateAPIKey(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.APIKey == "" {
		if rule.Required {
			return fmt.Errorf("api_key is required")
		}
		return nil
	}
	
	// Check if it's an environment variable reference
	if len(config.APIKey) > 2 && config.APIKey[:2] == "${" && config.APIKey[len(config.APIKey)-1] == '}' {
		// Environment variable reference, validate format
		envVar := config.APIKey[2 : len(config.APIKey)-1]
		if envVar == "" {
			return fmt.Errorf("invalid environment variable reference in api_key")
		}
		return nil
	}
	
	// Validate API key format (basic check)
	if len(config.APIKey) < 10 {
		return fmt.Errorf("api_key appears to be too short")
	}
	
	return nil
}

// validateMaxTokens validates the max_tokens field
func (mcv *ModelConfigValidator) validateMaxTokens(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.MaxTokens <= 0 {
		if rule.Required {
			return fmt.Errorf("max_tokens must be positive")
		}
		return nil
	}
	
	if config.MaxTokens < 1 {
		return fmt.Errorf("max_tokens must be at least 1")
	}
	
	if config.MaxTokens > 100000 {
		return fmt.Errorf("max_tokens must be at most 100000")
	}
	
	return nil
}

// validateTemperature validates the temperature field
func (mcv *ModelConfigValidator) validateTemperature(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.Temperature < 0.0 || config.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}
	
	return nil
}

// validateTimeout validates the timeout field
func (mcv *ModelConfigValidator) validateTimeout(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.Timeout == "" {
		if rule.Required {
			return fmt.Errorf("timeout is required")
		}
		return nil
	}
	
	// Try to parse as duration
	if _, err := time.ParseDuration(config.Timeout); err != nil {
		return fmt.Errorf("invalid timeout format: %s", config.Timeout)
	}
	
	return nil
}

// validateSupports validates the supports field
func (mcv *ModelConfigValidator) validateSupports(config *interfaces.ModelConfig, rule ModelValidationRule) error {
	if config.Supports == nil {
		if rule.Required {
			return fmt.Errorf("supports is required")
		}
		return nil
	}
	
	// Validate capability values
	if config.Supports.Streaming && config.Provider == "deepseek" {
		// DeepSeek supports streaming, so this is valid
	}
	
	if config.Supports.Vision && config.Provider == "deepseek" {
		return fmt.Errorf("deepseek provider does not support vision")
	}
	
	return nil
}

// validateProviderSpecific performs provider-specific validation
func (mcv *ModelConfigValidator) validateProviderSpecific(config *interfaces.ModelConfig) error {
	switch config.Provider {
	case "deepseek":
		return mcv.validateDeepSeekConfig(config)
	case "zai":
		return mcv.validateZAIConfig(config)
	case "openai":
		return mcv.validateOpenAIConfig(config)
	default:
		return nil // No specific validation for unknown providers
	}
}

// validateDeepSeekConfig validates DeepSeek-specific configuration
func (mcv *ModelConfigValidator) validateDeepSeekConfig(config *interfaces.ModelConfig) error {
	// DeepSeek-specific validations
	if config.BaseURL != "" {
		// Validate URL format if provided
		if !mcv.isValidURL(config.BaseURL) {
			return fmt.Errorf("invalid base_url format")
		}
	}
	
	return nil
}

// validateZAIConfig validates Z.AI-specific configuration
func (mcv *ModelConfigValidator) validateZAIConfig(config *interfaces.ModelConfig) error {
	// Z.AI-specific validations
	if config.BaseURL != "" {
		// Validate URL format if provided
		if !mcv.isValidURL(config.BaseURL) {
			return fmt.Errorf("invalid base_url format")
		}
	}
	
	// Z.AI models with vision must have specific model names
	if config.Supports != nil && config.Supports.Vision {
		visionModels := []string{"glm-4.6v", "glm-4.5v"}
		found := false
		for _, model := range visionModels {
			if config.ModelName == model {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("model %s is not a supported vision model for Z.AI provider", config.ModelName)
		}
	}
	
	return nil
}

// validateOpenAIConfig validates OpenAI-specific configuration
func (mcv *ModelConfigValidator) validateOpenAIConfig(config *interfaces.ModelConfig) error {
	// OpenAI-specific validations
	if config.BaseURL != "" {
		// Validate URL format if provided
		if !mcv.isValidURL(config.BaseURL) {
			return fmt.Errorf("invalid base_url format")
		}
	}
	
	return nil
}

// isValidURL performs basic URL validation
func (mcv *ModelConfigValidator) isValidURL(url string) bool {
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}

// addDefaultRules adds default validation rules
func (mcv *ModelConfigValidator) addDefaultRules() {
	mcv.rules = append(mcv.rules, []ModelValidationRule{
		{
			Field:   "provider",
			Required: true,
			Type:     "string",
		},
		{
			Field:   "model_name",
			Required: true,
			Type:     "string",
		},
		{
			Field:   "api_key",
			Required: true,
			Type:     "string",
		},
		{
			Field:   "max_tokens",
			Required: true,
			Type:     "number",
		},
		{
			Field:   "temperature",
			Required: true,
			Type:     "number",
		},
		{
			Field:   "timeout",
			Required: true,
			Type:     "string",
		},
		{
			Field:   "supports",
			Required: true,
			Type:     "object",
		},
	}...)
}

// ModelValidationRule represents a validation rule for model configuration
type ModelValidationRule struct {
	Field   string `json:"field"`
	Required bool   `json:"required"`
	Type    string `json:"type"`
}

// ModelInitializer handles model initialization and lifecycle
type ModelInitializer struct {
	registry *ModelRegistry
	logger   interfaces.Logger
}

// NewModelInitializer creates a new model initializer
func NewModelInitializer(registry *ModelRegistry, logger interfaces.Logger) *ModelInitializer {
	return &ModelInitializer{
		registry: registry,
		logger:   logger,
	}
}

// InitializeModels initializes all models from configuration
func (mi *ModelInitializer) InitializeModels(configs map[string]*interfaces.ModelConfig) (map[string]interfaces.PonchoModel, error) {
	if configs == nil {
		return nil, fmt.Errorf("model configs cannot be nil")
	}
	
	models := make(map[string]interfaces.PonchoModel)
	var errors []error
	
	for name, config := range configs {
		mi.logger.Debug("Initializing model", "name", name)
		
		model, err := mi.registry.CreateModel(config)
		if err != nil {
			mi.logger.Error("Failed to create model", 
				"name", name, 
				"provider", config.Provider, 
				"error", err)
			errors = append(errors, fmt.Errorf("failed to initialize model %s: %w", name, err))
			continue
		}
		
		models[name] = model
		mi.logger.Info("Model initialized successfully", "name", name)
	}
	
	if len(errors) > 0 {
		return models, fmt.Errorf("some models failed to initialize: %v", errors)
	}
	
	mi.logger.Info("All models initialized successfully", "count", len(models))
	return models, nil
}

// ModelHealthChecker checks health of model instances
type ModelHealthChecker struct {
	models map[string]interfaces.PonchoModel
	logger interfaces.Logger
}

// NewModelHealthChecker creates a new model health checker
func NewModelHealthChecker(models map[string]interfaces.PonchoModel, logger interfaces.Logger) *ModelHealthChecker {
	return &ModelHealthChecker{
		models: models,
		logger: logger,
	}
}

// CheckHealth checks health of all models
func (mhc *ModelHealthChecker) CheckHealth() map[string]*ModelHealthStatus {
	status := make(map[string]*ModelHealthStatus)
	
	for name, model := range mhc.models {
		healthStatus := &ModelHealthStatus{
			Name:    name,
			Provider: model.Provider(),
			Status:   "healthy",
			Message:  "Model is operational",
			Timestamp: time.Now(),
		}
		
		// Perform basic health check
		if err := mhc.checkModelHealth(model); err != nil {
			healthStatus.Status = "unhealthy"
			healthStatus.Message = err.Error()
			mhc.logger.Warn("Model health check failed", 
				"name", name, 
				"error", err)
		}
		
		status[name] = healthStatus
	}
	
	return status
}

// checkModelHealth performs basic health check on a model
func (mhc *ModelHealthChecker) checkModelHealth(model interfaces.PonchoModel) error {
	// Basic health check - verify model is properly configured
	if model.Name() == "" {
		return fmt.Errorf("model name is empty")
	}
	
	if model.Provider() == "" {
		return fmt.Errorf("model provider is empty")
	}
	
	// Could add more sophisticated health checks here:
	// - Test API connectivity
	// - Validate API key
	// - Check model availability
	
	return nil
}

// ModelHealthStatus represents health status of a model
type ModelHealthStatus struct {
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}