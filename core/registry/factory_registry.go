package registry

import (
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ModelFactoryRegistryImpl implements the ModelFactoryManager interface
type ModelFactoryRegistryImpl struct {
	factories map[string]interfaces.ModelFactory
	logger    interfaces.Logger
	mutex     sync.RWMutex
}

// NewModelFactoryRegistry creates a new model factory registry
func NewModelFactoryRegistry(logger interfaces.Logger) *ModelFactoryRegistryImpl {
	return &ModelFactoryRegistryImpl{
		factories: make(map[string]interfaces.ModelFactory),
		logger:    logger,
	}
}

// RegisterFactory registers a model factory
func (r *ModelFactoryRegistryImpl) RegisterFactory(provider string, factory interfaces.ModelFactory) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.logger.Debug("Registering model factory", "provider", provider)
	r.factories[provider] = factory
	r.logger.Info("Model factory registered", "provider", provider)

	return nil
}

// GetFactory returns a factory for the specified provider
func (r *ModelFactoryRegistryImpl) GetFactory(provider string) (interfaces.ModelFactory, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.factories[provider]
	if !exists {
		return nil, fmt.Errorf("no factory registered for provider: %s", provider)
	}

	return factory, nil
}

// GetSupportedProviders returns list of supported providers
func (r *ModelFactoryRegistryImpl) GetSupportedProviders() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	providers := make([]string, 0, len(r.factories))
	for provider := range r.factories {
		providers = append(providers, provider)
	}

	return providers
}

// CreateModel creates a model using the appropriate factory
func (r *ModelFactoryRegistryImpl) CreateModel(config *interfaces.ModelConfig) (interfaces.PonchoModel, error) {
	if config == nil {
		return nil, fmt.Errorf("model config cannot be nil")
	}

	if config.Provider == "" {
		return nil, fmt.Errorf("model provider is required")
	}

	factory, err := r.GetFactory(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get factory for provider %s: %w", config.Provider, err)
	}

	r.logger.Debug("Creating model using factory",
		"provider", config.Provider,
		"model_name", config.ModelName)

	model, err := factory.CreateModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create model with factory %s: %w", config.Provider, err)
	}

	r.logger.Info("Model created successfully",
		"provider", config.Provider,
		"model_name", config.ModelName,
		"model", model.Name())

	return model, nil
}

// ValidateConfig validates configuration using the appropriate factory
func (r *ModelFactoryRegistryImpl) ValidateConfig(config *interfaces.ModelConfig) error {
	if config == nil {
		return fmt.Errorf("model config cannot be nil")
	}

	if config.Provider == "" {
		return fmt.Errorf("model provider is required")
	}

	factory, err := r.GetFactory(config.Provider)
	if err != nil {
		return fmt.Errorf("failed to get factory for provider %s: %w", config.Provider, err)
	}

	return factory.ValidateConfig(config)
}

// ToolFactoryRegistryImpl implements the ToolFactoryManager interface
type ToolFactoryRegistryImpl struct {
	factories map[string]interfaces.ToolFactory
	logger    interfaces.Logger
	mutex     sync.RWMutex
}

// NewToolFactoryRegistry creates a new tool factory registry
func NewToolFactoryRegistry(logger interfaces.Logger) *ToolFactoryRegistryImpl {
	return &ToolFactoryRegistryImpl{
		factories: make(map[string]interfaces.ToolFactory),
		logger:    logger,
	}
}

// RegisterFactory registers a tool factory
func (r *ToolFactoryRegistryImpl) RegisterFactory(toolType string, factory interfaces.ToolFactory) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if toolType == "" {
		return fmt.Errorf("tool type cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.logger.Debug("Registering tool factory", "tool_type", toolType)
	r.factories[toolType] = factory
	r.logger.Info("Tool factory registered", "tool_type", toolType)

	return nil
}

// GetFactory returns a factory for the specified tool type
func (r *ToolFactoryRegistryImpl) GetFactory(toolType string) (interfaces.ToolFactory, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.factories[toolType]
	if !exists {
		return nil, fmt.Errorf("no factory registered for tool type: %s", toolType)
	}

	return factory, nil
}

// GetSupportedToolTypes returns list of supported tool types
func (r *ToolFactoryRegistryImpl) GetSupportedToolTypes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	toolTypes := make([]string, 0, len(r.factories))
	for toolType := range r.factories {
		toolTypes = append(toolTypes, toolType)
	}

	return toolTypes
}

// CreateTool creates a tool using the appropriate factory
func (r *ToolFactoryRegistryImpl) CreateTool(config *interfaces.ToolConfig, toolType string) (interfaces.PonchoTool, error) {
	if config == nil {
		return nil, fmt.Errorf("tool config cannot be nil")
	}

	if toolType == "" {
		// Try to get tool type from custom params
		if config.CustomParams != nil {
			if tt, ok := config.CustomParams["type"].(string); ok {
				toolType = tt
			}
		}
	}

	if toolType == "" {
		return nil, fmt.Errorf("tool type is required (must be passed as parameter or in config.CustomParams['type'])")
	}

	factory, err := r.GetFactory(toolType)
	if err != nil {
		return nil, fmt.Errorf("failed to get factory for tool type %s: %w", toolType, err)
	}

	r.logger.Debug("Creating tool using factory",
		"tool_type", toolType)

	tool, err := factory.CreateTool(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool with factory %s: %w", toolType, err)
	}

	r.logger.Info("Tool created successfully",
		"tool_type", toolType)

	return tool, nil
}

// ValidateConfig validates configuration using the appropriate factory
func (r *ToolFactoryRegistryImpl) ValidateConfig(config *interfaces.ToolConfig, toolType string) error {
	if config == nil {
		return fmt.Errorf("tool config cannot be nil")
	}

	if toolType == "" {
		// Try to get tool type from custom params
		if config.CustomParams != nil {
			if tt, ok := config.CustomParams["type"].(string); ok {
				toolType = tt
			}
		}
	}

	if toolType == "" {
		return fmt.Errorf("tool type is required (must be passed as parameter or in config.CustomParams['type'])")
	}

	factory, err := r.GetFactory(toolType)
	if err != nil {
		return fmt.Errorf("failed to get factory for tool type %s: %w", toolType, err)
	}

	return factory.ValidateConfig(config)
}