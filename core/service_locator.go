package core

import (
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/core/registry"
)

// ServiceLocator defines the interface for locating and managing factories
type ServiceLocator interface {
	// Model factory management
	GetModelFactory(provider string) (interfaces.ModelFactory, error)
	RegisterModelFactory(provider string, factory interfaces.ModelFactory) error
	GetModelFactoryManager() interfaces.ModelFactoryManager

	// Tool factory management
	GetToolFactory(toolType string) (interfaces.ToolFactory, error)
	RegisterToolFactory(toolType string, factory interfaces.ToolFactory) error
	GetToolFactoryManager() interfaces.ToolFactoryManager

	// Lifecycle
	Initialize() error
	Shutdown() error
}

// ServiceLocatorImpl implements the ServiceLocator interface
type ServiceLocatorImpl struct {
	logger             interfaces.Logger
	modelFactoryMgr    interfaces.ModelFactoryManager
	toolFactoryMgr     interfaces.ToolFactoryManager
	modelFactories     map[string]interfaces.ModelFactory
	toolFactories      map[string]interfaces.ToolFactory
	mutex              sync.RWMutex
	initialized        bool
}

// NewServiceLocator creates a new service locator
func NewServiceLocator(logger interfaces.Logger) *ServiceLocatorImpl {
	return &ServiceLocatorImpl{
		logger:         logger,
		modelFactories: make(map[string]interfaces.ModelFactory),
		toolFactories:  make(map[string]interfaces.ToolFactory),
	}
}

// Initialize initializes the service locator
func (sl *ServiceLocatorImpl) Initialize() error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	if sl.initialized {
		return nil
	}

	sl.logger.Info("Initializing service locator")

	// Initialize factory managers
	sl.modelFactoryMgr = registry.NewModelFactoryRegistry(sl.logger)
	sl.toolFactoryMgr = registry.NewToolFactoryRegistry(sl.logger)

	// Register default factories
	if err := sl.registerDefaultFactories(); err != nil {
		return fmt.Errorf("failed to register default factories: %w", err)
	}

	sl.initialized = true
	sl.logger.Info("Service locator initialized successfully")

	return nil
}

// Shutdown shuts down the service locator
func (sl *ServiceLocatorImpl) Shutdown() error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	if !sl.initialized {
		return nil
	}

	sl.logger.Info("Shutting down service locator")

	// Clear factories
	sl.modelFactories = make(map[string]interfaces.ModelFactory)
	sl.toolFactories = make(map[string]interfaces.ToolFactory)

	sl.initialized = false
	sl.logger.Info("Service locator shut down successfully")

	return nil
}

// GetModelFactory returns a model factory for the specified provider
func (sl *ServiceLocatorImpl) GetModelFactory(provider string) (interfaces.ModelFactory, error) {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	if !sl.initialized {
		return nil, fmt.Errorf("service locator not initialized")
	}

	return sl.modelFactoryMgr.GetFactory(provider)
}

// RegisterModelFactory registers a model factory
func (sl *ServiceLocatorImpl) RegisterModelFactory(provider string, factory interfaces.ModelFactory) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	if !sl.initialized {
		return fmt.Errorf("service locator not initialized")
	}

	if err := sl.modelFactoryMgr.RegisterFactory(provider, factory); err != nil {
		return fmt.Errorf("failed to register model factory for provider %s: %w", provider, err)
	}

	sl.logger.Debug("Model factory registered", "provider", provider)
	return nil
}

// GetModelFactoryManager returns the model factory manager
func (sl *ServiceLocatorImpl) GetModelFactoryManager() interfaces.ModelFactoryManager {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	return sl.modelFactoryMgr
}

// GetToolFactory returns a tool factory for the specified type
func (sl *ServiceLocatorImpl) GetToolFactory(toolType string) (interfaces.ToolFactory, error) {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	if !sl.initialized {
		return nil, fmt.Errorf("service locator not initialized")
	}

	return sl.toolFactoryMgr.GetFactory(toolType)
}

// RegisterToolFactory registers a tool factory
func (sl *ServiceLocatorImpl) RegisterToolFactory(toolType string, factory interfaces.ToolFactory) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	if !sl.initialized {
		return fmt.Errorf("service locator not initialized")
	}

	if err := sl.toolFactoryMgr.RegisterFactory(toolType, factory); err != nil {
		return fmt.Errorf("failed to register tool factory for type %s: %w", toolType, err)
	}

	sl.logger.Debug("Tool factory registered", "tool_type", toolType)
	return nil
}

// GetToolFactoryManager returns the tool factory manager
func (sl *ServiceLocatorImpl) GetToolFactoryManager() interfaces.ToolFactoryManager {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	return sl.toolFactoryMgr
}

// registerDefaultFactories registers default factories
func (sl *ServiceLocatorImpl) registerDefaultFactories() error {
	// Import and register model factories
	// This is done through registration, not direct imports to maintain clean architecture
	// The actual registration will be done by the framework initialization

	return nil
}