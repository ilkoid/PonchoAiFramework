package core

// PonchoFrameworkImpl - Main framework implementation and central orchestrator
//
// This file contains the core implementation of the PonchoFramework, serving as the
// central orchestrator for the entire PonchoFramework system. It manages the
// primary registries for models, tools, and flows, providing unified access to all
// framework components.
//
// Key responsibilities:
// - Registry management for models, tools, and flows with thread-safe operations
// - Lifecycle management with Start/Stop methods for graceful initialization
// - Configuration management integration with dynamic loading and validation
// - Metrics collection and monitoring for all framework operations
// - Request orchestration for model generation, tool execution, and flow processing
// - Health monitoring and status reporting for system observability
//
// This implementation follows the PonchoFramework interface and provides the
// foundation for all AI operations within the framework, coordinating between
// different components while maintaining performance and reliability.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/config"
	"github.com/ilkoid/PonchoAiFramework/core/registry"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoFrameworkImpl is the main implementation of PonchoFramework
type PonchoFrameworkImpl struct {
	// Core registries
	modelRegistry interfaces.PonchoModelRegistry
	toolRegistry  interfaces.PonchoToolRegistry
	flowRegistry  interfaces.PonchoFlowRegistry

	// Configuration and state
	config           *interfaces.PonchoFrameworkConfig
	configManager    config.ConfigManager
	logger           interfaces.Logger
	serviceLocator   ServiceLocator

	// Runtime state
	started   bool
	mutex     sync.RWMutex
	startTime time.Time

	// Metrics
	metrics *interfaces.PonchoMetrics
}

// NewPonchoFramework creates a new PonchoFramework instance
func NewPonchoFramework(cfg *interfaces.PonchoFrameworkConfig, logger interfaces.Logger) *PonchoFrameworkImpl {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Create default config manager
	configManager := config.NewConfigManager(config.ConfigOptions{
		FilePaths: []string{"config.yaml"},
		Logger:    logger,
	})

	// Create service locator
	serviceLocator := NewServiceLocator(logger)

	return &PonchoFrameworkImpl{
		modelRegistry: registry.NewPonchoModelRegistry(logger),
		toolRegistry:  registry.NewPonchoToolRegistry(logger),
		flowRegistry:  registry.NewPonchoFlowRegistry(logger),
		config:        cfg,
		configManager:  configManager,
		logger:        logger,
		serviceLocator: serviceLocator,
		metrics: &interfaces.PonchoMetrics{
			GeneratedRequests: &interfaces.GenerationMetrics{
				ByModel: make(map[string]*interfaces.ModelMetrics),
			},
			ToolExecutions: &interfaces.ToolMetrics{
				ByTool: make(map[string]int64),
			},
			FlowExecutions: &interfaces.FlowMetrics{
				ByFlow: make(map[string]int64),
			},
			Errors: &interfaces.ErrorMetrics{
				ByType:       make(map[string]int64),
				ByComponent:  make(map[string]int64),
				RecentErrors: make([]*interfaces.ErrorInfo, 0),
			},
			System:    &interfaces.SystemMetrics{},
			Timestamp: time.Now(),
		},
	}
}

