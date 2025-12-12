package base

import (
	"context"
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoBaseFlow provides a base implementation of PonchoFlow interface
type PonchoBaseFlow struct {
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

// NewPonchoBaseFlow creates a new base flow instance
func NewPonchoBaseFlow(name, description, version, category string) *PonchoBaseFlow {
	return &PonchoBaseFlow{
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

// Name returns the flow name
func (f *PonchoBaseFlow) Name() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.name
}

// Description returns the flow description
func (f *PonchoBaseFlow) Description() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.description
}

// Version returns the flow version
func (f *PonchoBaseFlow) Version() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.version
}

// Category returns the flow category
func (f *PonchoBaseFlow) Category() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.category
}

// Tags returns the flow tags
func (f *PonchoBaseFlow) Tags() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Return a copy to prevent external modification
	tags := make([]string, len(f.tags))
	copy(tags, f.tags)
	return tags
}

// Dependencies returns the flow dependencies
func (f *PonchoBaseFlow) Dependencies() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Return a copy to prevent external modification
	deps := make([]string, len(f.dependencies))
	copy(deps, f.dependencies)
	return deps
}

// InputSchema returns the input schema
func (f *PonchoBaseFlow) InputSchema() map[string]interface{} {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Return a copy to prevent external modification
	schema := make(map[string]interface{})
	for k, v := range f.inputSchema {
		schema[k] = v
	}
	return schema
}

// OutputSchema returns the output schema
func (f *PonchoBaseFlow) OutputSchema() map[string]interface{} {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Return a copy to prevent external modification
	schema := make(map[string]interface{})
	for k, v := range f.outputSchema {
		schema[k] = v
	}
	return schema
}

// Initialize initializes the flow with configuration
func (f *PonchoBaseFlow) Initialize(ctx context.Context, config map[string]interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.initialized {
		return fmt.Errorf("flow '%s' is already initialized", f.name)
	}

	// Validate configuration
	if err := f.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config for flow '%s': %w", f.name, err)
	}

	// Set configuration
	if config != nil {
		f.config = make(map[string]interface{})
		for k, v := range config {
			f.config[k] = v
		}
	}

	f.initialized = true
	f.logger.Info("Flow initialized", "flow", f.name, "version", f.version, "category", f.category)

	return nil
}

// Shutdown shuts down the flow and cleans up resources
func (f *PonchoBaseFlow) Shutdown(ctx context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !f.initialized {
		return nil
	}

	// Clear configuration
	f.config = make(map[string]interface{})
	f.initialized = false

	f.logger.Info("Flow shutdown", "flow", f.name, "version", f.version)

	return nil
}

// Execute executes the flow (must be implemented by concrete flows)
func (f *PonchoBaseFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if !f.isInitialized() {
		return nil, fmt.Errorf("flow '%s' is not initialized", f.name)
	}

	// Validate input against schema
	if err := f.validateInput(input); err != nil {
		return nil, fmt.Errorf("input validation failed for flow '%s': %w", f.name, err)
	}

	// This is a base implementation - concrete flows should override this
	return nil, fmt.Errorf("Execute method must be implemented by concrete flow")
}

// ExecuteStreaming executes the flow with streaming (must be implemented by concrete flows)
func (f *PonchoBaseFlow) ExecuteStreaming(ctx context.Context, input interface{}, callback interfaces.PonchoStreamCallback) error {
	if !f.isInitialized() {
		return fmt.Errorf("flow '%s' is not initialized", f.name)
	}

	// Validate input against schema
	if err := f.validateInput(input); err != nil {
		return fmt.Errorf("input validation failed for flow '%s': %w", f.name, err)
	}

	// This is a base implementation - concrete flows should override this
	return fmt.Errorf("ExecuteStreaming method must be implemented by concrete flow")
}

// validateInput validates input against the input schema
func (f *PonchoBaseFlow) validateInput(input interface{}) error {
	if f.inputSchema == nil || len(f.inputSchema) == 0 {
		return nil // No schema - skip validation
	}

	// Basic validation - concrete flows can override this for specific validation logic
	// TODO: Implement JSON schema validation
	f.logger.Debug("Input validation", "flow", f.name, "input_type", fmt.Sprintf("%T", input))

	return nil
}

