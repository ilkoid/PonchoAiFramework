package base

// This file provides the base implementation for tools in the PonchoFramework
// It implements the PonchoBaseTool struct that serves as a foundation for all tool implementations
// It provides common functionality like initialization, configuration management, and metadata handling
// It implements the PonchoTool interface with default behaviors
// It handles tool execution with validation and error handling
// It provides input/output schema validation capabilities
// It serves as a reusable base for specific tool implementations (S3, Wildberries, Vision, etc.)
// It includes logging and metrics collection for tool operations

import (
	"context"
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoBaseTool provides a base implementation of the PonchoTool interface
type PonchoBaseTool struct {
	name         string
	description  string
	version      string
	category     string
	tags         []string
	dependencies []string
	inputSchema  map[string]interface{}
	outputSchema map[string]interface{}
	config       map[string]interface{}
	logger       interfaces.Logger
	mutex        sync.RWMutex
	initialized  bool
}

// NewPonchoBaseTool creates a new base tool instance
func NewPonchoBaseTool(name, description, version, category string) *PonchoBaseTool {
	return &PonchoBaseTool{
		name:         name,
		description:  description,
		version:      version,
		category:     category,
		tags:         []string{},
		dependencies: []string{},
		inputSchema:  make(map[string]interface{}),
		outputSchema: make(map[string]interface{}),
		config:       make(map[string]interface{}),
		logger:       interfaces.NewDefaultLogger(),
	}
}

// Name returns the tool name
func (t *PonchoBaseTool) Name() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.name
}

// Description returns the tool description
func (t *PonchoBaseTool) Description() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.description
}

// Version returns the tool version
func (t *PonchoBaseTool) Version() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.version
}

// Category returns the tool category
func (t *PonchoBaseTool) Category() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.category
}

// Tags returns the tool tags
func (t *PonchoBaseTool) Tags() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Return a copy to prevent external modification
	tags := make([]string, len(t.tags))
	copy(tags, t.tags)
	return tags
}

// Dependencies returns the tool dependencies
func (t *PonchoBaseTool) Dependencies() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Return a copy to prevent external modification
	deps := make([]string, len(t.dependencies))
	copy(deps, t.dependencies)
	return deps
}

// InputSchema returns the input schema
func (t *PonchoBaseTool) InputSchema() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Return a copy to prevent external modification
	schema := make(map[string]interface{})
	for k, v := range t.inputSchema {
		schema[k] = v
	}
	return schema
}

// OutputSchema returns the output schema
func (t *PonchoBaseTool) OutputSchema() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Return a copy to prevent external modification
	schema := make(map[string]interface{})
	for k, v := range t.outputSchema {
		schema[k] = v
	}
	return schema
}

// Initialize initializes the tool with configuration
func (t *PonchoBaseTool) Initialize(ctx context.Context, config map[string]interface{}) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.initialized {
		return fmt.Errorf("tool '%s' is already initialized", t.name)
	}

	// Validate configuration
	if err := t.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config for tool '%s': %w", t.name, err)
	}

	// Set configuration
	if config != nil {
		t.config = make(map[string]interface{})
		for k, v := range config {
			t.config[k] = v
		}
	}

	t.initialized = true
	t.logger.Info("Tool initialized", "tool", t.name, "version", t.version, "category", t.category)

	return nil
}

// Shutdown shuts down the tool and cleans up resources
func (t *PonchoBaseTool) Shutdown(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.initialized {
		return nil
	}

	// Clear configuration
	t.config = make(map[string]interface{})
	t.initialized = false

	t.logger.Info("Tool shutdown", "tool", t.name, "version", t.version)

	return nil
}

// Execute executes the tool (must be implemented by concrete tools)
func (t *PonchoBaseTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if !t.isInitialized() {
		return nil, fmt.Errorf("tool '%s' is not initialized", t.name)
	}

	// Validate input against schema
	if err := t.Validate(input); err != nil {
		return nil, fmt.Errorf("input validation failed for tool '%s': %w", t.name, err)
	}

	// This is a base implementation - concrete tools should override this
	return nil, fmt.Errorf("Execute method must be implemented by concrete tool")
}

