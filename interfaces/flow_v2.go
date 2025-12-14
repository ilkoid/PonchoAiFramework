package interfaces

import (
	"context"
	"fmt"
	"time"
)

// PonchoFlowV2 defines the enhanced interface for flows with state management
type PonchoFlowV2 interface {
	// Basic identity (same as v1)
	Name() string
	Description() string
	Version() string
	Category() string
	Tags() []string
	Dependencies() []string

	// Enhanced execution with FlowContext
	Execute(ctx context.Context, input interface{}, flowCtx FlowContext) (interface{}, error)
	ExecuteStreaming(ctx context.Context, input interface{}, flowCtx FlowContext, callback PonchoStreamCallback) error

	// Context lifecycle management
	CreateContext() FlowContext
	ValidateContext(flowCtx FlowContext) error
	GetRequiredContextKeys() []string
	GetProvidedContextKeys() []string

	// Enhanced schema validation
	InputSchema() map[string]interface{}
	OutputSchema() map[string]interface{}
	ContextSchema() map[string]interface{} // Defines what context keys are expected/required

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error

	// Enhanced metadata
	GetExecutionPattern() ExecutionPattern
	GetEstimatedDuration() string // "fast", "medium", "slow"
	GetResourceRequirements() ResourceRequirements
}

// ExecutionPattern defines how the flow executes its steps
type ExecutionPattern string

const (
	ExecutionPatternSequential ExecutionPattern = "sequential"
	ExecutionPatternParallel   ExecutionPattern = "parallel"
	ExecutionPatternHybrid     ExecutionPattern = "hybrid"
)

// ResourceRequirements defines what resources the flow needs
type ResourceRequirements struct {
	RequiresVision   bool     `json:"requires_vision"`
	RequiresTools    []string `json:"requires_tools"`
	RequiresModels   []string `json:"requires_models"`
	EstimatedMemory  string   `json:"estimated_memory"`  // "low", "medium", "high"
	EstimatedCPU     string   `json:"estimated_cpu"`     // "low", "medium", "high"
	MaxConcurrency   int      `json:"max_concurrency"`
	TimeoutSeconds   int      `json:"timeout_seconds"`
}

// FlowStep defines a single step in a flow
type FlowStep interface {
	Name() string
	Description() string
	Execute(ctx context.Context, flowCtx FlowContext) error
	CanFail() bool
	Dependencies() []string
	Timeout() int // in seconds
	RetryCount() int
}

// ConditionalStep represents a step that executes conditionally
type ConditionalStep interface {
	FlowStep
	Condition(flowCtx FlowContext) bool
	GetTrueSteps() []FlowStep
	GetFalseSteps() []FlowStep
}

// ParallelStep represents a step that executes multiple sub-steps in parallel
type ParallelStep interface {
	FlowStep
	GetSubSteps() []FlowStep
	MaxConcurrency() int
	FailFast() bool
	MergeStrategy() MergeStrategy
}

// MergeStrategy defines how to merge results from parallel steps
type MergeStrategy interface {
	Merge(results []interface{}, flowCtx FlowContext) (interface{}, error)
	Name() string
}

// FlowContext is imported from core/context package
// We use an interface here to avoid circular dependencies
type FlowContext interface {
	// Basic state management
	Set(key string, value interface{}) error
	Get(key string) (interface{}, bool)
	Delete(key string) bool
	Has(key string) bool
	Clear()
	Keys() []string
	Size() int

	// Type-safe operations
	SetString(key, value string) error
	GetString(key string) (string, error)
	SetBytes(key string, value []byte) error
	GetBytes(key string) ([]byte, error)
	SetInt(key string, value int) error
	GetInt(key string) (int, error)
	SetFloat(key string, value float64) error
	GetFloat(key string) (float64, error)
	SetBool(key string, value bool) error
	GetBool(key string) (bool, error)

	// Array/List operations
	SetArray(key string, values []interface{}) error
	GetArray(key string) ([]interface{}, error)
	AppendToArray(key string, value interface{}) error
	GetArraySize(key string) (int, error)

	// Object operations
	SetObject(key string, obj interface{}) error
	GetObject(key string, target interface{}) error

	// Media-specific operations
	SetMedia(key string, media *MediaData) error
	GetMedia(key string) (*MediaData, error)
	GetAllMedia(prefix string) ([]*MediaData, error)
	AccumulateMedia(prefix string, mediaList []*MediaData) error

	// Metadata and utilities
	Clone() FlowContext
	Merge(other FlowContext) error

	// Serialization for persistence/debugging
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	ToJSON() (string, error)

	// Context lifecycle
	ID() string
	CreatedAt() time.Time
	Parent() FlowContext
	CreateChild() FlowContext

	// Logging and debugging
	SetLogger(logger Logger)
	GetLogger() Logger
	Dump() map[string]interface{}
	PrintState()
}