// validateConfig validates the flow configuration (can be overridden by concrete flows)
func (f *PonchoBaseFlow) validateConfig(config map[string]interface{}) error {
	// Basic validation - concrete flows can override this for specific requirements
	if config == nil {
		return nil
	}

	// Validate timeout if provided
	if timeout, ok := config["timeout"].(string); ok {
		if timeout == "" {
			return fmt.Errorf("timeout cannot be empty")
		}
	}

	// Validate parallel execution setting if provided
	if parallel, ok := config["parallel"].(bool); ok {
		// Parallel execution is valid, no specific validation needed
		_ = parallel
	}

	// Validate dependencies if provided
	if deps, ok := config["dependencies"].([]string); ok {
		for _, dep := range deps {
			if dep == "" {
				return fmt.Errorf("dependency cannot be empty")
			}
		}
	}

	return nil
}

// SetInputSchema sets the input schema
func (f *PonchoBaseFlow) SetInputSchema(schema map[string]interface{}) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if schema != nil {
		f.inputSchema = make(map[string]interface{})
		for k, v := range schema {
			f.inputSchema[k] = v
		}
	} else {
		f.inputSchema = make(map[string]interface{})
	}
}

// SetOutputSchema sets the output schema
func (f *PonchoBaseFlow) SetOutputSchema(schema map[string]interface{}) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if schema != nil {
		f.outputSchema = make(map[string]interface{})
		for k, v := range schema {
			f.outputSchema[k] = v
		}
	} else {
		f.outputSchema = make(map[string]interface{})
	}
}

// AddTag adds a tag to the flow
func (f *PonchoBaseFlow) AddTag(tag string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check if tag already exists
	for _, existingTag := range f.tags {
		if existingTag == tag {
			return // Tag already exists
		}
	}

	f.tags = append(f.tags, tag)
}

// RemoveTag removes a tag from the flow
func (f *PonchoBaseFlow) RemoveTag(tag string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i, existingTag := range f.tags {
		if existingTag == tag {
			f.tags = append(f.tags[:i], f.tags[i+1:]...)
			break
		}
	}
}

// AddDependency adds a dependency to the flow
func (f *PonchoBaseFlow) AddDependency(dependency string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check if dependency already exists
	for _, existingDep := range f.dependencies {
		if existingDep == dependency {
			return // Dependency already exists
		}
	}

	f.dependencies = append(f.dependencies, dependency)
}

// RemoveDependency removes a dependency from the flow
func (f *PonchoBaseFlow) RemoveDependency(dependency string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i, existingDep := range f.dependencies {
		if existingDep == dependency {
			f.dependencies = append(f.dependencies[:i], f.dependencies[i+1:]...)
			break
		}
	}
}

// GetConfig gets a configuration value by key
func (f *PonchoBaseFlow) GetConfig(key string) (interface{}, bool) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	value, exists := f.config[key]
	return value, exists
}

// SetConfig sets a configuration value by key
func (f *PonchoBaseFlow) SetConfig(key string, value interface{}) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.config == nil {
		f.config = make(map[string]interface{})
	}

	f.config[key] = value
}

// GetAllConfig returns a copy of all configuration
func (f *PonchoBaseFlow) GetAllConfig() map[string]interface{} {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	config := make(map[string]interface{})
	for k, v := range f.config {
		config[k] = v
	}

	return config
}

// SetLogger sets the logger for the flow
func (f *PonchoBaseFlow) SetLogger(logger interfaces.Logger) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.logger = logger
}

// GetLogger returns the current logger
func (f *PonchoBaseFlow) GetLogger() interfaces.Logger {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.logger
}

// SetName sets the flow name
func (f *PonchoBaseFlow) SetName(name string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.name = name
}

// SetDescription sets the flow description
func (f *PonchoBaseFlow) SetDescription(description string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.description = description
}

// SetVersion sets the flow version
func (f *PonchoBaseFlow) SetVersion(version string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.version = version
}

// SetCategory sets the flow category
func (f *PonchoBaseFlow) SetCategory(category string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.category = category
}

// isInitialized checks if the flow is initialized (internal method)
func (f *PonchoBaseFlow) isInitialized() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.initialized
}