// Start initializes and starts the framework
func (pf *PonchoFrameworkImpl) Start(ctx context.Context) error {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	if pf.started {
		return fmt.Errorf("framework is already started")
	}

	pf.logger.Info("Starting PonchoFramework")

	// Initialize configuration if needed
	if pf.config == nil {
		pf.config = &interfaces.PonchoFrameworkConfig{}
	}

	// Initialize service locator
	if pf.serviceLocator != nil {
		pf.logger.Info("Initializing service locator")
		if err := pf.serviceLocator.Initialize(); err != nil {
			pf.logger.Error("Failed to initialize service locator", "error", err)
			return fmt.Errorf("failed to initialize service locator: %w", err)
		}
	}

	// Load configuration
	if pf.configManager != nil {
		// Inject model factory manager into config manager
		if cm, ok := pf.configManager.(*config.ConfigManagerImpl); ok {
			cm.SetModelFactoryManager(pf.serviceLocator.GetModelFactoryManager())
		}

		pf.logger.Info("Loading configuration")
		if err := pf.configManager.Load(); err != nil {
			pf.logger.Error("Failed to load configuration", "error", err)
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Register factories from configuration
		if err := pf.registerFactoriesFromConfig(); err != nil {
			pf.logger.Error("Failed to register factories from configuration", "error", err)
			return fmt.Errorf("failed to register factories: %w", err)
		}

		// Load and register models from configuration
		if err := pf.loadAndRegisterModels(ctx); err != nil {
			pf.logger.Error("Failed to load and register models", "error", err)
			return fmt.Errorf("failed to load and register models: %w", err)
		}

		// TODO: Load and register tools from configuration
		// TODO: Load and register flows from configuration
	}

	pf.started = true
	pf.startTime = time.Now()
	pf.logger.Info("PonchoFramework started successfully")

	return nil
}

// Stop gracefully stops the framework
func (pf *PonchoFrameworkImpl) Stop(ctx context.Context) error {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	if !pf.started {
		return fmt.Errorf("framework is not started")
	}

	pf.logger.Info("Stopping PonchoFramework")

	// Shutdown service locator
	if pf.serviceLocator != nil {
		pf.logger.Info("Shutting down service locator")
		if err := pf.serviceLocator.Shutdown(); err != nil {
			pf.logger.Error("Failed to shutdown service locator", "error", err)
			// Continue with shutdown even if service locator fails
		}
	}

	// TODO: Shutdown components gracefully
	// This will be implemented in future phases

	pf.started = false
	pf.logger.Info("PonchoFramework stopped successfully")

	return nil
}

// RegisterModel registers a model with the framework
func (pf *PonchoFrameworkImpl) RegisterModel(name string, model interfaces.PonchoModel) error {
	if err := pf.modelRegistry.Register(name, model); err != nil {
		pf.logger.Error("Failed to register model", "name", name, "error", err)
		return err
	}

	// Initialize the model with framework configuration
	if pf.config != nil && pf.config.Models != nil {
		if modelConfig, exists := pf.config.Models[name]; exists {
			if err := model.Initialize(context.Background(), map[string]interface{}{
				"provider":      modelConfig.Provider,
				"model_name":    modelConfig.ModelName,
				"api_key":       modelConfig.APIKey,
				"base_url":      modelConfig.BaseURL,
				"max_tokens":    modelConfig.MaxTokens,
				"temperature":   modelConfig.Temperature,
				"timeout":       modelConfig.Timeout,
				"supports":      modelConfig.Supports,
				"custom_params": modelConfig.CustomParams,
			}); err != nil {
				pf.logger.Error("Failed to initialize model", "name", name, "error", err)
				return fmt.Errorf("failed to initialize model '%s': %w", name, err)
			}
		}
	}

	pf.logger.Info("Model registered and initialized", "name", name)
	return nil
}

// RegisterTool registers a tool with the framework
func (pf *PonchoFrameworkImpl) RegisterTool(name string, tool interfaces.PonchoTool) error {
	if err := pf.toolRegistry.Register(name, tool); err != nil {
		pf.logger.Error("Failed to register tool", "name", name, "error", err)
		return err
	}

	// Initialize the tool with framework configuration
	if pf.config != nil && pf.config.Tools != nil {
		if toolConfig, exists := pf.config.Tools[name]; exists {
			if err := tool.Initialize(context.Background(), map[string]interface{}{
				"enabled":       toolConfig.Enabled,
				"timeout":       toolConfig.Timeout,
				"retry":         toolConfig.Retry,
				"cache":         toolConfig.Cache,
				"dependencies":  toolConfig.Dependencies,
				"custom_params": toolConfig.CustomParams,
			}); err != nil {
				pf.logger.Error("Failed to initialize tool", "name", name, "error", err)
				return fmt.Errorf("failed to initialize tool '%s': %w", name, err)
			}
		}
	}

	pf.logger.Info("Tool registered and initialized", "name", name)
	return nil
}

// RegisterFlow registers a flow with the framework
func (pf *PonchoFrameworkImpl) RegisterFlow(name string, flow interfaces.PonchoFlow) error {
	if err := pf.flowRegistry.Register(name, flow); err != nil {
		pf.logger.Error("Failed to register flow", "name", name, "error", err)
		return err
	}

	// Initialize the flow with framework configuration
	if pf.config != nil && pf.config.Flows != nil {
		if flowConfig, exists := pf.config.Flows[name]; exists {
			if err := flow.Initialize(context.Background(), map[string]interface{}{
				"enabled":       flowConfig.Enabled,
				"timeout":       flowConfig.Timeout,
				"parallel":      flowConfig.Parallel,
				"dependencies":  flowConfig.Dependencies,
				"custom_params": flowConfig.CustomParams,
			}); err != nil {
				pf.logger.Error("Failed to initialize flow", "name", name, "error", err)
				return fmt.Errorf("failed to initialize flow '%s': %w", name, err)
			}
		}
	}

	pf.logger.Info("Flow registered and initialized", "name", name)
	return nil
}

// Generate generates a response using a model
func (pf *PonchoFrameworkImpl) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if !pf.isStarted() {
		return nil, fmt.Errorf("framework is not started")
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		pf.recordGenerationMetrics(req.Model, duration, 0, true)
	}()

	pf.logger.Debug("Generating response", "model", req.Model)

	model, err := pf.modelRegistry.Get(req.Model)
	if err != nil {
		pf.recordError("model", "model_not_found")
		return nil, fmt.Errorf("model '%s' not found: %w", req.Model, err)
	}

	response, err := model.Generate(ctx, req)
	if err != nil {
		pf.recordError("model", "generation_failed")
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	pf.logger.Debug("Generation completed", "model", req.Model, "tokens", response.Usage.TotalTokens)
	return response, nil
}

// GenerateStreaming generates a streaming response using a model
func (pf *PonchoFrameworkImpl) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	if !pf.isStarted() {
		return fmt.Errorf("framework is not started")
	}

	pf.logger.Debug("Starting streaming generation", "model", req.Model)

	model, err := pf.modelRegistry.Get(req.Model)
	if err != nil {
		pf.recordError("model", "model_not_found")
		return fmt.Errorf("model '%s' not found: %w", req.Model, err)
	}

	if !model.SupportsStreaming() {
		pf.recordError("model", "streaming_not_supported")
		return fmt.Errorf("model '%s' does not support streaming", req.Model)
	}

	err = model.GenerateStreaming(ctx, req, callback)
	if err != nil {
		pf.recordError("model", "streaming_failed")
		return fmt.Errorf("streaming generation failed: %w", err)
	}

	pf.logger.Debug("Streaming generation completed", "model", req.Model)
	return nil
}