// MediaData represents media content (avoid circular import)
type MediaData struct {
	URL      string                 `json:"url,omitempty"`
	Bytes    []byte                 `json:"-"`
	MimeType string                 `json:"mime_type"`
	Size     int64                  `json:"size"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BaseFlow provides a default implementation for common flow operations
type BaseFlow struct {
	name                 string
	description          string
	version              string
	category             string
	tags                 []string
	dependencies         []string
	steps                []FlowStep
	contextSchema        map[string]interface{}
	inputSchema          map[string]interface{}
	outputSchema         map[string]interface{}
	resourceRequirements ResourceRequirements
	executionPattern     ExecutionPattern
	logger               Logger
	initialized          bool
}

// NewBaseFlow creates a new base flow with default values
func NewBaseFlow(name, description, version, category string) *BaseFlow {
	return &BaseFlow{
		name:        name,
		description: description,
		version:     version,
		category:    category,
		tags:        make([]string, 0),
		dependencies: make([]string, 0),
		steps:       make([]FlowStep, 0),
		inputSchema: make(map[string]interface{}),
		outputSchema: make(map[string]interface{}),
		contextSchema: make(map[string]interface{}),
		resourceRequirements: ResourceRequirements{
			MaxConcurrency: 1,
			TimeoutSeconds: 300, // 5 minutes default
		},
		executionPattern: ExecutionPatternSequential,
		logger:          NewDefaultLogger(),
	}
}

// Getters and setters
func (bf *BaseFlow) Name() string { return bf.name }
func (bf *BaseFlow) Description() string { return bf.description }
func (bf *BaseFlow) Version() string { return bf.version }
func (bf *BaseFlow) Category() string { return bf.category }
func (bf *BaseFlow) Tags() []string { return bf.tags }
func (bf *BaseFlow) Dependencies() []string { return bf.dependencies }
func (bf *BaseFlow) InputSchema() map[string]interface{} { return bf.inputSchema }
func (bf *BaseFlow) OutputSchema() map[string]interface{} { return bf.outputSchema }
func (bf *BaseFlow) ContextSchema() map[string]interface{} { return bf.contextSchema }
func (bf *BaseFlow) GetExecutionPattern() ExecutionPattern { return bf.executionPattern }
func (bf *BaseFlow) GetEstimatedDuration() string { return "medium" }
func (bf *BaseFlow) GetResourceRequirements() ResourceRequirements { return bf.resourceRequirements }

func (bf *BaseFlow) AddTag(tag string) {
	bf.tags = append(bf.tags, tag)
}

func (bf *BaseFlow) AddDependency(dependency string) {
	bf.dependencies = append(bf.dependencies, dependency)
}

func (bf *BaseFlow) AddStep(step FlowStep) {
	bf.steps = append(bf.steps, step)
}

func (bf *BaseFlow) SetExecutionPattern(pattern ExecutionPattern) {
	bf.executionPattern = pattern
}

func (bf *BaseFlow) SetResourceRequirements(req ResourceRequirements) {
	bf.resourceRequirements = req
}

func (bf *BaseFlow) GetRequiredContextKeys() []string {
	var keys []string
	// Extract keys from context schema if defined
	if bf.contextSchema != nil {
		if required, ok := bf.contextSchema["required"].([]string); ok {
			keys = append(keys, required...)
		}
	}
	return keys
}

func (bf *BaseFlow) GetProvidedContextKeys() []string {
	var keys []string
	// Extract keys that this flow provides to context
	for _, step := range bf.steps {
		// This would need to be implemented by specific step types
		// For now, we'll use step name as a key
		keys = append(keys, step.Name()+"_result")
	}
	return keys
}

// Default implementations
func (bf *BaseFlow) CreateContext() FlowContext {
	// Return a simple context implementation
	// In real implementation, this would return core/context.BaseFlowContext
	return nil // TODO: Import and return actual context
}

func (bf *BaseFlow) ValidateContext(flowCtx FlowContext) error {
	if flowCtx == nil {
		return fmt.Errorf("flow context cannot be nil")
	}

	// Validate required context keys
	requiredKeys := bf.GetRequiredContextKeys()
	for _, key := range requiredKeys {
		if !flowCtx.Has(key) {
			return fmt.Errorf("required context key '%s' is missing", key)
		}
	}

	return nil
}

func (bf *BaseFlow) Initialize(ctx context.Context, config map[string]interface{}) error {
	if bf.initialized {
		return fmt.Errorf("flow '%s' is already initialized", bf.name)
	}

	// Initialize based on config
	if logger, ok := config["logger"].(Logger); ok {
		bf.logger = logger
	}

	bf.initialized = true
	bf.logger.Info("Flow initialized", "name", bf.name, "version", bf.version)
	return nil
}

func (bf *BaseFlow) Shutdown(ctx context.Context) error {
	if !bf.initialized {
		return nil
	}

	bf.initialized = false
	bf.logger.Info("Flow shutdown", "name", bf.name)
	return nil
}

// ExecuteSequential provides default sequential execution
func (bf *BaseFlow) ExecuteSequential(ctx context.Context, input interface{}, flowCtx FlowContext) (interface{}, error) {
	// Store initial input in context
	if err := flowCtx.Set("input", input); err != nil {
		return nil, fmt.Errorf("failed to store input in context: %w", err)
	}

	// Execute steps sequentially
	for _, step := range bf.steps {
		bf.logger.Debug("Executing step", "step", step.Name(), "flow", bf.name)

		if err := step.Execute(ctx, flowCtx); err != nil {
			if step.CanFail() {
				bf.logger.Warn("Step failed but can continue",
					"step", step.Name(),
					"error", err.Error())
				continue
			}
			return nil, fmt.Errorf("step '%s' failed: %w", step.Name(), err)
		}

		bf.logger.Debug("Step completed successfully", "step", step.Name())
	}

	// Return final result from context
	if flowCtx.Has("output") {
		result, _ := flowCtx.Get("output")
		return result, nil
	}

	// Default return: entire context
	return flowCtx, nil
}

// Execute provides a basic implementation
func (bf *BaseFlow) Execute(ctx context.Context, input interface{}, flowCtx FlowContext) (interface{}, error) {
	if !bf.initialized {
		return nil, fmt.Errorf("flow '%s' is not initialized", bf.name)
	}

	if err := bf.ValidateContext(flowCtx); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	switch bf.executionPattern {
	case ExecutionPatternSequential:
		return bf.ExecuteSequential(ctx, input, flowCtx)
	case ExecutionPatternParallel:
		// TODO: Implement parallel execution
		return bf.ExecuteSequential(ctx, input, flowCtx)
	default:
		return bf.ExecuteSequential(ctx, input, flowCtx)
	}
}

// ExecuteStreaming provides default streaming implementation
func (bf *BaseFlow) ExecuteStreaming(
	ctx context.Context,
	input interface{},
	flowCtx FlowContext,
	callback PonchoStreamCallback,
) error {
	// For now, just execute non-streaming and send result as single message
	result, err := bf.Execute(ctx, input, flowCtx)
	if err != nil {
		return err
	}

	// Convert result to JSON string for streaming
	resultJSON := fmt.Sprintf("%v", result)
	callback(&PonchoStreamChunk{
		Delta: &PonchoMessage{
			Role: PonchoRoleAssistant,
			Content: []*PonchoContentPart{
				{
					Type: PonchoContentTypeText,
					Text: resultJSON,
				},
			},
		},
		Done: true,
		Metadata: map[string]interface{}{
			"flow":    bf.name,
			"version": bf.version,
		},
	})

	return nil
}