package registry

import (
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoFlowRegistry is the implementation of the flow registry
type PonchoFlowRegistry struct {
	flows  map[string]interfaces.PonchoFlow
	mutex  sync.RWMutex
	logger interfaces.Logger
	// Map of category to flow names
	categories map[string][]string
}

// NewPonchoFlowRegistry creates a new flow registry
func NewPonchoFlowRegistry(logger interfaces.Logger) *PonchoFlowRegistry {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &PonchoFlowRegistry{
		flows:      make(map[string]interfaces.PonchoFlow),
		categories: make(map[string][]string),
		logger:     logger,
	}
}

// Register registers a flow with the registry
func (r *PonchoFlowRegistry) Register(name string, flow interfaces.PonchoFlow) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("flow name cannot be empty")
	}

	if flow == nil {
		return fmt.Errorf("flow cannot be nil")
	}

	if _, exists := r.flows[name]; exists {
		return fmt.Errorf("flow '%s' is already registered", name)
	}

	r.flows[name] = flow

	// Add to category map
	category := flow.Category()
	if category != "" {
		if _, exists := r.categories[category]; !exists {
			r.categories[category] = make([]string, 0)
		}
		r.categories[category] = append(r.categories[category], name)
	}

	r.logger.Info("Flow registered", "name", name, "category", category)

	return nil
}

// Get retrieves a flow from the registry
func (r *PonchoFlowRegistry) Get(name string) (interfaces.PonchoFlow, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	flow, exists := r.flows[name]
	if !exists {
		return nil, fmt.Errorf("flow '%s' not found in registry", name)
	}

	return flow, nil
}

// List returns a list of all registered flow names
func (r *PonchoFlowRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.flows))
	for name := range r.flows {
		names = append(names, name)
	}

	return names
}

// ListByCategory returns a list of flow names in a specific category
func (r *PonchoFlowRegistry) ListByCategory(category string) []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if category == "" {
		return r.List()
	}

	flows, exists := r.categories[category]
	if !exists {
		return []string{}
	}

	// Return a copy to prevent modification
	result := make([]string, len(flows))
	copy(result, flows)
	return result
}

// Unregister removes a flow from the registry
func (r *PonchoFlowRegistry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	flow, exists := r.flows[name]
	if !exists {
		return fmt.Errorf("flow '%s' not found in registry", name)
	}

	// Remove from category map
	category := flow.Category()
	if category != "" {
		if flows, exists := r.categories[category]; exists {
			for i, flowName := range flows {
				if flowName == name {
					// Remove from slice
					r.categories[category] = append(flows[:i], flows[i+1:]...)
					break
				}
			}
			// If category is empty, remove it
			if len(r.categories[category]) == 0 {
				delete(r.categories, category)
			}
		}
	}

	delete(r.flows, name)
	r.logger.Info("Flow unregistered", "name", name)

	return nil
}

// Clear removes all flows from the registry
func (r *PonchoFlowRegistry) Clear() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.flows = make(map[string]interfaces.PonchoFlow)
	r.categories = make(map[string][]string)
	r.logger.Info("Flow registry cleared")

	return nil
}

// Count returns the number of registered flows
func (r *PonchoFlowRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.flows)
}

// Has checks if a flow is registered
func (r *PonchoFlowRegistry) Has(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.flows[name]
	return exists
}

// GetCategories returns a list of all categories
func (r *PonchoFlowRegistry) GetCategories() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := make([]string, 0, len(r.categories))
	for category := range r.categories {
		categories = append(categories, category)
	}

	return categories
}

// GetByTag returns all flows that have a specific tag
func (r *PonchoFlowRegistry) GetByTag(tag string) []interfaces.PonchoFlow {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []interfaces.PonchoFlow
	for _, flow := range r.flows {
		for _, flowTag := range flow.Tags() {
			if flowTag == tag {
				result = append(result, flow)
				break
			}
		}
	}

	return result
}

// GetByDependency returns all flows that depend on a specific component
func (r *PonchoFlowRegistry) GetByDependency(dependency string) []interfaces.PonchoFlow {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []interfaces.PonchoFlow
	for _, flow := range r.flows {
		for _, dep := range flow.Dependencies() {
			if dep == dependency {
				result = append(result, flow)
				break
			}
		}
	}

	return result
}

// ValidateDependencies checks if all dependencies of a flow are registered
func (r *PonchoFlowRegistry) ValidateDependencies(flow interfaces.PonchoFlow, modelRegistry interfaces.PonchoModelRegistry, toolRegistry interfaces.PonchoToolRegistry) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, dep := range flow.Dependencies() {
		// Check if dependency is a flow
		if r.Has(dep) {
			continue
		}

		// Check if dependency is a model
		if modelRegistry != nil {
			model, err := modelRegistry.Get(dep)
			if err == nil && model != nil {
				continue
			}
		}

		// Check if dependency is a tool
		if toolRegistry != nil {
			tool, err := toolRegistry.Get(dep)
			if err == nil && tool != nil {
				continue
			}
		}

		return fmt.Errorf("dependency '%s' for flow '%s' not found in any registry", dep, flow.Name())
	}

	return nil
}

// SetLogger sets the logger for the registry
func (r *PonchoFlowRegistry) SetLogger(logger interfaces.Logger) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if logger != nil {
		r.logger = logger
	}
}