// ExecuteTool executes a tool
func (pf *PonchoFrameworkImpl) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error) {
	if !pf.isStarted() {
		return nil, fmt.Errorf("framework is not started")
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		pf.recordToolExecutionMetrics(toolName, duration, true)
	}()

	pf.logger.Debug("Executing tool", "name", toolName)

	tool, err := pf.toolRegistry.Get(toolName)
	if err != nil {
		pf.recordError("tool", "tool_not_found")
		return nil, fmt.Errorf("tool '%s' not found: %w", toolName, err)
	}

	// Validate input
	if err := tool.Validate(input); err != nil {
		pf.recordError("tool", "validation_failed")
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	result, err := tool.Execute(ctx, input)
	if err != nil {
		pf.recordError("tool", "execution_failed")
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	pf.logger.Debug("Tool execution completed", "name", toolName)
	return result, nil
}

// ExecuteFlow executes a flow
func (pf *PonchoFrameworkImpl) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error) {
	if !pf.isStarted() {
		return nil, fmt.Errorf("framework is not started")
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		pf.recordFlowExecutionMetrics(flowName, duration, true)
	}()

	pf.logger.Debug("Executing flow", "name", flowName)

	flow, err := pf.flowRegistry.Get(flowName)
	if err != nil {
		pf.recordError("flow", "flow_not_found")
		return nil, fmt.Errorf("flow '%s' not found: %w", flowName, err)
	}

	// Validate dependencies
	if err := pf.flowRegistry.ValidateDependencies(flow, pf.modelRegistry, pf.toolRegistry); err != nil {
		pf.recordError("flow", "dependency_validation_failed")
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	result, err := flow.Execute(ctx, input)
	if err != nil {
		pf.recordError("flow", "execution_failed")
		return nil, fmt.Errorf("flow execution failed: %w", err)
	}

	pf.logger.Debug("Flow execution completed", "name", flowName)
	return result, nil
}

// ExecuteFlowStreaming executes a flow with streaming
func (pf *PonchoFrameworkImpl) ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback interfaces.PonchoStreamCallback) error {
	if !pf.isStarted() {
		return fmt.Errorf("framework is not started")
	}

	pf.logger.Debug("Starting streaming flow execution", "name", flowName)

	flow, err := pf.flowRegistry.Get(flowName)
	if err != nil {
		pf.recordError("flow", "flow_not_found")
		return fmt.Errorf("flow '%s' not found: %w", flowName, err)
	}

	// Validate dependencies
	if err := pf.flowRegistry.ValidateDependencies(flow, pf.modelRegistry, pf.toolRegistry); err != nil {
		pf.recordError("flow", "dependency_validation_failed")
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	err = flow.ExecuteStreaming(ctx, input, callback)
	if err != nil {
		pf.recordError("flow", "streaming_execution_failed")
		return fmt.Errorf("streaming flow execution failed: %w", err)
	}

	pf.logger.Debug("Streaming flow execution completed", "name", flowName)
	return nil
}

// GetModelRegistry returns the model registry
func (pf *PonchoFrameworkImpl) GetModelRegistry() interfaces.PonchoModelRegistry {
	return pf.modelRegistry
}

// GetToolRegistry returns the tool registry
func (pf *PonchoFrameworkImpl) GetToolRegistry() interfaces.PonchoToolRegistry {
	return pf.toolRegistry
}

// GetFlowRegistry returns the flow registry
func (pf *PonchoFrameworkImpl) GetFlowRegistry() interfaces.PonchoFlowRegistry {
	return pf.flowRegistry
}

// GetConfig returns the current configuration
func (pf *PonchoFrameworkImpl) GetConfig() *interfaces.PonchoFrameworkConfig {
	return pf.config
}

// ReloadConfig reloads the framework configuration
func (pf *PonchoFrameworkImpl) ReloadConfig(ctx context.Context) error {
	pf.logger.Info("Reloading configuration")
	// TODO: Implement configuration reloading
	// This will be implemented in future phases
	return nil
}

// Health returns the health status of the framework
func (pf *PonchoFrameworkImpl) Health(ctx context.Context) (*interfaces.PonchoHealthStatus, error) {
	pf.mutex.RLock()
	defer pf.mutex.RUnlock()

	status := &interfaces.PonchoHealthStatus{
		Timestamp:  time.Now(),
		Version:    "1.0.0", // TODO: Get from build info
		Components: make(map[string]*interfaces.ComponentHealth),
		Uptime:     time.Since(pf.startTime),
	}

	// Check framework status
	if pf.started {
		status.Status = "healthy"
		status.Components["framework"] = &interfaces.ComponentHealth{
			Status:  "healthy",
			Message: "Framework is running",
		}
	} else {
		status.Status = "unhealthy"
		status.Components["framework"] = &interfaces.ComponentHealth{
			Status:  "unhealthy",
			Message: "Framework is not started",
		}
	}

	// TODO: Add health checks for other components
	// This will be implemented in future phases

	return status, nil
}

// Metrics returns framework metrics
func (pf *PonchoFrameworkImpl) Metrics(ctx context.Context) (*interfaces.PonchoMetrics, error) {
	pf.mutex.RLock()
	defer pf.mutex.RUnlock()

	// Update timestamp
	pf.metrics.Timestamp = time.Now()

	// TODO: Collect real-time metrics
	// This will be implemented in future phases

	return pf.metrics, nil
}

// Helper methods

func (pf *PonchoFrameworkImpl) isStarted() bool {
	pf.mutex.RLock()
	defer pf.mutex.RUnlock()
	return pf.started
}

func (pf *PonchoFrameworkImpl) recordGenerationMetrics(model string, duration int64, tokens int, success bool) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	metrics := pf.metrics.GeneratedRequests
	metrics.TotalRequests++

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	// Update model-specific metrics
	if metrics.ByModel[model] == nil {
		metrics.ByModel[model] = &interfaces.ModelMetrics{}
	}

	modelMetrics := metrics.ByModel[model]
	modelMetrics.Requests++

	// Calculate success rate properly for both success and failure cases
	previousSuccesses := int64(modelMetrics.SuccessRate * float64(modelMetrics.Requests-1))
	if success {
		previousSuccesses++
	}
	modelMetrics.SuccessRate = float64(previousSuccesses) / float64(modelMetrics.Requests)
	
	modelMetrics.TotalTokens += int64(tokens)
	
	// Update average latency (simplified)
	modelMetrics.AvgLatency = (modelMetrics.AvgLatency + float64(duration)) / 2.0
}

func (pf *PonchoFrameworkImpl) recordToolExecutionMetrics(tool string, duration int64, success bool) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	metrics := pf.metrics.ToolExecutions
	metrics.TotalExecutions++

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	// Update tool-specific count
	metrics.ByTool[tool]++
}