// Validate validates the input against the input schema
func (t *PonchoBaseTool) Validate(input interface{}) error {
	if t.inputSchema == nil || len(t.inputSchema) == 0 {
		return nil // No schema - skip validation
	}

	// Basic validation - concrete tools can override this for specific validation logic
	// TODO: Implement JSON schema validation
	t.logger.Debug("Input validation", "tool", t.name, "input_type", fmt.Sprintf("%T", input))

	return nil
}

// validateConfig validates the tool configuration (can be overridden by concrete tools)
func (t *PonchoBaseTool) validateConfig(config map[string]interface{}) error {
	// Basic validation - concrete tools can override this for specific requirements
	if config == nil {
		return nil
	}

	// Validate timeout if provided
	if timeout, ok := config["timeout"].(string); ok {
		if timeout == "" {
			return fmt.Errorf("timeout cannot be empty")
		}
	}

	// Validate retry configuration if provided
	if retryConfig, ok := config["retry"].(map[string]interface{}); ok {
		if maxAttempts, ok := retryConfig["max_attempts"].(int); ok {
			if maxAttempts <= 0 {
				return fmt.Errorf("retry max_attempts must be positive")
			}
			if maxAttempts > 10 {
				return fmt.Errorf("retry max_attempts cannot exceed 10")
			}
		}

		if backoff, ok := retryConfig["backoff"].(string); ok {
			if backoff != "linear" && backoff != "exponential" {
				return fmt.Errorf("retry backoff must be 'linear' or 'exponential'")
			}
		}
	}

	return nil
}

// SetInputSchema sets the input schema
func (t *PonchoBaseTool) SetInputSchema(schema map[string]interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if schema != nil {
		t.inputSchema = make(map[string]interface{})
		for k, v := range schema {
			t.inputSchema[k] = v
		}
	} else {
		t.inputSchema = make(map[string]interface{})
	}
}

// SetOutputSchema sets the output schema
func (t *PonchoBaseTool) SetOutputSchema(schema map[string]interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if schema != nil {
		t.outputSchema = make(map[string]interface{})
		for k, v := range schema {
			t.outputSchema[k] = v
		}
	} else {
		t.outputSchema = make(map[string]interface{})
	}
}

// AddTag adds a tag to the tool
func (t *PonchoBaseTool) AddTag(tag string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Check if tag already exists
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return // Tag already exists
		}
	}

	t.tags = append(t.tags, tag)
}

// RemoveTag removes a tag from the tool
func (t *PonchoBaseTool) RemoveTag(tag string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for i, existingTag := range t.tags {
		if existingTag == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			break
		}
	}
}

// AddDependency adds a dependency to the tool
func (t *PonchoBaseTool) AddDependency(dependency string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Check if dependency already exists
	for _, existingDep := range t.dependencies {
		if existingDep == dependency {
			return // Dependency already exists
		}
	}

	t.dependencies = append(t.dependencies, dependency)
}

// RemoveDependency removes a dependency from the tool
func (t *PonchoBaseTool) RemoveDependency(dependency string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for i, existingDep := range t.dependencies {
		if existingDep == dependency {
			t.dependencies = append(t.dependencies[:i], t.dependencies[i+1:]...)
			break
		}
	}
}

// GetConfig gets a configuration value by key
func (t *PonchoBaseTool) GetConfig(key string) (interface{}, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	value, exists := t.config[key]
	return value, exists
}

// SetConfig sets a configuration value by key
func (t *PonchoBaseTool) SetConfig(key string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.config == nil {
		t.config = make(map[string]interface{})
	}

	t.config[key] = value
}

