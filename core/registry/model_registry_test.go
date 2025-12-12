package registry

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockModel is a mock implementation of PonchoModel for testing
type MockModel struct {
	name               string
	provider           string
	maxTokens          int
	defaultTemperature float32
	capabilities       interfaces.ModelCapabilities
}

func NewMockModel(name, provider string, capabilities interfaces.ModelCapabilities) *MockModel {
	return &MockModel{
		name:               name,
		provider:           provider,
		maxTokens:          4000,
		defaultTemperature: 0.7,
		capabilities:       capabilities,
	}
}

// Implement PonchoModel interface
func (m *MockModel) Name() string                { return m.name }
func (m *MockModel) Provider() string            { return m.provider }
func (m *MockModel) MaxTokens() int              { return m.maxTokens }
func (m *MockModel) DefaultTemperature() float32 { return m.defaultTemperature }
func (m *MockModel) SupportsStreaming() bool     { return m.capabilities.Streaming }
func (m *MockModel) SupportsTools() bool         { return m.capabilities.Tools }
func (m *MockModel) SupportsVision() bool        { return m.capabilities.Vision }
func (m *MockModel) SupportsSystemRole() bool    { return m.capabilities.System }
func (m *MockModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	return &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Mock response",
				},
			},
		},
		Usage: &interfaces.PonchoUsage{
			PromptTokens:     10,
			CompletionTokens: 10,
			TotalTokens:      20,
		},
		FinishReason: interfaces.PonchoFinishReasonStop,
	}, nil
}
func (m *MockModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	chunk := &interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Mock streaming response",
				},
			},
		},
		Done:         true,
		FinishReason: interfaces.PonchoFinishReasonStop,
	}
	return callback(chunk)
}
func (m *MockModel) Initialize(ctx context.Context, config map[string]interface{}) error { return nil }
func (m *MockModel) Shutdown(ctx context.Context) error                                  { return nil }

// MockLogger is a mock implementation of Logger for testing
type MockLogger struct{}

func (l *MockLogger) Debug(msg string, fields ...interface{}) {}
func (l *MockLogger) Info(msg string, fields ...interface{})  {}
func (l *MockLogger) Warn(msg string, fields ...interface{})  {}
func (l *MockLogger) Error(msg string, fields ...interface{}) {}

func TestModelRegistry_Register(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})

	// Test valid registration
	model := NewMockModel("test-model", "test-provider", interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    false,
		System:    true,
	})

	err := registry.Register("test-model", model)
	if err != nil {
		t.Errorf("Failed to register valid model: %v", err)
	}

	// Test duplicate registration
	err = registry.Register("test-model", model)
	if err == nil {
		t.Error("Expected error when registering duplicate model, got nil")
	}

	// Test empty name
	err = registry.Register("", model)
	if err == nil {
		t.Error("Expected error when registering model with empty name, got nil")
	}

	// Test nil model
	err = registry.Register("nil-model", nil)
	if err == nil {
		t.Error("Expected error when registering nil model, got nil")
	}
}

func TestModelRegistry_Get(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})
	model := NewMockModel("test-model", "test-provider", interfaces.ModelCapabilities{})

	// Register a model
	err := registry.Register("test-model", model)
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	// Test getting existing model
	retrievedModel, err := registry.Get("test-model")
	if err != nil {
		t.Errorf("Failed to get existing model: %v", err)
	}
	if retrievedModel == nil {
		t.Error("Retrieved model is nil")
	}

	// Test getting non-existent model
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent model, got nil")
	}
}

func TestModelRegistry_List(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})

	// Empty registry
	models := registry.List()
	if len(models) != 0 {
		t.Errorf("Expected empty list, got %d models", len(models))
	}

	// Register models
	model1 := NewMockModel("model1", "provider1", interfaces.ModelCapabilities{})
	model2 := NewMockModel("model2", "provider2", interfaces.ModelCapabilities{})

	registry.Register("model1", model1)
	registry.Register("model2", model2)

	// List should contain both models
	models = registry.List()
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	// Check if both models are in the list
	found1, found2 := false, false
	for _, name := range models {
		if name == "model1" {
			found1 = true
		}
		if name == "model2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("List does not contain all registered models: %v", models)
	}
}