func (pf *PonchoFrameworkImpl) recordFlowExecutionMetrics(flow string, duration int64, success bool) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	metrics := pf.metrics.FlowExecutions
	metrics.TotalExecutions++

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	// Update flow-specific count
	metrics.ByFlow[flow]++
}

func (pf *PonchoFrameworkImpl) recordError(component, errorType string) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	metrics := pf.metrics.Errors
	metrics.TotalErrors++

	metrics.ByType[errorType]++
	metrics.ByComponent[component]++

	// Add to recent errors (keep last 10)
	errorInfo := &interfaces.ErrorInfo{
		Timestamp: time.Now(),
		Component: component,
		Type:      errorType,
	}

	metrics.RecentErrors = append([]*interfaces.ErrorInfo{errorInfo}, metrics.RecentErrors...)
	if len(metrics.RecentErrors) > 10 {
		metrics.RecentErrors = metrics.RecentErrors[:10]
	}
}

// registerFactoriesFromConfig registers factories based on configuration
func (pf *PonchoFrameworkImpl) registerFactoriesFromConfig() error {
	if pf.serviceLocator == nil {
		pf.logger.Info("No service locator available, skipping factory registration")
		return nil
	}

	pf.logger.Info("Registering factories")

	// Import and register model factories
	// Note: These imports maintain clean architecture as the core only knows about interfaces
	// The actual factory implementations are registered by the framework initialization

	// Register model factories
	// DeepSeek factory
	// Note: In a production environment, these would be loaded dynamically based on config
	// For now, we'll use reflection or a registration mechanism

	// TODO: Load factory types from configuration and register them dynamically
	// This would involve reading config to know which factories to register
	// and then instantiating them without direct imports in the core

	return nil
}

