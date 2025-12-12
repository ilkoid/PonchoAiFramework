package registry

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockFlow is a mock implementation of PonchoFlow for testing
type MockFlow struct {
	name         string
	description  string
	version      string
	category     string
	tags         []string
	dependencies []string
	inputSchema  map[string]interface{}
	outputSchema map[string]interface{}
}

func NewMockFlow(name, description, version, category string, tags []string, dependencies []string) *MockFlow {
	return &MockFlow{
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

// Implement PonchoFlow interface
func (f *MockFlow) Name() string                         { return f.name }
func (f *MockFlow) Description() string                  { return f.description }
func (f *MockFlow) Version() string                      { return f.version }
func (f *MockFlow) Category() string                     { return f.category }
func (f *MockFlow) Tags() []string                       { return f.tags }
func (f *MockFlow) Dependencies() []string               { return f.dependencies }
func (f *MockFlow) InputSchema() map[string]interface{}  { return f.inputSchema }
func (f *MockFlow) OutputSchema() map[string]interface{} { return f.outputSchema }
func (f *MockFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return map[string]interface{}{"result": "mock flow output"}, nil
}
func (f *MockFlow) ExecuteStreaming(ctx context.Context, input interface{}, callback interfaces.PonchoStreamCallback) error {
	chunk := &interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "mock streaming flow output",
				},
			},
		},
		Done:         true,
		FinishReason: interfaces.PonchoFinishReasonStop,
	}
	return callback(chunk)
}
func (f *MockFlow) Initialize(ctx context.Context, config map[string]interface{}) error { return nil }
func (f *MockFlow) Shutdown(ctx context.Context) error                                  { return nil }

func TestFlowRegistry_Register(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Test valid registration
	flow := NewMockFlow("test-flow", "Test flow", "1.0.0", "test", []string{"mock"}, []string{})

	err := registry.Register("test-flow", flow)
	if err != nil {
		t.Errorf("Failed to register valid flow: %v", err)
	}

	// Test duplicate registration
	err = registry.Register("test-flow", flow)
	if err == nil {
		t.Error("Expected error when registering duplicate flow, got nil")
	}

	// Test empty name
	err = registry.Register("", flow)
	if err == nil {
		t.Error("Expected error when registering flow with empty name, got nil")
	}

	// Test nil flow
	err = registry.Register("nil-flow", nil)
	if err == nil {
		t.Error("Expected error when registering nil flow, got nil")
	}
}

func TestFlowRegistry_Get(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})
	flow := NewMockFlow("test-flow", "Test flow", "1.0.0", "test", []string{"mock"}, []string{})

	// Register a flow
	err := registry.Register("test-flow", flow)
	if err != nil {
		t.Fatalf("Failed to register flow: %v", err)
	}

	// Test getting existing flow
	retrievedFlow, err := registry.Get("test-flow")
	if err != nil {
		t.Errorf("Failed to get existing flow: %v", err)
	}
	if retrievedFlow == nil {
		t.Error("Retrieved flow is nil")
	}

	// Test getting non-existent flow
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent flow, got nil")
	}
}

func TestFlowRegistry_List(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Empty registry
	flows := registry.List()
	if len(flows) != 0 {
		t.Errorf("Expected empty list, got %d flows", len(flows))
	}

	// Register flows
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "cat1", []string{"tag1"}, []string{})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "cat2", []string{"tag2"}, []string{})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)

	// List should contain both flows
	flows = registry.List()
	if len(flows) != 2 {
		t.Errorf("Expected 2 flows, got %d", len(flows))
	}

	// Check if both flows are in the list
	found1, found2 := false, false
	for _, name := range flows {
		if name == "flow1" {
			found1 = true
		}
		if name == "flow2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("List does not contain all registered flows: %v", flows)
	}
}