func TestModelRegistry_Unregister(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})
	model := NewMockModel("test-model", "test-provider", interfaces.ModelCapabilities{})

	// Register a model
	registry.Register("test-model", model)

	// Unregister existing model
	err := registry.Unregister("test-model")
	if err != nil {
		t.Errorf("Failed to unregister existing model: %v", err)
	}

	// Verify model is unregistered
	_, err = registry.Get("test-model")
	if err == nil {
		t.Error("Expected error after unregistering model, got nil")
	}

	// Unregister non-existent model
	err = registry.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent model, got nil")
	}
}

func TestModelRegistry_Clear(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})

	// Register models
	model1 := NewMockModel("model1", "provider1", interfaces.ModelCapabilities{})
	model2 := NewMockModel("model2", "provider2", interfaces.ModelCapabilities{})

	registry.Register("model1", model1)
	registry.Register("model2", model2)

	// Clear registry
	err := registry.Clear()
	if err != nil {
		t.Errorf("Failed to clear registry: %v", err)
	}

	// Verify registry is empty
	models := registry.List()
	if len(models) != 0 {
		t.Errorf("Expected empty registry after clear, got %d models", len(models))
	}
}

func TestModelRegistry_GetByProvider(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})

	// Register models with different providers
	model1 := NewMockModel("model1", "provider1", interfaces.ModelCapabilities{})
	model2 := NewMockModel("model2", "provider1", interfaces.ModelCapabilities{})
	model3 := NewMockModel("model3", "provider2", interfaces.ModelCapabilities{})

	registry.Register("model1", model1)
	registry.Register("model2", model2)
	registry.Register("model3", model3)

	// Get models by provider
	provider1Models := registry.GetByProvider("provider1")
	if len(provider1Models) != 2 {
		t.Errorf("Expected 2 models for provider1, got %d", len(provider1Models))
	}

	provider2Models := registry.GetByProvider("provider2")
	if len(provider2Models) != 1 {
		t.Errorf("Expected 1 model for provider2, got %d", len(provider2Models))
	}

	nonExistentProviderModels := registry.GetByProvider("non-existent")
	if len(nonExistentProviderModels) != 0 {
		t.Errorf("Expected 0 models for non-existent provider, got %d", len(nonExistentProviderModels))
	}
}

func TestModelRegistry_GetByCapability(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})

	// Register models with different capabilities
	model1 := NewMockModel("model1", "provider1", interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    true,
		System:    true,
	})
	model2 := NewMockModel("model2", "provider1", interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    false,
		System:    true,
	})
	model3 := NewMockModel("model3", "provider2", interfaces.ModelCapabilities{
		Streaming: false,
		Tools:     false,
		Vision:    false,
		System:    true,
	})

	registry.Register("model1", model1)
	registry.Register("model2", model2)
	registry.Register("model3", model3)

	// Get models by capability
	visionModels := registry.GetByCapability(true, false, false, false)
	if len(visionModels) != 1 {
		t.Errorf("Expected 1 model with vision capability, got %d", len(visionModels))
	}

	streamingAndToolsModels := registry.GetByCapability(false, true, true, false)
	if len(streamingAndToolsModels) != 2 {
		t.Errorf("Expected 2 models with streaming and tools capabilities, got %d", len(streamingAndToolsModels))
	}

	allCapabilitiesModels := registry.GetByCapability(true, true, true, true)
	if len(allCapabilitiesModels) != 1 {
		t.Errorf("Expected 1 model with all capabilities, got %d", len(allCapabilitiesModels))
	}
}

func TestModelRegistry_Count(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})

	// Empty registry
	count := registry.Count()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Register models
	model1 := NewMockModel("model1", "provider1", interfaces.ModelCapabilities{})
	model2 := NewMockModel("model2", "provider2", interfaces.ModelCapabilities{})

	registry.Register("model1", model1)
	registry.Register("model2", model2)

	// Count should be 2
	count = registry.Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Unregister a model
	registry.Unregister("model1")

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

func TestModelRegistry_Has(t *testing.T) {
	registry := NewPonchoModelRegistry(&MockLogger{})
	model := NewMockModel("test-model", "test-provider", interfaces.ModelCapabilities{})

	// Register a model
	registry.Register("test-model", model)

	// Check if model exists
	if !registry.Has("test-model") {
		t.Error("Expected Has to return true for existing model")
	}

	// Check if non-existent model exists
	if registry.Has("non-existent") {
		t.Error("Expected Has to return false for non-existent model")
	}

	// Unregister model
	registry.Unregister("test-model")

	// Check if unregistered model exists
	if registry.Has("test-model") {
		t.Error("Expected Has to return false for unregistered model")
	}
}
