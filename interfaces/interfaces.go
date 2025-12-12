package interfaces

import (
	"context"
)

// PonchoModel defines the interface for all AI models in the framework
type PonchoModel interface {
	// Core generation methods
	Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
	GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error

	// Capability checks
	SupportsStreaming() bool
	SupportsTools() bool
	SupportsVision() bool
	SupportsSystemRole() bool

	// Metadata
	Name() string
	Provider() string
	MaxTokens() int
	DefaultTemperature() float32

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
}

// PonchoTool defines the interface for all tools in the framework
type PonchoTool interface {
	// Identity
	Name() string
	Description() string
	Version() string

	// Core execution
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// Schema validation
	InputSchema() map[string]interface{}
	OutputSchema() map[string]interface{}
	Validate(input interface{}) error

	// Metadata
	Category() string
	Tags() []string
	Dependencies() []string

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
}

// PonchoFlow defines the interface for all workflows in the framework
type PonchoFlow interface {
	// Identity
	Name() string
	Description() string
	Version() string

	// Core execution
	Execute(ctx context.Context, input interface{}) (interface{}, error)
	ExecuteStreaming(ctx context.Context, input interface{}, callback PonchoStreamCallback) error

	// Schema validation
	InputSchema() map[string]interface{}
	OutputSchema() map[string]interface{}

	// Metadata
	Category() string
	Tags() []string
	Dependencies() []string

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
}

// PonchoStreamCallback is the callback function for streaming responses
type PonchoStreamCallback func(chunk *PonchoStreamChunk) error

// Registry interfaces

// PonchoModelRegistry defines the interface for model registry
type PonchoModelRegistry interface {
	Register(name string, model PonchoModel) error
	Get(name string) (PonchoModel, error)
	List() []string
	Unregister(name string) error
	Clear() error
}

// PonchoToolRegistry defines the interface for tool registry
type PonchoToolRegistry interface {
	Register(name string, tool PonchoTool) error
	Get(name string) (PonchoTool, error)
	List() []string
	ListByCategory(category string) []string
	Unregister(name string) error
	Clear() error
}

// PonchoFlowRegistry defines the interface for flow registry
type PonchoFlowRegistry interface {
	Register(name string, flow PonchoFlow) error
	Get(name string) (PonchoFlow, error)
	List() []string
	ListByCategory(category string) []string
	Unregister(name string) error
	Clear() error
	ValidateDependencies(flow PonchoFlow, modelRegistry PonchoModelRegistry, toolRegistry PonchoToolRegistry) error
}

// PonchoFramework defines the main framework interface
type PonchoFramework interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Component registration
	RegisterModel(name string, model PonchoModel) error
	RegisterTool(name string, tool PonchoTool) error
	RegisterFlow(name string, flow PonchoFlow) error

	// Core operations
	Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
	GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error
	ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
	ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error)
	ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback PonchoStreamCallback) error

	// Registry access
	GetModelRegistry() PonchoModelRegistry
	GetToolRegistry() PonchoToolRegistry
	GetFlowRegistry() PonchoFlowRegistry

	// Configuration
	GetConfig() *PonchoFrameworkConfig
	ReloadConfig(ctx context.Context) error

	// Health and status
	Health(ctx context.Context) (*PonchoHealthStatus, error)
	Metrics(ctx context.Context) (*PonchoMetrics, error)
}

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// Validator interface for schema validation
type Validator interface {
	Validate(data interface{}, schema map[string]interface{}) error
	SetSchema(schema map[string]interface{}) error
	GetSchema() map[string]interface{}
}

// ConfigLoader interface for configuration loading
type ConfigLoader interface {
	Load(path string) (*PonchoFrameworkConfig, error)
	LoadFromString(content string) (*PonchoFrameworkConfig, error)
	Validate(config *PonchoFrameworkConfig) error
}

// MetricsCollector interface for metrics collection
type MetricsCollector interface {
	RecordGeneration(model string, duration int64, tokens int, success bool)
	RecordToolExecution(tool string, duration int64, success bool)
	RecordFlowExecution(flow string, duration int64, success bool)
	RecordError(component string, errorType string)
	GetMetrics() *PonchoMetrics
	Reset()
}