// ValidateDependencies validates that all dependencies are available
func (f *PonchoBaseFlow) ValidateDependencies(availableFlows []string, availableTools []string) error {
	f.mutex.RLock()
	dependencies := make([]string, len(f.dependencies))
	copy(dependencies, f.dependencies)
	f.mutex.RUnlock()

	for _, dep := range dependencies {
		// Check if dependency is a flow
		found := false
		for _, available := range availableFlows {
			if available == dep {
				found = true
				break
			}
		}

		// Check if dependency is a tool
		if !found {
			for _, available := range availableTools {
				if available == dep {
					found = true
					break
				}
			}
		}

		if !found {
			return fmt.Errorf("flow '%s' depends on '%s' which is not available", f.name, dep)
		}
	}

	return nil
}

// PrepareOutput prepares a standard output structure
func (f *PonchoBaseFlow) PrepareOutput(data interface{}, metadata map[string]interface{}) map[string]interface{} {
	output := map[string]interface{}{
		"flow":      f.name,
		"version":   f.version,
		"timestamp": fmt.Sprintf("%d", 0), // TODO: Use actual timestamp
		"data":      data,
	}

	if metadata != nil {
		output["metadata"] = metadata
	}

	return output
}

// PrepareError prepares a standard error structure
func (f *PonchoBaseFlow) PrepareError(err error, context map[string]interface{}) map[string]interface{} {
	errorOutput := map[string]interface{}{
		"flow":      f.name,
		"version":   f.version,
		"timestamp": fmt.Sprintf("%d", 0), // TODO: Use actual timestamp
		"error":     err.Error(),
		"success":   false,
	}

	if context != nil {
		errorOutput["context"] = context
	}

	return errorOutput
}

// LogExecution logs flow execution details
func (f *PonchoBaseFlow) LogExecution(ctx context.Context, input interface{}, output interface{}, err error, duration int64) {
	if err != nil {
		f.logger.Error("Flow execution failed",
			"flow", f.name,
			"input_type", fmt.Sprintf("%T", input),
			"error", err.Error(),
			"duration_ms", duration,
		)
	} else {
		f.logger.Info("Flow executed successfully",
			"flow", f.name,
			"input_type", fmt.Sprintf("%T", input),
			"output_type", fmt.Sprintf("%T", output),
			"duration_ms", duration,
		)
	}
}

// HasTag checks if the flow has a specific tag
func (f *PonchoBaseFlow) HasTag(tag string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for _, existingTag := range f.tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// HasDependency checks if the flow has a specific dependency
func (f *PonchoBaseFlow) HasDependency(dependency string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for _, existingDep := range f.dependencies {
		if existingDep == dependency {
			return true
		}
	}
	return false
}

// GetTagCount returns the number of tags
func (f *PonchoBaseFlow) GetTagCount() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return len(f.tags)
}

// GetDependencyCount returns the number of dependencies
func (f *PonchoBaseFlow) GetDependencyCount() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return len(f.dependencies)
}

// ClearTags clears all tags
func (f *PonchoBaseFlow) ClearTags() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.tags = []string{}
}

// ClearDependencies clears all dependencies
func (f *PonchoBaseFlow) ClearDependencies() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.dependencies = []string{}
}

// ValidateOutput validates output against the output schema
func (f *PonchoBaseFlow) ValidateOutput(output interface{}) error {
	if f.outputSchema == nil || len(f.outputSchema) == 0 {
		return nil // No schema - skip validation
	}

	// Basic validation - concrete flows can override this for specific validation logic
	// TODO: Implement JSON schema validation
	f.logger.Debug("Output validation", "flow", f.name, "output_type", fmt.Sprintf("%T", output))

	return nil
}

// IsParallelEnabled checks if parallel execution is enabled
func (f *PonchoBaseFlow) IsParallelEnabled() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if parallel, ok := f.config["parallel"].(bool); ok {
		return parallel
	}
	return false
}

// GetTimeout returns the configured timeout duration
func (f *PonchoBaseFlow) GetTimeout() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if timeout, ok := f.config["timeout"].(string); ok {
		return timeout
	}
	return "" // No timeout configured
}

// SetParallelEnabled enables or disables parallel execution
func (f *PonchoBaseFlow) SetParallelEnabled(parallel bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.config == nil {
		f.config = make(map[string]interface{})
	}
	f.config["parallel"] = parallel
}

// SetTimeout sets the timeout duration
func (f *PonchoBaseFlow) SetTimeout(timeout string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.config == nil {
		f.config = make(map[string]interface{})
	}
	f.config["timeout"] = timeout
}
