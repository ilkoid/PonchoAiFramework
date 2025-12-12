package registry

import (
	"context"
	"testing"
)

// MockTool is a mock implementation of PonchoTool for testing
type MockTool struct {
	name         string
	description  string
	version      string
	category     string
	tags         []string
	dependencies []string
	inputSchema  map[string]interface{}
	outputSchema map[string]interface{}
}

func NewMockTool(name, description, version, category string, tags []string, dependencies []string) *MockTool {
	return &MockTool{
		name:         name,
		description:  description,
		version:      version,
		category:     category,
		tags:         tags,
		dependencies: dependencies,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type": "string",
				},
			},
		},
		outputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"result": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}
}

// Implement PonchoTool interface
func (t *MockTool) Name() string                         { return t.name }
func (t *MockTool) Description() string                  { return t.description }
func (t *MockTool) Version() string                      { return t.version }
func (t *MockTool) Category() string                     { return t.category }
func (t *MockTool) Tags() []string                       { return t.tags }
func (t *MockTool) Dependencies() []string               { return t.dependencies }
func (t *MockTool) InputSchema() map[string]interface{}  { return t.inputSchema }
func (t *MockTool) OutputSchema() map[string]interface{} { return t.outputSchema }
func (t *MockTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return map[string]interface{}{"result": "mock output"}, nil
}
func (t *MockTool) Validate(input interface{}) error                                    { return nil }
func (t *MockTool) Initialize(ctx context.Context, config map[string]interface{}) error { return nil }
func (t *MockTool) Shutdown(ctx context.Context) error                                  { return nil }

func TestToolRegistry_Register(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Test valid registration
	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test", []string{"mock"}, []string{})

	err := registry.Register("test-tool", tool)
	if err != nil {
		t.Errorf("Failed to register valid tool: %v", err)
	}

	// Test duplicate registration
	err = registry.Register("test-tool", tool)
	if err == nil {
		t.Error("Expected error when registering duplicate tool, got nil")
	}

	// Test empty name
	err = registry.Register("", tool)
	if err == nil {
		t.Error("Expected error when registering tool with empty name, got nil")
	}

	// Test nil tool
	err = registry.Register("nil-tool", nil)
	if err == nil {
		t.Error("Expected error when registering nil tool, got nil")
	}
}

func TestToolRegistry_Get(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})
	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test", []string{"mock"}, []string{})

	// Register a tool
	err := registry.Register("test-tool", tool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Test getting existing tool
	retrievedTool, err := registry.Get("test-tool")
	if err != nil {
		t.Errorf("Failed to get existing tool: %v", err)
	}
	if retrievedTool == nil {
		t.Error("Retrieved tool is nil")
	}

	// Test getting non-existent tool
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent tool, got nil")
	}
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Empty registry
	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("Expected empty list, got %d tools", len(tools))
	}

	// Register tools
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "cat1", []string{"tag1"}, []string{})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "cat2", []string{"tag2"}, []string{})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)

	// List should contain both tools
	tools = registry.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Check if both tools are in the list
	found1, found2 := false, false
	for _, name := range tools {
		if name == "tool1" {
			found1 = true
		}
		if name == "tool2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("List does not contain all registered tools: %v", tools)
	}
}

func TestToolRegistry_ListByCategory(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Register tools with different categories
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "category1", []string{}, []string{})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "category1", []string{}, []string{})
	tool3 := NewMockTool("tool3", "Tool 3", "1.0.0", "category2", []string{}, []string{})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)
	registry.Register("tool3", tool3)

	// Test getting tools by category
	category1Tools := registry.ListByCategory("category1")
	if len(category1Tools) != 2 {
		t.Errorf("Expected 2 tools for category1, got %d", len(category1Tools))
	}

	category2Tools := registry.ListByCategory("category2")
	if len(category2Tools) != 1 {
		t.Errorf("Expected 1 tool for category2, got %d", len(category2Tools))
	}

	nonExistentCategoryTools := registry.ListByCategory("non-existent")
	if len(nonExistentCategoryTools) != 0 {
		t.Errorf("Expected 0 tools for non-existent category, got %d", len(nonExistentCategoryTools))
	}

	// Test empty category (should return all tools)
	allTools := registry.ListByCategory("")
	if len(allTools) != 3 {
		t.Errorf("Expected 3 tools for empty category, got %d", len(allTools))
	}
}

func TestToolRegistry_Unregister(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})
	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test", []string{"mock"}, []string{})

	// Register a tool
	registry.Register("test-tool", tool)

	// Unregister existing tool
	err := registry.Unregister("test-tool")
	if err != nil {
		t.Errorf("Failed to unregister existing tool: %v", err)
	}

	// Verify tool is unregistered
	_, err = registry.Get("test-tool")
	if err == nil {
		t.Error("Expected error after unregistering tool, got nil")
	}

	// Unregister non-existent tool
	err = registry.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent tool, got nil")
	}
}