func TestFlowRegistry_ListByCategory(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Register flows with different categories
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "category1", []string{}, []string{})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "category1", []string{}, []string{})
	flow3 := NewMockFlow("flow3", "Flow 3", "1.0.0", "category2", []string{}, []string{})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)
	registry.Register("flow3", flow3)

	// Test getting flows by category
	category1Flows := registry.ListByCategory("category1")
	if len(category1Flows) != 2 {
		t.Errorf("Expected 2 flows for category1, got %d", len(category1Flows))
	}

	category2Flows := registry.ListByCategory("category2")
	if len(category2Flows) != 1 {
		t.Errorf("Expected 1 flow for category2, got %d", len(category2Flows))
	}

	nonExistentCategoryFlows := registry.ListByCategory("non-existent")
	if len(nonExistentCategoryFlows) != 0 {
		t.Errorf("Expected 0 flows for non-existent category, got %d", len(nonExistentCategoryFlows))
	}

	// Test empty category (should return all flows)
	allFlows := registry.ListByCategory("")
	if len(allFlows) != 3 {
		t.Errorf("Expected 3 flows for empty category, got %d", len(allFlows))
	}
}

func TestFlowRegistry_Unregister(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})
	flow := NewMockFlow("test-flow", "Test flow", "1.0.0", "test", []string{"mock"}, []string{})

	// Register a flow
	registry.Register("test-flow", flow)

	// Unregister existing flow
	err := registry.Unregister("test-flow")
	if err != nil {
		t.Errorf("Failed to unregister existing flow: %v", err)
	}

	// Verify flow is unregistered
	_, err = registry.Get("test-flow")
	if err == nil {
		t.Error("Expected error after unregistering flow, got nil")
	}

	// Unregister non-existent flow
	err = registry.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent flow, got nil")
	}
}

func TestFlowRegistry_Clear(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Register flows
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "cat1", []string{}, []string{})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "cat2", []string{}, []string{})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)

	// Clear registry
	err := registry.Clear()
	if err != nil {
		t.Errorf("Failed to clear registry: %v", err)
	}

	// Verify registry is empty
	flows := registry.List()
	if len(flows) != 0 {
		t.Errorf("Expected empty registry after clear, got %d flows", len(flows))
	}

	// Verify categories are also cleared
	categories := registry.GetCategories()
	if len(categories) != 0 {
		t.Errorf("Expected empty categories after clear, got %d", len(categories))
	}
}

func TestFlowRegistry_Count(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Empty registry
	count := registry.Count()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Register flows
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "cat1", []string{}, []string{})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "cat2", []string{}, []string{})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)

	// Count should be 2
	count = registry.Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Unregister a flow
	registry.Unregister("flow1")

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

func TestFlowRegistry_Has(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})
	flow := NewMockFlow("test-flow", "Test flow", "1.0.0", "test", []string{"mock"}, []string{})

	// Register a flow
	registry.Register("test-flow", flow)

	// Check if flow exists
	if !registry.Has("test-flow") {
		t.Error("Expected Has to return true for existing flow")
	}

	// Check if non-existent flow exists
	if registry.Has("non-existent") {
		t.Error("Expected Has to return false for non-existent flow")
	}

	// Unregister flow
	registry.Unregister("test-flow")

	// Check if unregistered flow exists
	if registry.Has("test-flow") {
		t.Error("Expected Has to return false for unregistered flow")
	}
}

func TestFlowRegistry_GetCategories(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Empty registry
	categories := registry.GetCategories()
	if len(categories) != 0 {
		t.Errorf("Expected empty categories list, got %d", len(categories))
	}

	// Register flows with different categories
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "cat1", []string{}, []string{})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "cat2", []string{}, []string{})
	flow3 := NewMockFlow("flow3", "Flow 3", "1.0.0", "cat1", []string{}, []string{})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)
	registry.Register("flow3", flow3)

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

func TestFlowRegistry_GetByTag(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Register flows with different tags
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "cat1", []string{"tag1", "tag2"}, []string{})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "cat2", []string{"tag1"}, []string{})
	flow3 := NewMockFlow("flow3", "Flow 3", "1.0.0", "cat1", []string{"tag3"}, []string{})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)
	registry.Register("flow3", flow3)

	// Get flows by tag
	tag1Flows := registry.GetByTag("tag1")
	if len(tag1Flows) != 2 {
		t.Errorf("Expected 2 flows with tag1, got %d", len(tag1Flows))
	}

	tag2Flows := registry.GetByTag("tag2")
	if len(tag2Flows) != 1 {
		t.Errorf("Expected 1 flow with tag2, got %d", len(tag2Flows))
	}

	tag3Flows := registry.GetByTag("tag3")
	if len(tag3Flows) != 1 {
		t.Errorf("Expected 1 flow with tag3, got %d", len(tag3Flows))
	}

	nonExistentTagFlows := registry.GetByTag("non-existent")
	if len(nonExistentTagFlows) != 0 {
		t.Errorf("Expected 0 flows with non-existent tag, got %d", len(nonExistentTagFlows))
	}
}