// GetAllConfig returns a copy of all configuration
func (t *PonchoBaseTool) GetAllConfig() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	config := make(map[string]interface{})
	for k, v := range t.config {
		config[k] = v
	}

	return config
}

// SetLogger sets the logger for the tool
func (t *PonchoBaseTool) SetLogger(logger interfaces.Logger) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.logger = logger
}

// GetLogger returns the current logger
func (t *PonchoBaseTool) GetLogger() interfaces.Logger {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.logger
}

// SetName sets the tool name
func (t *PonchoBaseTool) SetName(name string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.name = name
}

// SetDescription sets the tool description
func (t *PonchoBaseTool) SetDescription(description string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.description = description
}

// SetVersion sets the tool version
func (t *PonchoBaseTool) SetVersion(version string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.version = version
}

// SetCategory sets the tool category
func (t *PonchoBaseTool) SetCategory(category string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.category = category
}

// isInitialized checks if the tool is initialized (internal method)
func (t *PonchoBaseTool) isInitialized() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.initialized
}

// ValidateDependencies validates that all dependencies are available
func (t *PonchoBaseTool) ValidateDependencies(availableTools []string) error {
	t.mutex.RLock()
	dependencies := make([]string, len(t.dependencies))
	copy(dependencies, t.dependencies)
	t.mutex.RUnlock()

	for _, dep := range dependencies {
		found := false
		for _, available := range availableTools {
			if available == dep {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tool '%s' depends on '%s' which is not available", t.name, dep)
		}
	}

	return nil
}

// PrepareOutput prepares a standard output structure
func (t *PonchoBaseTool) PrepareOutput(data interface{}, metadata map[string]interface{}) map[string]interface{} {
	output := map[string]interface{}{
		"tool":      t.name,
		"version":   t.version,
		"timestamp": fmt.Sprintf("%d", 0), // TODO: Use actual timestamp
		"data":      data,
	}

	if metadata != nil {
		output["metadata"] = metadata
	}

	return output
}

// PrepareError prepares a standard error structure
func (t *PonchoBaseTool) PrepareError(err error, context map[string]interface{}) map[string]interface{} {
	errorOutput := map[string]interface{}{
		"tool":      t.name,
		"version":   t.version,
		"timestamp": fmt.Sprintf("%d", 0), // TODO: Use actual timestamp
		"error":     err.Error(),
		"success":   false,
	}

	if context != nil {
		errorOutput["context"] = context
	}

	return errorOutput
}

// LogExecution logs tool execution details
func (t *PonchoBaseTool) LogExecution(ctx context.Context, input interface{}, output interface{}, err error, duration int64) {
	if err != nil {
		t.logger.Error("Tool execution failed",
			"tool", t.name,
			"input_type", fmt.Sprintf("%T", input),
			"error", err.Error(),
			"duration_ms", duration,
		)
	} else {
		t.logger.Info("Tool executed successfully",
			"tool", t.name,
			"input_type", fmt.Sprintf("%T", input),
			"output_type", fmt.Sprintf("%T", output),
			"duration_ms", duration,
		)
	}
}

// HasTag checks if the tool has a specific tag
func (t *PonchoBaseTool) HasTag(tag string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, existingTag := range t.tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// HasDependency checks if the tool has a specific dependency
func (t *PonchoBaseTool) HasDependency(dependency string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, existingDep := range t.dependencies {
		if existingDep == dependency {
			return true
		}
	}
	return false
}

// GetTagCount returns the number of tags
func (t *PonchoBaseTool) GetTagCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return len(t.tags)
}

// GetDependencyCount returns the number of dependencies
func (t *PonchoBaseTool) GetDependencyCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return len(t.dependencies)
}

// ClearTags clears all tags
func (t *PonchoBaseTool) ClearTags() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tags = []string{}
}

// ClearDependencies clears all dependencies
func (t *PonchoBaseTool) ClearDependencies() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.dependencies = []string{}
}
