package interfaces

// ModelFactory defines the interface for creating model instances
type ModelFactory interface {
	// CreateModel creates a model instance from configuration
	CreateModel(config *ModelConfig) (PonchoModel, error)

	// ValidateConfig validates model-specific configuration
	ValidateConfig(config *ModelConfig) error

	// GetProvider returns the provider this factory supports
	GetProvider() string
}

// ToolFactory defines the interface for creating tool instances
type ToolFactory interface {
	// CreateTool creates a tool instance from configuration
	CreateTool(config *ToolConfig) (PonchoTool, error)

	// ValidateConfig validates tool-specific configuration
	ValidateConfig(config *ToolConfig) error

	// GetToolType returns the type of tool this factory creates
	GetToolType() string
}

// ModelFactoryManager defines the interface for managing model factories
type ModelFactoryManager interface {
	// RegisterFactory registers a model factory for a provider
	RegisterFactory(provider string, factory ModelFactory) error

	// GetFactory returns a factory for the specified provider
	GetFactory(provider string) (ModelFactory, error)

	// GetSupportedProviders returns list of supported providers
	GetSupportedProviders() []string

	// CreateModel creates a model using the appropriate factory
	CreateModel(config *ModelConfig) (PonchoModel, error)

	// ValidateConfig validates configuration using the appropriate factory
	ValidateConfig(config *ModelConfig) error
}

// ToolFactoryManager defines the interface for managing tool factories
type ToolFactoryManager interface {
	// RegisterFactory registers a tool factory for a type
	RegisterFactory(toolType string, factory ToolFactory) error

	// GetFactory returns a factory for the specified tool type
	GetFactory(toolType string) (ToolFactory, error)

	// GetSupportedToolTypes returns list of supported tool types
	GetSupportedToolTypes() []string

	// CreateTool creates a tool using the appropriate factory
	// toolType can be passed separately or extracted from config.CustomParams["type"]
	CreateTool(config *ToolConfig, toolType string) (PonchoTool, error)

	// ValidateConfig validates configuration using the appropriate factory
	// toolType can be passed separately or extracted from config.CustomParams["type"]
	ValidateConfig(config *ToolConfig, toolType string) error
}