func TestFlowRegistry_GetByDependency(t *testing.T) {
	registry := NewPonchoFlowRegistry(&MockLogger{})

	// Register flows with different dependencies
	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "cat1", []string{}, []string{"dep1", "dep2"})
	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "cat2", []string{}, []string{"dep1"})
	flow3 := NewMockFlow("flow3", "Flow 3", "1.0.0", "cat1", []string{}, []string{"dep3"})

	registry.Register("flow1", flow1)
	registry.Register("flow2", flow2)
	registry.Register("flow3", flow3)

	// Get flows by dependency
	dep1Flows := registry.GetByDependency("dep1")
	if len(dep1Flows) != 2 {
		t.Errorf("Expected 2 flows with dep1, got %d", len(dep1Flows))
	}

	dep2Flows := registry.GetByDependency("dep2")
	if len(dep2Flows) != 1 {
		t.Errorf("Expected 1 flow with dep2, got %d", len(dep2Flows))
	}

	dep3Flows := registry.GetByDependency("dep3")
	if len(dep3Flows) != 1 {
		t.Errorf("Expected 1 flow with dep3, got %d", len(dep3Flows))
	}

	nonExistentDepFlows := registry.GetByDependency("non-existent")
	if len(nonExistentDepFlows) != 0 {
		t.Errorf("Expected 0 flows with non-existent dependency, got %d", len(nonExistentDepFlows))
	}
}

// TODO: Implement ValidateDependencies test - requires NewMockModel and NewMockTool
// func TestFlowRegistry_ValidateDependencies(t *testing.T) {
// 	registry := NewPonchoFlowRegistry(&MockLogger{})
// 	modelRegistry := NewPonchoModelRegistry(&MockLogger{})
// 	toolRegistry := NewPonchoToolRegistry(&MockLogger{})

// 	// Register some models and tools for dependency validation
// 	model := NewMockModel("test-model", "test-provider", interfaces.ModelCapabilities{})
// 	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test", []string{}, []string{})

// 	modelRegistry.Register("test-model", model)
// 	toolRegistry.Register("test-tool", tool)

// 	// Test flow with valid dependencies
// 	flow1 := NewMockFlow("flow1", "Flow 1", "1.0.0", "test", []string{}, []string{"test-model", "test-tool"})
// 	registry.Register("flow1", flow1)

// 	err := registry.ValidateDependencies(flow1, modelRegistry, toolRegistry)
// 	if err != nil {
// 		t.Errorf("Expected no validation error for flow with valid dependencies: %v", err)
// 	}

// 	// Test flow with missing dependency
// 	flow2 := NewMockFlow("flow2", "Flow 2", "1.0.0", "test", []string{}, []string{"missing-dependency"})
// 	registry.Register("flow2", flow2)

// 	err = registry.ValidateDependencies(flow2, modelRegistry, toolRegistry)
// 	if err == nil {
// 		t.Error("Expected validation error for flow with missing dependency, got nil")
// 	}

// 	// Test flow with mixed valid and invalid dependencies
// 	flow3 := NewMockFlow("flow3", "Flow 3", "1.0.0", "test", []string{}, []string{"test-model", "missing-dep"})
// 	registry.Register("flow3", flow3)

// 	err = registry.ValidateDependencies(flow3, modelRegistry, toolRegistry)
// 	if err == nil {
// 		t.Error("Expected validation error for flow with mixed dependencies, got nil")
// 	}

// 	// Test flow with dependency on another flow
// 	flow4 := NewMockFlow("flow4", "Flow 4", "1.0.0", "test", []string{}, []string{"flow1"})
// 	registry.Register("flow4", flow4)

// 	err = registry.ValidateDependencies(flow4, modelRegistry, toolRegistry)
// 	if err != nil {
// 		t.Errorf("Expected no validation error for flow with flow dependency: %v", err)
// 	}
// }
