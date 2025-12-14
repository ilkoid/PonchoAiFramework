package config

// This file implements the tool factory for the PonchoFramework.
// It provides factory methods for creating tool instances from configuration.
// It supports multiple tool types (S3, Wildberries, Vision, etc.).
// It handles tool initialization with proper configuration and dependencies.
// It provides tool capability detection and validation.
// It serves as the central mechanism for tool instantiation.
// It includes error handling for unsupported tool types.
// It enables dynamic tool loading from configuration files.
// It abstracts tool creation complexity from the main framework.

import (
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ToolFactory interface for creating tool instances
type ToolFactory interface {
	// CreateTool creates a tool instance from configuration
	CreateTool(config *interfaces.ToolConfig) (interfaces.PonchoTool, error)

	// ValidateConfig validates tool configuration
	ValidateConfig(config *interfaces.ToolConfig) error

	// GetToolType returns the tool type this factory supports
	GetToolType() string
}

// ToolRegistry manages tool factories and instances
type ToolRegistry struct {
	factories map[string]ToolFactory
	logger    interfaces.Logger
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(logger interfaces.Logger) *ToolRegistry {
	return &ToolRegistry{
		factories: make(map[string]ToolFactory),
		logger:    logger,
	}
}

// RegisterFactory registers a tool factory for a tool type
func (tr *ToolRegistry) RegisterFactory(toolType string, factory ToolFactory) error {
	if toolType == "" {
		return fmt.Errorf("tool type cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	tr.logger.Debug("Registering tool factory", "tool_type", toolType)

	tr.factories[toolType] = factory

	tr.logger.Info("Tool factory registered", "tool_type", toolType)
	return nil
}

// CreateTool creates a tool instance from configuration
func (tr *ToolRegistry) CreateTool(toolName string, config *interfaces.ToolConfig) (interfaces.PonchoTool, error) {
	if config == nil {
		return nil, fmt.Errorf("tool config cannot be nil")
	}

	// Determine tool type from tool name or config
	toolType := tr.determineToolType(toolName, config)

	factory, exists := tr.factories[toolType]
	if !exists {
		return nil, fmt.Errorf("no factory registered for tool type: %s", toolType)
	}

	tr.logger.Debug("Creating tool instance",
		"tool_name", toolName,
		"tool_type", toolType)

	// Validate configuration first
	if err := factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid tool config for type %s: %w", toolType, err)
	}

	// Create tool instance
	tool, err := factory.CreateTool(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool for type %s: %w", toolType, err)
	}

	tr.logger.Info("Tool instance created",
		"tool_name", toolName,
		"tool_type", toolType,
		"tool", tool.Name())

	return tool, nil
}

// GetSupportedToolTypes returns list of supported tool types
func (tr *ToolRegistry) GetSupportedToolTypes() []string {
	types := make([]string, 0, len(tr.factories))
	for toolType := range tr.factories {
		types = append(types, toolType)
	}
	return types
}

// determineToolType determines tool type from name and configuration
func (tr *ToolRegistry) determineToolType(toolName string, config *interfaces.ToolConfig) string {
	// Try to determine from tool name first
	switch {
	case toolName == "article_importer":
		return "article_importer"
	case toolName == "s3_storage":
		return "s3_storage"
	case toolName == "wb_categories":
		return "wildberries_api"
	case toolName == "vision_analyzer":
		return "vision_analysis"
	}

	// Try to determine from custom parameters
	if config.CustomParams != nil {
		if toolType, ok := config.CustomParams["tool_type"].(string); ok {
			return toolType
		}
	}

	// Default fallback
	return "generic"
}

// ToolConfigValidator validates tool configurations
type ToolConfigValidator struct {
	rules  []ToolValidationRule
	logger interfaces.Logger
}

// NewToolConfigValidator creates a new tool config validator
func NewToolConfigValidator(logger interfaces.Logger) *ToolConfigValidator {
	validator := &ToolConfigValidator{
		rules:  make([]ToolValidationRule, 0),
		logger: logger,
	}

	// Add default validation rules
	validator.addDefaultRules()

	return validator
}

// ValidateConfig validates a tool configuration
func (tcv *ToolConfigValidator) ValidateConfig(config *interfaces.ToolConfig) error {
	if config == nil {
		return fmt.Errorf("tool config cannot be nil")
	}

	var errors []error

	// Apply all validation rules
	for _, rule := range tcv.rules {
		if err := tcv.applyRule(config, rule); err != nil {
			errors = append(errors, err)
		}
	}

	// Tool-specific validation
	if err := tcv.validateToolSpecific(config); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("tool config validation failed: %v", errors)
	}

	return nil
}

// AddRule adds a validation rule
func (tcv *ToolConfigValidator) AddRule(rule ToolValidationRule) {
	tcv.rules = append(tcv.rules, rule)
}

// applyRule applies a single validation rule
func (tcv *ToolConfigValidator) applyRule(config *interfaces.ToolConfig, rule ToolValidationRule) error {
	switch rule.Field {
	case "enabled":
		return tcv.validateEnabled(config, rule)
	case "timeout":
		return tcv.validateTimeout(config, rule)
	case "retry":
		return tcv.validateRetry(config, rule)
	case "cache":
		return tcv.validateCache(config, rule)
	default:
		// Unknown field, skip
		return nil
	}
}

// validateEnabled validates the enabled field
func (tcv *ToolConfigValidator) validateEnabled(config *interfaces.ToolConfig, rule ToolValidationRule) error {
	// enabled is optional, so no validation needed if not required
	return nil
}

// validateTimeout validates the timeout field
func (tcv *ToolConfigValidator) validateTimeout(config *interfaces.ToolConfig, rule ToolValidationRule) error {
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

// validateRetry validates the retry field
func (tcv *ToolConfigValidator) validateRetry(config *interfaces.ToolConfig, rule ToolValidationRule) error {
	if config.Retry == nil {
		if rule.Required {
			return fmt.Errorf("retry configuration is required")
		}
		return nil
	}

	// Validate retry configuration
	if config.Retry.MaxAttempts < 1 {
		return fmt.Errorf("retry max_attempts must be at least 1")
	}

	if config.Retry.MaxAttempts > 10 {
		return fmt.Errorf("retry max_attempts cannot exceed 10")
	}

	if config.Retry.Backoff != "linear" && config.Retry.Backoff != "exponential" {
		return fmt.Errorf("retry backoff must be 'linear' or 'exponential'")
	}

	if config.Retry.BaseDelay == "" {
		return fmt.Errorf("retry base_delay is required")
	}

	if _, err := time.ParseDuration(config.Retry.BaseDelay); err != nil {
		return fmt.Errorf("invalid retry base_delay format: %s", config.Retry.BaseDelay)
	}

	return nil
}

// validateCache validates the cache field
func (tcv *ToolConfigValidator) validateCache(config *interfaces.ToolConfig, rule ToolValidationRule) error {
	if config.Cache == nil {
		if rule.Required {
			return fmt.Errorf("cache configuration is required")
		}
		return nil
	}

	// Validate cache configuration
	if config.Cache.TTL == "" {
		return fmt.Errorf("cache ttl is required")
	}

	if _, err := time.ParseDuration(config.Cache.TTL); err != nil {
		return fmt.Errorf("invalid cache ttl format: %s", config.Cache.TTL)
	}

	if config.Cache.MaxSize < 1 {
		return fmt.Errorf("cache max_size must be at least 1")
	}

	return nil
}

// validateToolSpecific performs tool-specific validation
func (tcv *ToolConfigValidator) validateToolSpecific(config *interfaces.ToolConfig) error {
	// This would be extended with tool-specific validation logic
	// For now, just basic validation
	return nil
}

// addDefaultRules adds default validation rules
func (tcv *ToolConfigValidator) addDefaultRules() {
	tcv.rules = append(tcv.rules, []ToolValidationRule{
		{
			Field:    "enabled",
			Required: false,
			Type:     "boolean",
		},
		{
			Field:    "timeout",
			Required: false,
			Type:     "string",
		},
		{
			Field:    "retry",
			Required: false,
			Type:     "object",
		},
		{
			Field:    "cache",
			Required: false,
			Type:     "object",
		},
	}...)
}

// ToolValidationRule represents a validation rule for tool configuration
type ToolValidationRule struct {
	Field    string `json:"field"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

// ToolInitializer handles tool initialization and lifecycle
type ToolInitializer struct {
	registry *ToolRegistry
	logger   interfaces.Logger
}

// NewToolInitializer creates a new tool initializer
func NewToolInitializer(registry *ToolRegistry, logger interfaces.Logger) *ToolInitializer {
	return &ToolInitializer{
		registry: registry,
		logger:   logger,
	}
}

// InitializeTools initializes all tools from configuration
func (ti *ToolInitializer) InitializeTools(configs map[string]*interfaces.ToolConfig, globalConfig *interfaces.PonchoFrameworkConfig) (map[string]interfaces.PonchoTool, error) {
	if configs == nil {
		return nil, fmt.Errorf("tool configs cannot be nil")
	}

	tools := make(map[string]interfaces.PonchoTool)
	var errors []error

	for name, config := range configs {
		ti.logger.Debug("Initializing tool", "name", name)

		// Merge global configuration into tool config
		mergedConfig := ti.mergeGlobalConfig(config, globalConfig)

		tool, err := ti.registry.CreateTool(name, mergedConfig)
		if err != nil {
			ti.logger.Error("Failed to create tool",
				"name", name,
				"error", err)
			errors = append(errors, fmt.Errorf("failed to initialize tool %s: %w", name, err))
			continue
		}

		tools[name] = tool
		ti.logger.Info("Tool initialized successfully", "name", name)
	}

	if len(errors) > 0 {
		return tools, fmt.Errorf("some tools failed to initialize: %v", errors)
	}

	ti.logger.Info("All tools initialized successfully", "count", len(tools))
	return tools, nil
}

// mergeGlobalConfig merges global configuration into tool config
func (ti *ToolInitializer) mergeGlobalConfig(toolConfig *interfaces.ToolConfig, globalConfig *interfaces.PonchoFrameworkConfig) *interfaces.ToolConfig {
	merged := *toolConfig // Copy the tool config

	// Merge S3 configuration if available
	if globalConfig.S3 != nil {
		if merged.CustomParams == nil {
			merged.CustomParams = make(map[string]interface{})
		}

		// Add S3 config to custom params for tools that need it
		merged.CustomParams["s3"] = map[string]interface{}{
			"url":      globalConfig.S3.URL,
			"region":   globalConfig.S3.Region,
			"bucket":   globalConfig.S3.Bucket,
			"endpoint": globalConfig.S3.Endpoint,
			"use_ssl":  globalConfig.S3.UseSSL,
		}
	}

	// Merge security configuration if available
	if globalConfig.Security != nil {
		if merged.CustomParams == nil {
			merged.CustomParams = make(map[string]interface{})
		}

		// Add API keys to custom params
		merged.CustomParams["api_keys"] = globalConfig.Security.APIKeys
	}

	return &merged
}

// ToolHealthChecker checks health of tool instances
type ToolHealthChecker struct {
	tools  map[string]interfaces.PonchoTool
	logger interfaces.Logger
}

// NewToolHealthChecker creates a new tool health checker
func NewToolHealthChecker(tools map[string]interfaces.PonchoTool, logger interfaces.Logger) *ToolHealthChecker {
	return &ToolHealthChecker{
		tools:  tools,
		logger: logger,
	}
}

// CheckHealth checks health of all tools
func (thc *ToolHealthChecker) CheckHealth() map[string]*ToolHealthStatus {
	status := make(map[string]*ToolHealthStatus)

	for name, tool := range thc.tools {
		healthStatus := &ToolHealthStatus{
			Name:      name,
			Category:  tool.Category(),
			Status:    "healthy",
			Message:   "Tool is operational",
			Timestamp: time.Now(),
		}

		// Perform basic health check
		if err := thc.checkToolHealth(tool); err != nil {
			healthStatus.Status = "unhealthy"
			healthStatus.Message = err.Error()
			thc.logger.Warn("Tool health check failed",
				"name", name,
				"error", err)
		}

		status[name] = healthStatus
	}

	return status
}

// checkToolHealth performs basic health check on a tool
func (thc *ToolHealthChecker) checkToolHealth(tool interfaces.PonchoTool) error {
	// Basic health check - verify tool is properly configured
	if tool.Name() == "" {
		return fmt.Errorf("tool name is empty")
	}

	if tool.Category() == "" {
		return fmt.Errorf("tool category is empty")
	}

	// Could add more sophisticated health checks here:
	// - Test external service connectivity
	// - Validate API keys
	// - Check tool-specific dependencies

	return nil
}

// ToolHealthStatus represents health status of a tool
type ToolHealthStatus struct {
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
