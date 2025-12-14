package flow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// FlowBuilder provides a fluent API for building flows
type FlowBuilder struct {
	name         string
	description  string
	version      string
	category     string
	steps        []FlowStep
	config       *FlowConfig
	requirements ResourceRequirements
	logger       interfaces.Logger
}

// FlowConfig provides configuration for built flows
type FlowConfig struct {
	Timeout         time.Duration
	MaxRetries      int
	EnableParallel  bool
	MaxConcurrency  int
	FailFast        bool
	EnableRecovery  bool
	Metadata        map[string]interface{}
}

// ResourceRequirements defines resource needs for the flow
type ResourceRequirements struct {
	RequiresVision  bool     `json:"requires_vision"`
	RequiresTools   []string `json:"requires_tools"`
	RequiresModels  []string `json:"requires_models"`
	EstimatedMemory string   `json:"estimated_memory"`
	EstimatedCPU    string   `json:"estimated_cpu"`
	MaxConcurrency  int      `json:"max_concurrency"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
}

// NewFlowBuilder creates a new flow builder
func NewFlowBuilder(name string) *FlowBuilder {
	return &FlowBuilder{
		name:     name,
		version:  "1.0.0",
		category: "general",
		steps:    make([]FlowStep, 0),
		config: &FlowConfig{
			Timeout:         5 * time.Minute,
			MaxRetries:      3,
			EnableParallel:  false,
			MaxConcurrency:  1,
			FailFast:        true,
			EnableRecovery:  false,
		},
		requirements: ResourceRequirements{
			MaxConcurrency: 1,
			TimeoutSeconds: 300,
		},
		logger: interfaces.NewDefaultLogger(),
	}
}

// Fluent API methods
func (fb *FlowBuilder) Description(desc string) *FlowBuilder {
	fb.description = desc
	return fb
}

func (fb *FlowBuilder) Version(version string) *FlowBuilder {
	fb.version = version
	return fb
}

func (fb *FlowBuilder) Category(category string) *FlowBuilder {
	fb.category = category
	return fb
}

func (fb *FlowBuilder) Timeout(timeout time.Duration) *FlowBuilder {
	fb.config.Timeout = timeout
	fb.requirements.TimeoutSeconds = int(timeout.Seconds())
	return fb
}

func (fb *FlowBuilder) EnableParallel(enable bool) *FlowBuilder {
	fb.config.EnableParallel = enable
	if enable {
		fb.requirements.MaxConcurrency = fb.config.MaxConcurrency
	}
	return fb
}

func (fb *FlowBuilder) MaxConcurrency(max int) *FlowBuilder {
	fb.config.MaxConcurrency = max
	fb.requirements.MaxConcurrency = max
	return fb
}

func (fb *FlowBuilder) FailFast(failFast bool) *FlowBuilder {
	fb.config.FailFast = failFast
	return fb
}

func (fb *FlowBuilder) RequiresVision() *FlowBuilder {
	fb.requirements.RequiresVision = true
	return fb
}

func (fb *FlowBuilder) RequiresTool(toolName string) *FlowBuilder {
	fb.requirements.RequiresTools = append(fb.requirements.RequiresTools, toolName)
	return fb
}

func (fb *FlowBuilder) RequiresModel(modelName string) *FlowBuilder {
	fb.requirements.RequiresModels = append(fb.requirements.RequiresModels, modelName)
	return fb
}

// Step creates a new step builder
func (fb *FlowBuilder) Step(name string) *StepBuilder {
	return &StepBuilder{
		stepName: name,
		builder:  fb,
	}
}

// Build constructs the flow from the builder configuration
func (fb *FlowBuilder) Build() (interfaces.PonchoFlowV2, error) {
	if len(fb.steps) == 0 {
		return nil, fmt.Errorf("flow '%s' has no steps", fb.name)
	}

	// Create base flow
	baseFlow := interfaces.NewBaseFlow(fb.name, fb.description, fb.version, fb.category)

	// Set resource requirements
	baseFlow.SetResourceRequirements(interfaces.ResourceRequirements{
		RequiresVision:   fb.requirements.RequiresVision,
		RequiresTools:    fb.requirements.RequiresTools,
		RequiresModels:   fb.requirements.RequiresModels,
		EstimatedMemory:  fb.requirements.EstimatedMemory,
		EstimatedCPU:     fb.requirements.EstimatedCPU,
		MaxConcurrency:   fb.requirements.MaxConcurrency,
		TimeoutSeconds:   fb.requirements.TimeoutSeconds,
	})

	// Set execution pattern
	pattern := interfaces.ExecutionPatternSequential
	if fb.config.EnableParallel {
		pattern = interfaces.ExecutionPatternHybrid
	}
	baseFlow.SetExecutionPattern(pattern)

	// Add steps to base flow
	for _, step := range fb.steps {
		baseFlow.AddStep(step)
	}

	fb.logger.Info("Flow built successfully",
		"name", fb.name,
		"steps", len(fb.steps),
		"pattern", pattern,
	)

	return baseFlow, nil
}

// StepBuilder provides fluent API for configuring individual steps
type StepBuilder struct {
	stepName   string
	builder    *FlowBuilder
	step       FlowStep
	stepConfig *StepConfig
}

// StepConfig provides configuration for individual steps
type StepConfig struct {
	Timeout     time.Duration
	MaxRetries  int
	CanFail     bool
	Parallel    bool
	Conditions  []string  // Context keys that must exist
	Outputs     []string  // Context keys that this step provides
	Metadata    map[string]interface{}
}

// Tool creates a step that executes a tool
func (sb *StepBuilder) Tool(tool interfaces.PonchoTool, inputKey string) *ToolStepBuilder {
	step := &ToolStep{
		name:      sb.stepName,
		tool:      tool,
		inputKey:  inputKey,
		outputKey: sb.stepName + "_result",
		config:    sb.stepConfig,
	}

	sb.step = step
	return &ToolStepBuilder{
		step:    step,
		builder: sb.builder,
	}
}

// Model creates a step that executes a model
func (sb *StepBuilder) Model(model interfaces.PonchoModel, promptKey string) *ModelStepBuilder {
	step := &ModelStep{
		name:       sb.stepName,
		model:      model,
		promptKey:  promptKey,
		outputKey:  sb.stepName + "_result",
		config:     sb.stepConfig,
	}

	sb.step = step
	return &ModelStepBuilder{
		step:    step,
		builder: sb.builder,
	}
}

// Custom creates a step with custom execution logic
func (sb *StepBuilder) Custom(executor StepExecutor) *CustomStepBuilder {
	step := &CustomStep{
		name:     sb.stepName,
		executor: executor,
		config:   sb.stepConfig,
	}

	sb.step = step
	return &CustomStepBuilder{
		step:    step,
		builder: sb.builder,
	}
}

// Parallel creates a parallel step that executes multiple sub-steps
func (sb *StepBuilder) Parallel() *ParallelStepBuilder {
	step := &ParallelStep{
		name:       sb.stepName,
		subSteps:   make([]FlowStep, 0),
		config:     sb.stepConfig,
		failFast:   true,
		maxConcurrency: 5,
	}

	sb.step = step
	return &ParallelStepBuilder{
		step:    step,
		builder: sb.builder,
	}
}

// Conditional creates a conditional step
func (sb *StepBuilder) Conditional(condition ConditionFunc) *ConditionalStepBuilder {
	step := &ConditionalStep{
		name:      sb.stepName,
		condition: condition,
		trueSteps: make([]FlowStep, 0),
		falseSteps: make([]FlowStep, 0),
		config:    sb.stepConfig,
	}

	sb.step = step
	return &ConditionalStepBuilder{
		step:    step,
		builder: sb.builder,
	}
}

// Timeout sets timeout for the step
func (sb *StepBuilder) Timeout(timeout time.Duration) *StepBuilder {
	if sb.stepConfig == nil {
		sb.stepConfig = &StepConfig{}
	}
	sb.stepConfig.Timeout = timeout
	return sb
}

// CanFail sets whether the step can fail without stopping the flow
func (sb *StepBuilder) CanFail(canFail bool) *StepBuilder {
	if sb.stepConfig == nil {
		sb.stepConfig = &StepConfig{}
	}
	sb.stepConfig.CanFail = canFail
	return sb
}

// Requires adds context key requirements for this step
func (sb *StepBuilder) Requires(keys ...string) *StepBuilder {
	if sb.stepConfig == nil {
		sb.stepConfig = &StepConfig{}
	}
	sb.stepConfig.Conditions = append(sb.stepConfig.Conditions, keys...)
	return sb
}

// Provides specifies context keys that this step provides
func (sb *StepBuilder) Provides(keys ...string) *StepBuilder {
	if sb.stepConfig == nil {
		sb.stepConfig = &StepConfig{}
	}
	sb.stepConfig.Outputs = append(sb.stepConfig.Outputs, keys...)
	return sb
}

// Continue adds the step to the flow and returns to flow builder
func (sb *StepBuilder) Continue() *FlowBuilder {
	if sb.step == nil {
		sb.builder.logger.Warn("Step not properly configured", "step", sb.stepName)
		return sb.builder
	}

	sb.builder.steps = append(sb.builder.steps, sb.step)
	return sb.builder
}

// ToolStepBuilder provides fluent API for tool steps
type ToolStepBuilder struct {
	step    *ToolStep
	builder *FlowBuilder
}

func (tsb *ToolStepBuilder) Output(key string) *ToolStepBuilder {
	tsb.step.outputKey = key
	return tsb
}

func (tsb *ToolStepBuilder) Input(key string) *ToolStepBuilder {
	tsb.step.inputKey = key
	return tsb
}

func (tsb *ToolStepBuilder) Timeout(timeout time.Duration) *ToolStepBuilder {
	if tsb.step.config == nil {
		tsb.step.config = &StepConfig{}
	}
	tsb.step.config.Timeout = timeout
	return tsb
}

func (tsb *ToolStepBuilder) CanFail(canFail bool) *ToolStepBuilder {
	if tsb.step.config == nil {
		tsb.step.config = &StepConfig{}
	}
	tsb.step.config.CanFail = canFail
	return tsb
}

func (tsb *ToolStepBuilder) Continue() *FlowBuilder {
	tsb.builder.steps = append(tsb.builder.steps, tsb.step)
	return tsb.builder
}

// ModelStepBuilder provides fluent API for model steps
type ModelStepBuilder struct {
	step    *ModelStep
	builder *FlowBuilder
}

func (msb *ModelStepBuilder) Input(key string) *ModelStepBuilder {
	msb.step.inputKey = key
	return msb
}

func (msb *ModelStepBuilder) Output(key string) *ModelStepBuilder {
	msb.step.outputKey = key
	return msb
}

func (msb *ModelStepBuilder) Inputs(keys ...string) *ModelStepBuilder {
	msb.step.inputKeys = keys
	return msb
}

func (msb *ModelStepBuilder) Temperature(temp float32) *ModelStepBuilder {
	msb.step.temperature = &temp
	return msb
}

func (msb *ModelStepBuilder) MaxTokens(tokens int) *ModelStepBuilder {
	msb.step.maxTokens = &tokens
	return msb
}

func (msb *ModelStepBuilder) WithMedia(mediaKeys ...string) *ModelStepBuilder {
	msb.step.mediaKeys = mediaKeys
	return msb
}

func (msb *ModelStepBuilder) Continue() *FlowBuilder {
	msb.builder.steps = append(msb.builder.steps, msb.step)
	return msb.builder
}

// CustomStepBuilder provides fluent API for custom steps
type CustomStepBuilder struct {
	step    *CustomStep
	builder *FlowBuilder
}

func (csb *CustomStepBuilder) Timeout(timeout time.Duration) *CustomStepBuilder {
	if csb.step.config == nil {
		csb.step.config = &StepConfig{}
	}
	csb.step.config.Timeout = timeout
	return csb
}

func (csb *CustomStepBuilder) Continue() *FlowBuilder {
	csb.builder.steps = append(csb.builder.steps, csb.step)
	return csb.builder
}

// ParallelStepBuilder provides fluent API for parallel steps
type ParallelStepBuilder struct {
	step    *ParallelStep
	builder *FlowBuilder
}

func (psb *ParallelStepBuilder) AddSubStep(step FlowStep) *ParallelStepBuilder {
	psb.step.subSteps = append(psb.step.subSteps, step)
	return psb
}

func (psb *ParallelStepBuilder) MaxConcurrency(max int) *ParallelStepBuilder {
	psb.step.maxConcurrency = max
	return psb
}

func (psb *ParallelStepBuilder) FailFast(failFast bool) *ParallelStepBuilder {
	psb.step.failFast = failFast
	return psb
}

func (psb *ParallelStepBuilder) Continue() *FlowBuilder {
	psb.builder.steps = append(psb.builder.steps, psb.step)
	return psb.builder
}

// ConditionalStepBuilder provides fluent API for conditional steps
type ConditionalStepBuilder struct {
	step    *ConditionalStep
	builder *FlowBuilder
}

func (csb *ConditionalStepBuilder) True(step FlowStep) *ConditionalStepBuilder {
	csb.step.trueSteps = append(csb.step.trueSteps, step)
	return csb
}

func (csb *ConditionalStepBuilder) False(step FlowStep) *ConditionalStepBuilder {
	csb.step.falseSteps = append(csb.step.falseSteps, step)
	return csb
}

func (csb *ConditionalStepBuilder) Continue() *FlowBuilder {
	csb.builder.steps = append(csb.builder.steps, csb.step)
	return csb.builder
}

// FlowStep interface and implementations

type FlowStep interface {
	Name() string
	Description() string
	Execute(ctx context.Context, flowCtx interfaces.FlowContext) error
	CanFail() bool
	Dependencies() []string
	Timeout() int
	RetryCount() int
}

type StepExecutor func(ctx context.Context, flowCtx interfaces.FlowContext) error

type ConditionFunc func(flowCtx interfaces.FlowContext) bool

// Type definitions for step implementations
type ToolStep struct {
	name      string
	tool      interfaces.PonchoTool
	inputKey  string
	outputKey string
	config    *StepConfig
}

type ModelStep struct {
	name         string
	model        interfaces.PonchoModel
	promptKey    string
	inputKey     string
	outputKey    string
	inputKeys    []string
	mediaKeys    []string
	temperature  *float32
	maxTokens    *int
	hasPrompt    bool
	config       *StepConfig
}

type CustomStep struct {
	name     string
	executor StepExecutor
	config   *StepConfig
}

type ParallelStep struct {
	name           string
	subSteps       []FlowStep
	failFast       bool
	maxConcurrency int
	config         *StepConfig
}

type ConditionalStep struct {
	name       string
	condition  ConditionFunc
	trueSteps  []FlowStep
	falseSteps []FlowStep
	config     *StepConfig
}

// ToolStep methods
func (ts *ToolStep) Name() string {
	return ts.name
}

func (ts *ToolStep) Description() string {
	return fmt.Sprintf("Execute tool %s", ts.tool.Name())
}

func (ts *ToolStep) Execute(ctx context.Context, flowCtx interfaces.FlowContext) error {
	// Get input from context
	var input interface{}
	if ts.inputKey != "" {
		if val, has := flowCtx.Get(ts.inputKey); has {
			input = val
		}
	}

	// Execute tool
	output, err := ts.tool.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	// Store output
	if ts.outputKey != "" {
		if err := flowCtx.Set(ts.outputKey, output); err != nil {
			return fmt.Errorf("failed to store output: %w", err)
		}
	}

	return nil
}

func (ts *ToolStep) CanFail() bool {
	if ts.config != nil {
		return ts.config.CanFail
	}
	return false
}

func (ts *ToolStep) Dependencies() []string {
	return []string{}
}

func (ts *ToolStep) Timeout() int {
	if ts.config != nil {
		return int(ts.config.Timeout.Seconds())
	}
	return 30 // default 30 seconds
}

func (ts *ToolStep) RetryCount() int {
	if ts.config != nil {
		return ts.config.MaxRetries
	}
	return 3 // default 3 retries
}

// ModelStep methods
func (ms *ModelStep) Name() string {
	return ms.name
}

func (ms *ModelStep) Description() string {
	return fmt.Sprintf("Execute model %s", ms.model.Name())
}

func (ms *ModelStep) Execute(ctx context.Context, flowCtx interfaces.FlowContext) error {
	// Build request
	var content []*interfaces.PonchoContentPart

	// Add text content
	if ms.inputKey != "" {
		if val, has := flowCtx.Get(ms.inputKey); has {
			content = append(content, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeText,
				Text: fmt.Sprintf("%v", val),
			})
		}
	}

	// Add media content
	for _, key := range ms.mediaKeys {
		if mediaData, err := flowCtx.GetMedia(key); err == nil {
			content = append(content, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeMedia,
				Media: &interfaces.PonchoMediaPart{
					URL: mediaData.URL,
				},
			})
		}
	}

	// Create request
	req := &interfaces.PonchoModelRequest{
		Model:       ms.model.Name(),
		Messages:    []*interfaces.PonchoMessage{
			{
				Role:    interfaces.PonchoRoleUser,
				Content: content,
			},
		},
	}

	// Set optional parameters
	if ms.temperature != nil {
		req.Temperature = ms.temperature
	}
	if ms.maxTokens != nil {
		req.MaxTokens = ms.maxTokens
	}

	// Execute model
	resp, err := ms.model.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("model execution failed: %w", err)
	}

	// Store output
	if ms.outputKey != "" && resp.Message != nil {
		if len(resp.Message.Content) > 0 {
			if err := flowCtx.Set(ms.outputKey, resp.Message.Content[0].Text); err != nil {
				return fmt.Errorf("failed to store output: %w", err)
			}
		}
	}

	return nil
}

func (ms *ModelStep) CanFail() bool {
	if ms.config != nil {
		return ms.config.CanFail
	}
	return false
}

func (ms *ModelStep) Dependencies() []string {
	return []string{}
}

func (ms *ModelStep) Timeout() int {
	if ms.config != nil {
		return int(ms.config.Timeout.Seconds())
	}
	return 60 // default 60 seconds
}

func (ms *ModelStep) RetryCount() int {
	if ms.config != nil {
		return ms.config.MaxRetries
	}
	return 3 // default 3 retries
}

// CustomStep methods
func (cs *CustomStep) Name() string {
	return cs.name
}

func (cs *CustomStep) Description() string {
	return fmt.Sprintf("Execute custom step %s", cs.name)
}

func (cs *CustomStep) Execute(ctx context.Context, flowCtx interfaces.FlowContext) error {
	return cs.executor(ctx, flowCtx)
}

func (cs *CustomStep) CanFail() bool {
	if cs.config != nil {
		return cs.config.CanFail
	}
	return false
}

func (cs *CustomStep) Dependencies() []string {
	return []string{}
}

func (cs *CustomStep) Timeout() int {
	if cs.config != nil {
		return int(cs.config.Timeout.Seconds())
	}
	return 30
}

func (cs *CustomStep) RetryCount() int {
	if cs.config != nil {
		return cs.config.MaxRetries
	}
	return 3
}

// ParallelStep methods
func (ps *ParallelStep) Name() string {
	return ps.name
}

func (ps *ParallelStep) Description() string {
	return fmt.Sprintf("Execute %d sub-steps in parallel", len(ps.subSteps))
}

func (ps *ParallelStep) Execute(ctx context.Context, flowCtx interfaces.FlowContext) error {
	// Create sub-context for each step
	subCtxs := make([]interfaces.FlowContext, len(ps.subSteps))

	// Execute all steps
	var wg sync.WaitGroup
	errChan := make(chan error, len(ps.subSteps))

	for i, step := range ps.subSteps {
		wg.Add(1)
		go func(idx int, s FlowStep) {
			defer wg.Done()

			// Create sub-context
			subCtx := flowCtx.CreateChild()
			subCtxs[idx] = subCtx

			// Execute step
			if err := s.Execute(ctx, subCtx); err != nil {
				errChan <- fmt.Errorf("sub-step %s failed: %w", s.Name(), err)
				if ps.failFast {
					return
				}
			}
		}(i, step)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("parallel execution failed: %v", errors)
	}

	// Merge all sub-contexts back into main context
	for _, subCtx := range subCtxs {
		if err := flowCtx.Merge(subCtx); err != nil {
			return fmt.Errorf("failed to merge sub-context: %w", err)
		}
	}

	return nil
}

func (ps *ParallelStep) CanFail() bool {
	if ps.config != nil {
		return ps.config.CanFail
	}
	return false
}

func (ps *ParallelStep) Dependencies() []string {
	return []string{}
}

func (ps *ParallelStep) Timeout() int {
	if ps.config != nil {
		return int(ps.config.Timeout.Seconds())
	}
	return 120 // default 2 minutes
}

func (ps *ParallelStep) RetryCount() int {
	if ps.config != nil {
		return ps.config.MaxRetries
	}
	return 1 // parallel steps typically retry once
}

// ConditionalStep methods
func (cs *ConditionalStep) Name() string {
	return cs.name
}

func (cs *ConditionalStep) Description() string {
	return fmt.Sprintf("Conditional execution step")
}

func (cs *ConditionalStep) Execute(ctx context.Context, flowCtx interfaces.FlowContext) error {
	if cs.condition(flowCtx) {
		// Execute true steps
		for _, step := range cs.trueSteps {
			if err := step.Execute(ctx, flowCtx); err != nil {
				return err
			}
		}
	} else {
		// Execute false steps
		for _, step := range cs.falseSteps {
			if err := step.Execute(ctx, flowCtx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *ConditionalStep) CanFail() bool {
	if cs.config != nil {
		return cs.config.CanFail
	}
	return false
}

func (cs *ConditionalStep) Dependencies() []string {
	return []string{}
}

func (cs *ConditionalStep) Timeout() int {
	if cs.config != nil {
		return int(cs.config.Timeout.Seconds())
	}
	return 30
}

func (cs *ConditionalStep) RetryCount() int {
	if cs.config != nil {
		return cs.config.MaxRetries
	}
	return 3
}