package registry

import (
	"fmt"
	"sync"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PonchoToolRegistry is the implementation of the tool registry
type PonchoToolRegistry struct {
	tools  map[string]interfaces.PonchoTool
	mutex  sync.RWMutex
	logger interfaces.Logger
	// Map of category to tool names
	categories map[string][]string
}

// NewPonchoToolRegistry creates a new tool registry
func NewPonchoToolRegistry(logger interfaces.Logger) *PonchoToolRegistry {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &PonchoToolRegistry{
		tools:      make(map[string]interfaces.PonchoTool),
		categories: make(map[string][]string),
		logger:     logger,
	}
}

// Register registers a tool with the registry
func (r *PonchoToolRegistry) Register(name string, tool interfaces.PonchoTool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' is already registered", name)
	}

	r.tools[name] = tool

	// Add to category map
	category := tool.Category()
	if category != "" {
		if _, exists := r.categories[category]; !exists {
			r.categories[category] = make([]string, 0)
		}
		r.categories[category] = append(r.categories[category], name)
	}

	r.logger.Info("Tool registered", "name", name, "category", category)

	return nil
}

// Get retrieves a tool from the registry
func (r *PonchoToolRegistry) Get(name string) (interfaces.PonchoTool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found in registry", name)
	}

	return tool, nil
}

// List returns a list of all registered tool names
func (r *PonchoToolRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}

// ListByCategory returns a list of tool names in a specific category
func (r *PonchoToolRegistry) ListByCategory(category string) []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if category == "" {
		return r.List()
	}

	tools, exists := r.categories[category]
	if !exists {
		return []string{}
	}

	// Return a copy to prevent modification
	result := make([]string, len(tools))
	copy(result, tools)
	return result
}

// Unregister removes a tool from the registry
func (r *PonchoToolRegistry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tool, exists := r.tools[name]
	if !exists {
		return fmt.Errorf("tool '%s' not found in registry", name)
	}

	// Remove from category map
	category := tool.Category()
	if category != "" {
		if tools, exists := r.categories[category]; exists {
			for i, toolName := range tools {
				if toolName == name {
					// Remove from slice
					r.categories[category] = append(tools[:i], tools[i+1:]...)
					break
				}
			}
			// If category is empty, remove it
			if len(r.categories[category]) == 0 {
				delete(r.categories, category)
			}
		}
	}

	delete(r.tools, name)
	r.logger.Info("Tool unregistered", "name", name)

	return nil
}

// Clear removes all tools from the registry
func (r *PonchoToolRegistry) Clear() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tools = make(map[string]interfaces.PonchoTool)
	r.categories = make(map[string][]string)
	r.logger.Info("Tool registry cleared")

	return nil
}

// Count returns the number of registered tools
func (r *PonchoToolRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.tools)
}

// Has checks if a tool is registered
func (r *PonchoToolRegistry) Has(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.tools[name]
	return exists
}

// GetCategories returns a list of all categories
func (r *PonchoToolRegistry) GetCategories() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := make([]string, 0, len(r.categories))
	for category := range r.categories {
		categories = append(categories, category)
	}

	return categories
}

// GetByTag returns all tools that have a specific tag
func (r *PonchoToolRegistry) GetByTag(tag string) []interfaces.PonchoTool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []interfaces.PonchoTool
	for _, tool := range r.tools {
		for _, toolTag := range tool.Tags() {
			if toolTag == tag {
				result = append(result, tool)
				break
			}
		}
	}

	return result
}

// GetByDependency returns all tools that depend on a specific component
func (r *PonchoToolRegistry) GetByDependency(dependency string) []interfaces.PonchoTool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []interfaces.PonchoTool
	for _, tool := range r.tools {
		for _, dep := range tool.Dependencies() {
			if dep == dependency {
				result = append(result, tool)
				break
			}
		}
	}

	return result
}

// SetLogger sets the logger for the registry
func (r *PonchoToolRegistry) SetLogger(logger interfaces.Logger) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if logger != nil {
		r.logger = logger
	}
}