// loadAndRegisterModels loads and registers models from configuration
func (pf *PonchoFrameworkImpl) loadAndRegisterModels(ctx context.Context) error {
	if pf.configManager == nil {
		pf.logger.Info("No config manager available, skipping model loading")
		return nil
	}

	pf.logger.Info("Loading models from configuration")

	// Get model configurations from config manager
	modelConfigs, err := pf.configManager.GetModelConfigs()
	if err != nil {
		return fmt.Errorf("failed to get model configurations: %w", err)
	}

	if len(modelConfigs) == 0 {
		pf.logger.Info("No model configurations found")
		return nil
	}

	// Load and initialize models using config manager
	models, err := pf.configManager.LoadAndInitializeModels()
	if err != nil {
		return fmt.Errorf("failed to load and initialize models: %w", err)
	}

	// Register models with framework
	for name, model := range models {
		if err := pf.modelRegistry.Register(name, model); err != nil {
			pf.logger.Error("Failed to register model", "name", name, "error", err)
			return fmt.Errorf("failed to register model %s: %w", name, err)
		}

		pf.logger.Info("Model registered successfully",
			"name", name,
			"provider", model.Provider(),
			"model_name", model.Name())
	}

	pf.logger.Info("Models loaded and registered successfully", "count", len(models))
	return nil
}

// reloadModels reloads models from configuration
func (pf *PonchoFrameworkImpl) reloadModels(ctx context.Context) error {
	pf.logger.Info("Reloading models")

	// Clear existing models
	if err := pf.modelRegistry.Clear(); err != nil {
		return fmt.Errorf("failed to clear model registry: %w", err)
	}

	// Load and register models again
	return pf.loadAndRegisterModels(ctx)
}

// GetConfigManager returns the config manager
func (pf *PonchoFrameworkImpl) GetConfigManager() config.ConfigManager {
	return pf.configManager
}

// GetModelFactoryManager returns the model factory manager
func (pf *PonchoFrameworkImpl) GetModelFactoryManager() interfaces.ModelFactoryManager {
	if pf.serviceLocator == nil {
		return nil
	}

	return pf.serviceLocator.GetModelFactoryManager()
}

// GetServiceLocator returns the service locator
func (pf *PonchoFrameworkImpl) GetServiceLocator() ServiceLocator {
	return pf.serviceLocator
}
