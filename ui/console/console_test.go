package console

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockFramework implements PonchoFramework for testing
type MockFramework struct {
	models  map[string]interfaces.PonchoModel
	responses map[string]string
}

func NewMockFramework() *MockFramework {
	return &MockFramework{
		models: make(map[string]interfaces.PonchoModel),
		responses: make(map[string]string),
	}
}

func (m *MockFramework) Start(ctx context.Context) error {
	return nil
}

func (m *MockFramework) Stop(ctx context.Context) error {
	return nil
}

func (m *MockFramework) RegisterModel(name string, model interfaces.PonchoModel) error {
	m.models[name] = model
	return nil
}

func (m *MockFramework) RegisterTool(name string, tool interfaces.PonchoTool) error {
	return nil
}

func (m *MockFramework) RegisterFlow(name string, flow interfaces.PonchoFlow) error {
	return nil
}

func (m *MockFramework) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	response := "Mock response for: " + req.Model
	if m.responses[req.Model] != "" {
		response = m.responses[req.Model]
	}

	return &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: response,
				},
			},
		},
		FinishReason: interfaces.PonchoFinishReasonStop,
	}, nil
}

func (m *MockFramework) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	response := "Mock streaming response for: " + req.Model
	if m.responses[req.Model] != "" {
		response = m.responses[req.Model]
	}

	// Simulate streaming by sending chunks
	words := strings.Split(response, " ")
	for i, word := range words {
		chunk := &interfaces.PonchoStreamChunk{
			Delta: &interfaces.PonchoMessage{
				Role: interfaces.PonchoRoleAssistant,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: word,
					},
				},
			},
			Done: i == len(words)-1,
		}
		if i > 0 {
			chunk.Delta.Content[0].Text = " " + word
		}
		if err := callback(chunk); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond) // Simulate network delay
	}

	return nil
}

func (m *MockFramework) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockFramework) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockFramework) ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback interfaces.PonchoStreamCallback) error {
	return nil
}

func (m *MockFramework) GetModelRegistry() interfaces.PonchoModelRegistry {
	return &MockModelRegistry{models: m.models}
}

func (m *MockFramework) GetToolRegistry() interfaces.PonchoToolRegistry {
	return &MockToolRegistry{}
}

func (m *MockFramework) GetFlowRegistry() interfaces.PonchoFlowRegistry {
	return &MockFlowRegistry{}
}

func (m *MockFramework) GetConfig() *interfaces.PonchoFrameworkConfig {
	return &interfaces.PonchoFrameworkConfig{}
}

func (m *MockFramework) ReloadConfig(ctx context.Context) error {
	return nil
}

func (m *MockFramework) Health(ctx context.Context) (*interfaces.PonchoHealthStatus, error) {
	return &interfaces.PonchoHealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "test",
		Components: map[string]*interfaces.ComponentHealth{
			"test": {
				Status: "healthy",
			},
		},
	}, nil
}

func (m *MockFramework) Metrics(ctx context.Context) (*interfaces.PonchoMetrics, error) {
	return &interfaces.PonchoMetrics{}, nil
}

func (m *MockFramework) SetResponse(model, response string) {
	m.responses[model] = response
}

// MockModelRegistry implements PonchoModelRegistry for testing
type MockModelRegistry struct {
	models map[string]interfaces.PonchoModel
}

func (m *MockModelRegistry) Register(name string, model interfaces.PonchoModel) error {
	if m.models == nil {
		m.models = make(map[string]interfaces.PonchoModel)
	}
	m.models[name] = model
	return nil
}

func (m *MockModelRegistry) Get(name string) (interfaces.PonchoModel, error) {
	if m.models == nil {
		m.models = make(map[string]interfaces.PonchoModel)
	}
	return m.models[name], nil
}

func (m *MockModelRegistry) List() []string {
	if m.models == nil {
		return []string{}
	}
	names := make([]string, 0, len(m.models))
	for name := range m.models {
		names = append(names, name)
	}
	return names
}

func (m *MockModelRegistry) Unregister(name string) error {
	if m.models != nil {
		delete(m.models, name)
	}
	return nil
}

func (m *MockModelRegistry) Clear() error {
	m.models = make(map[string]interfaces.PonchoModel)
	return nil
}

// MockToolRegistry implements PonchoToolRegistry for testing
type MockToolRegistry struct{}

func (m *MockToolRegistry) Register(name string, tool interfaces.PonchoTool) error {
	return nil
}

func (m *MockToolRegistry) Get(name string) (interfaces.PonchoTool, error) {
	return nil, nil
}

func (m *MockToolRegistry) List() []string {
	return []string{}
}

func (m *MockToolRegistry) ListByCategory(category string) []string {
	return []string{}
}

func (m *MockToolRegistry) Unregister(name string) error {
	return nil
}

func (m *MockToolRegistry) Clear() error {
	return nil
}

