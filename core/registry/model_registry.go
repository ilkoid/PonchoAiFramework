package registry

// This file implements the model registry for the PonchoFramework.
// It provides thread-safe storage and management of AI models.
// It implements the PonchoModelRegistry interface with full CRUD operations.
// It supports model registration, retrieval, listing, and removal.
// It provides concurrent access protection using sync.RWMutex.
// It includes validation for model registration and operations.
// It serves as the central repository for all model instances in the framework.
// It provides logging for all registry operations and error handling.

import (
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoModelRegistry is the implementation of the model registry
type PonchoModelRegistry struct {
	models map[string]interfaces.PonchoModel
	mutex  sync.RWMutex
	logger interfaces.Logger
}

// NewPonchoModelRegistry creates a new model registry
func NewPonchoModelRegistry(logger interfaces.Logger) *PonchoModelRegistry {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &PonchoModelRegistry{
		models: make(map[string]interfaces.PonchoModel),
		logger: logger,
	}
}

// Register registers a model with the registry
func (r *PonchoModelRegistry) Register(name string, model interfaces.PonchoModel) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}

	if _, exists := r.models[name]; exists {
		return fmt.Errorf("model '%s' is already registered", name)
	}

	r.models[name] = model
	r.logger.Info("Model registered", "name", name, "provider", model.Provider())

	return nil
}

// Get retrieves a model from the registry
func (r *PonchoModelRegistry) Get(name string) (interfaces.PonchoModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	model, exists := r.models[name]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in registry", name)
	}

	return model, nil
}

// List returns a list of all registered model names
func (r *PonchoModelRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.models))
	for name := range r.models {
		names = append(names, name)
	}

	return names
}

// Unregister removes a model from the registry
func (r *PonchoModelRegistry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.models[name]; !exists {
		return fmt.Errorf("model '%s' not found in registry", name)
	}

	delete(r.models, name)
	r.logger.Info("Model unregistered", "name", name)

	return nil
}

// Clear removes all models from the registry
func (r *PonchoModelRegistry) Clear() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.models = make(map[string]interfaces.PonchoModel)
	r.logger.Info("Model registry cleared")

	return nil
}

// Count returns the number of registered models
func (r *PonchoModelRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.models)
}

// Has checks if a model is registered
func (r *PonchoModelRegistry) Has(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.models[name]
	return exists
}

// GetByProvider returns all models from a specific provider
func (r *PonchoModelRegistry) GetByProvider(provider string) []interfaces.PonchoModel {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []interfaces.PonchoModel
	for _, model := range r.models {
		if model.Provider() == provider {
			result = append(result, model)
		}
	}

	return result
}

// GetByCapability returns all models that support a specific capability
func (r *PonchoModelRegistry) GetByCapability(supportsVision, supportsTools, supportsStreaming, supportsSystemRole bool) []interfaces.PonchoModel {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []interfaces.PonchoModel
	for _, model := range r.models {
		if (supportsVision && !model.SupportsVision()) ||
			(supportsTools && !model.SupportsTools()) ||
			(supportsStreaming && !model.SupportsStreaming()) ||
			(supportsSystemRole && !model.SupportsSystemRole()) {
			continue
		}
		result = append(result, model)
	}

	return result
}

// SetLogger sets the logger for the registry
func (r *PonchoModelRegistry) SetLogger(logger interfaces.Logger) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if logger != nil {
		r.logger = logger
	}
}