func TestToolRegistry_Clear(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Register tools
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "cat1", []string{}, []string{})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "cat2", []string{}, []string{})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)

	// Clear registry
	err := registry.Clear()
	if err != nil {
		t.Errorf("Failed to clear registry: %v", err)
	}

	// Verify registry is empty
	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("Expected empty registry after clear, got %d tools", len(tools))
	}

	// Verify categories are also cleared
	categories := registry.GetCategories()
	if len(categories) != 0 {
		t.Errorf("Expected empty categories after clear, got %d", len(categories))
	}
}

func TestToolRegistry_Count(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Empty registry
	count := registry.Count()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Register tools
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "cat1", []string{}, []string{})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "cat2", []string{}, []string{})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)

	// Count should be 2
	count = registry.Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Unregister a tool
	registry.Unregister("tool1")

	// Count should be 1
	count = registry.Count()
	if count != 1 {
		t.Errorf("Expected count 1 after unregister, got %d", count)
	}

	// Clear registry
	registry.Clear()

	// Count should be 0
	count = registry.Count()
	if count != 0 {
		t.Errorf("Expected count 0 after clear, got %d", count)
	}
}

func TestToolRegistry_Has(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})
	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test", []string{"mock"}, []string{})

	// Register a tool
	registry.Register("test-tool", tool)

	// Check if tool exists
	if !registry.Has("test-tool") {
		t.Error("Expected Has to return true for existing tool")
	}

	// Check if non-existent tool exists
	if registry.Has("non-existent") {
		t.Error("Expected Has to return false for non-existent tool")
	}

	// Unregister tool
	registry.Unregister("test-tool")

	// Check if unregistered tool exists
	if registry.Has("test-tool") {
		t.Error("Expected Has to return false for unregistered tool")
	}
}

func TestToolRegistry_GetCategories(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Empty registry
	categories := registry.GetCategories()
	if len(categories) != 0 {
		t.Errorf("Expected empty categories list, got %d", len(categories))
	}

	// Register tools with different categories
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "cat1", []string{}, []string{})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "cat2", []string{}, []string{})
	tool3 := NewMockTool("tool3", "Tool 3", "1.0.0", "cat1", []string{}, []string{})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)
	registry.Register("tool3", tool3)

	// Should have 2 categories
	categories = registry.GetCategories()
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}

	// Check if both categories are present
	found1, found2 := false, false
	for _, cat := range categories {
		if cat == "cat1" {
			found1 = true
		}
		if cat == "cat2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("Categories list does not contain all categories: %v", categories)
	}
}

func TestToolRegistry_GetByTag(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Register tools with different tags
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "cat1", []string{"tag1", "tag2"}, []string{})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "cat2", []string{"tag1"}, []string{})
	tool3 := NewMockTool("tool3", "Tool 3", "1.0.0", "cat1", []string{"tag3"}, []string{})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)
	registry.Register("tool3", tool3)

	// Get tools by tag
	tag1Tools := registry.GetByTag("tag1")
	if len(tag1Tools) != 2 {
		t.Errorf("Expected 2 tools with tag1, got %d", len(tag1Tools))
	}

	tag2Tools := registry.GetByTag("tag2")
	if len(tag2Tools) != 1 {
		t.Errorf("Expected 1 tool with tag2, got %d", len(tag2Tools))
	}

	tag3Tools := registry.GetByTag("tag3")
	if len(tag3Tools) != 1 {
		t.Errorf("Expected 1 tool with tag3, got %d", len(tag3Tools))
	}

	nonExistentTagTools := registry.GetByTag("non-existent")
	if len(nonExistentTagTools) != 0 {
		t.Errorf("Expected 0 tools with non-existent tag, got %d", len(nonExistentTagTools))
	}
}

func TestToolRegistry_GetByDependency(t *testing.T) {
	registry := NewPonchoToolRegistry(&MockLogger{})

	// Register tools with different dependencies
	tool1 := NewMockTool("tool1", "Tool 1", "1.0.0", "cat1", []string{}, []string{"dep1", "dep2"})
	tool2 := NewMockTool("tool2", "Tool 2", "1.0.0", "cat2", []string{}, []string{"dep1"})
	tool3 := NewMockTool("tool3", "Tool 3", "1.0.0", "cat1", []string{}, []string{"dep3"})

	registry.Register("tool1", tool1)
	registry.Register("tool2", tool2)
	registry.Register("tool3", tool3)

	// Get tools by dependency
	dep1Tools := registry.GetByDependency("dep1")
	if len(dep1Tools) != 2 {
		t.Errorf("Expected 2 tools with dep1, got %d", len(dep1Tools))
	}

	dep2Tools := registry.GetByDependency("dep2")
	if len(dep2Tools) != 1 {
		t.Errorf("Expected 1 tool with dep2, got %d", len(dep2Tools))
	}

	dep3Tools := registry.GetByDependency("dep3")
	if len(dep3Tools) != 1 {
		t.Errorf("Expected 1 tool with dep3, got %d", len(dep3Tools))
	}

	nonExistentDepTools := registry.GetByDependency("non-existent")
	if len(nonExistentDepTools) != 0 {
		t.Errorf("Expected 0 tools with non-existent dependency, got %d", len(nonExistentDepTools))
	}
}