func (m *MockToolRegistry) ValidateDependencies(flow interfaces.PonchoFlow, modelRegistry interfaces.PonchoModelRegistry, toolRegistry interfaces.PonchoToolRegistry) error {
	return nil
}

// MockFlowRegistry implements PonchoFlowRegistry for testing
type MockFlowRegistry struct{}

func (m *MockFlowRegistry) Register(name string, flow interfaces.PonchoFlow) error {
	return nil
}

func (m *MockFlowRegistry) Get(name string) (interfaces.PonchoFlow, error) {
	return nil, nil
}

func (m *MockFlowRegistry) List() []string {
	return []string{}
}

func (m *MockFlowRegistry) ListByCategory(category string) []string {
	return []string{}
}

func (m *MockFlowRegistry) Unregister(name string) error {
	return nil
}

func (m *MockFlowRegistry) Clear() error {
	return nil
}

func (m *MockFlowRegistry) ValidateDependencies(flow interfaces.PonchoFlow, modelRegistry interfaces.PonchoModelRegistry, toolRegistry interfaces.PonchoToolRegistry) error {
	return nil
}

// Test SimpleConsoleUI creation
func TestNewSimpleConsoleUI(t *testing.T) {
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}
	framework := NewMockFramework()
	logger := interfaces.NewNoOpLogger()

	ui := NewSimpleConsoleUI(input, output, framework, nil, logger)

	if ui.In != input {
		t.Errorf("Expected input to be set correctly")
	}

	if ui.Out != output {
		t.Errorf("Expected output to be set correctly")
	}

	if ui.Framework != framework {
		t.Errorf("Expected framework to be set correctly")
	}

	if ui.Logger != logger {
		t.Errorf("Expected logger to be set correctly")
	}
}

// Test FlowObserver implementation
func TestFlowObserver(t *testing.T) {
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}
	framework := NewMockFramework()
	logger := interfaces.NewNoOpLogger()

	ui := NewSimpleConsoleUI(input, output, framework, nil, logger)

	// Test OnEvent
	event := FlowEvent{
		Time:   time.Now(),
		Step:   "test_step",
		Status: "completed",
		Detail: "Test detail",
	}

	ui.OnEvent(event)

	// Check that event was stored
	if ui.lastFlowEvent == nil {
		t.Fatal("Expected lastFlowEvent to be set")
	}

	if ui.lastFlowEvent.Step != "test_step" {
		t.Errorf("Expected step to be 'test_step', got %s", ui.lastFlowEvent.Step)
	}

	// Check output
	outputStr := output.String()
	if !strings.Contains(outputStr, "test_step") {
		t.Errorf("Expected output to contain 'test_step', got: %s", outputStr)
	}
}

// Test agent configuration
func TestDefaultAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()

	if config.ModelName != "deepseek-chat" {
		t.Errorf("Expected default model name to be 'deepseek-chat', got %s", config.ModelName)
	}

	if config.System == "" {
		t.Errorf("Expected system prompt to be set")
	}

	if *config.Temperature != 0.7 {
		t.Errorf("Expected temperature to be 0.7, got %f", *config.Temperature)
	}

	if config.MaxTokens != 2000 {
		t.Errorf("Expected max tokens to be 2000, got %d", config.MaxTokens)
	}

	if !config.Stream {
		t.Errorf("Expected streaming to be enabled by default")
	}
}

// Test message building with system prompt
func TestBuildMessagesWithSystem(t *testing.T) {
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}
	framework := NewMockFramework()
	logger := interfaces.NewNoOpLogger()

	ui := NewSimpleConsoleUI(input, output, framework, nil, logger)

	// Test with empty history
	systemPrompt := "You are a helpful assistant"
	messages := ui.buildMessagesWithSystem(systemPrompt)

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != interfaces.PonchoRoleSystem {
		t.Errorf("Expected first message to be system message")
	}

	// Test with existing system message
	ui.chatHistory = []*interfaces.PonchoMessage{
		{
			Role: interfaces.PonchoRoleSystem,
			Content: []*interfaces.PonchoContentPart{
				{Type: interfaces.PonchoContentTypeText, Text: "Existing system"},
			},
		},
	}

	messages = ui.buildMessagesWithSystem(systemPrompt)

	if len(messages) != 1 {
		t.Errorf("Expected 1 message when system already exists, got %d", len(messages))
	}

	// Test with user message
	ui.chatHistory = []*interfaces.PonchoMessage{
		{
			Role: interfaces.PonchoRoleUser,
			Content: []*interfaces.PonchoContentPart{
				{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
			},
		},
	}

	messages = ui.buildMessagesWithSystem(systemPrompt)

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages (system + user), got %d", len(messages))
	}

	if messages[0].Role != interfaces.PonchoRoleSystem {
		t.Errorf("Expected first message to be system")
	}

	if messages[1].Role != interfaces.PonchoRoleUser {
		t.Errorf("Expected second message to be user")
	}
}