package core

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Mock implementations for testing

type MockModel struct {
	*base.PonchoBaseModel
	generateFunc          func(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error)
	generateStreamingFunc func(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error
	supportsStreaming     bool
	supportsTools         bool
	supportsVision        bool
	supportsSystem        bool
}

func NewMockModel(name, provider string) *MockModel {
	baseModel := base.NewPonchoBaseModel(name, provider, interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    true,
		System:    true,
	})

	return &MockModel{
		PonchoBaseModel:   baseModel,
		supportsStreaming: true,
		supportsTools:     true,
		supportsVision:    true,
		supportsSystem:    true,
	}
}

func (m *MockModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}

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
			CompletionTokens: 5,
			TotalTokens:      15,
		},
		FinishReason: interfaces.PonchoFinishReasonStop,
	}, nil
}

func (m *MockModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	if m.generateStreamingFunc != nil {
		return m.generateStreamingFunc(ctx, req, callback)
	}

	// Simulate streaming response
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
		Done: true,
	}

	return callback(chunk)
}

func (m *MockModel) SupportsStreaming() bool {
	return m.supportsStreaming
}

func (m *MockModel) SupportsTools() bool {
	return m.supportsTools
}

func (m *MockModel) SupportsVision() bool {
	return m.supportsVision
}

func (m *MockModel) SupportsSystemRole() bool {
	return m.supportsSystem
}

type MockTool struct {
	*base.PonchoBaseTool
	executeFunc func(ctx context.Context, input interface{}) (interface{}, error)
}

func NewMockTool(name, description, version, category string) *MockTool {
	baseTool := base.NewPonchoBaseTool(name, description, version, category)

	return &MockTool{
		PonchoBaseTool: baseTool,
	}
}

func (t *MockTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if t.executeFunc != nil {
		return t.executeFunc(ctx, input)
	}

	return map[string]interface{}{
		"result": "mock tool result",
		"input":  input,
	}, nil
}

type MockFlow struct {
	*base.PonchoBaseFlow
	executeFunc func(ctx context.Context, input interface{}) (interface{}, error)
}

func NewMockFlow(name, description, version, category string) *MockFlow {
	baseFlow := base.NewPonchoBaseFlow(name, description, version, category)

	return &MockFlow{
		PonchoBaseFlow: baseFlow,
	}
}

func (f *MockFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if f.executeFunc != nil {
		return f.executeFunc(ctx, input)
	}

	return map[string]interface{}{
		"result": "mock flow result",
		"input":  input,
	}, nil
}

func (f *MockFlow) ExecuteStreaming(ctx context.Context, input interface{}, callback interfaces.PonchoStreamCallback) error {
	chunk := &interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "mock flow streaming response",
				},
			},
		},
		Done: true,
	}

	return callback(chunk)
}

// Test Framework creation and initialization

func TestNewPonchoFramework(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{
		Models: make(map[string]*interfaces.ModelConfig),
		Tools:  make(map[string]*interfaces.ToolConfig),
		Flows:  make(map[string]*interfaces.FlowConfig),
	}

	framework := NewPonchoFramework(config, logger)

	if framework == nil {
		t.Fatal("Expected framework to be created")
	}

	if framework.GetConfig() != config {
		t.Error("Expected config to be set")
	}

	if framework.GetModelRegistry() == nil {
		t.Error("Expected model registry to be initialized")
	}

	if framework.GetToolRegistry() == nil {
		t.Error("Expected tool registry to be initialized")
	}

	if framework.GetFlowRegistry() == nil {
		t.Error("Expected flow registry to be initialized")
	}
}

func TestNewPonchoFrameworkWithNilLogger(t *testing.T) {
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, nil)

	if framework == nil {
		t.Fatal("Expected framework to be created with default logger")
	}
}

func TestPonchoFrameworkStart(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// First start should succeed
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	// Second start should fail
	err = framework.Start(ctx)
	if err == nil {
		t.Error("Expected second start to fail")
	}

	// Stop the framework
	err = framework.Stop(ctx)
	if err != nil {
		t.Fatalf("Expected stop to succeed, got error: %v", err)
	}

	// Start again should succeed
	err = framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected restart to succeed, got error: %v", err)
	}
}

func TestPonchoFrameworkStop(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Stop without start should fail
	err := framework.Stop(ctx)
	if err == nil {
		t.Error("Expected stop without start to fail")
	}

	// Start and then stop should succeed
	err = framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	err = framework.Stop(ctx)
	if err != nil {
		t.Fatalf("Expected stop to succeed, got error: %v", err)
	}
}

// Test Model registration and generation

func TestRegisterModel(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"test-model": {
				Provider:    "test",
				ModelName:   "test-model",
				APIKey:      "test-key",
				MaxTokens:   1000,
				Temperature: 0.7,
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    true,
					System:    true,
				},
			},
		},
	}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	model := NewMockModel("test-model", "test")
	err = framework.RegisterModel("test-model", model)
	if err != nil {
		t.Fatalf("Expected model registration to succeed, got error: %v", err)
	}

	// Verify model is in registry
	retrievedModel, err := framework.GetModelRegistry().Get("test-model")
	if err != nil {
		t.Fatalf("Expected to retrieve model, got error: %v", err)
	}

	if retrievedModel.Name() != "test-model" {
		t.Error("Expected retrieved model to have correct name")
	}
}

func TestGenerate(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	model := NewMockModel("test-model", "test")
	err = framework.RegisterModel("test-model", model)
	if err != nil {
		t.Fatalf("Expected model registration to succeed, got error: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	response, err := framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected generation to succeed, got error: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	if response.Message == nil {
		t.Error("Expected message to be non-nil")
	}

	if response.Usage == nil {
		t.Error("Expected usage to be non-nil")
	}
}

func TestGenerateModelNotFound(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "non-existent-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	_, err = framework.Generate(ctx, req)
	if err == nil {
		t.Error("Expected generation to fail for non-existent model")
	}
}

func TestGenerateFrameworkNotStarted(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	_, err := framework.Generate(ctx, req)
	if err == nil {
		t.Error("Expected generation to fail when framework not started")
	}
}

func TestGenerateStreaming(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	model := NewMockModel("test-model", "test")
	err = framework.RegisterModel("test-model", model)
	if err != nil {
		t.Fatalf("Expected model registration to succeed, got error: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	var callbackCalled bool
	callback := func(chunk *interfaces.PonchoStreamChunk) error {
		callbackCalled = true
		return nil
	}

	err = framework.GenerateStreaming(ctx, req, callback)
	if err != nil {
		t.Fatalf("Expected streaming generation to succeed, got error: %v", err)
	}

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
}

// Test Tool registration and execution

func TestRegisterTool(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{
		Tools: map[string]*interfaces.ToolConfig{
			"test-tool": {
				Enabled: true,
				Timeout: "30s",
			},
		},
	}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test")
	err = framework.RegisterTool("test-tool", tool)
	if err != nil {
		t.Fatalf("Expected tool registration to succeed, got error: %v", err)
	}

	// Verify tool is in registry
	retrievedTool, err := framework.GetToolRegistry().Get("test-tool")
	if err != nil {
		t.Fatalf("Expected to retrieve tool, got error: %v", err)
	}

	if retrievedTool.Name() != "test-tool" {
		t.Error("Expected retrieved tool to have correct name")
	}
}

func TestExecuteTool(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test")
	err = framework.RegisterTool("test-tool", tool)
	if err != nil {
		t.Fatalf("Expected tool registration to succeed, got error: %v", err)
	}

	input := map[string]interface{}{
		"test_param": "test_value",
	}

	result, err := framework.ExecuteTool(ctx, "test-tool", input)
	if err != nil {
		t.Fatalf("Expected tool execution to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if resultMap["result"] != "mock tool result" {
		t.Error("Expected mock tool result")
	}
}

func TestExecuteToolNotFound(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	input := map[string]interface{}{
		"test_param": "test_value",
	}

	_, err = framework.ExecuteTool(ctx, "non-existent-tool", input)
	if err == nil {
		t.Error("Expected tool execution to fail for non-existent tool")
	}
}

// Test Flow registration and execution

func TestRegisterFlow(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{
		Flows: map[string]*interfaces.FlowConfig{
			"test-flow": {
				Enabled:  true,
				Timeout:  "60s",
				Parallel: false,
			},
		},
	}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	flow := NewMockFlow("test-flow", "Test flow", "1.0.0", "test")
	err = framework.RegisterFlow("test-flow", flow)
	if err != nil {
		t.Fatalf("Expected flow registration to succeed, got error: %v", err)
	}

	// Verify flow is in registry
	retrievedFlow, err := framework.GetFlowRegistry().Get("test-flow")
	if err != nil {
		t.Fatalf("Expected to retrieve flow, got error: %v", err)
	}

	if retrievedFlow.Name() != "test-flow" {
		t.Error("Expected retrieved flow to have correct name")
	}
}

func TestExecuteFlow(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	flow := NewMockFlow("test-flow", "Test flow", "1.0.0", "test")
	err = framework.RegisterFlow("test-flow", flow)
	if err != nil {
		t.Fatalf("Expected flow registration to succeed, got error: %v", err)
	}

	input := map[string]interface{}{
		"test_param": "test_value",
	}

	result, err := framework.ExecuteFlow(ctx, "test-flow", input)
	if err != nil {
		t.Fatalf("Expected flow execution to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if resultMap["result"] != "mock flow result" {
		t.Error("Expected mock flow result")
	}
}

func TestExecuteFlowNotFound(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	input := map[string]interface{}{
		"test_param": "test_value",
	}

	_, err = framework.ExecuteFlow(ctx, "non-existent-flow", input)
	if err == nil {
		t.Error("Expected flow execution to fail for non-existent flow")
	}
}

// Test Health and Metrics

func TestHealth(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Health check when not started
	health, err := framework.Health(ctx)
	if err != nil {
		t.Fatalf("Expected health check to succeed, got error: %v", err)
	}

	if health.Status != "unhealthy" {
		t.Errorf("Expected unhealthy status when not started, got: %s", health.Status)
	}

	// Start framework
	err = framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	// Health check when started
	health, err = framework.Health(ctx)
	if err != nil {
		t.Fatalf("Expected health check to succeed, got error: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected healthy status when started, got: %s", health.Status)
	}

	if health.Version == "" {
		t.Error("Expected version to be set")
	}

	if health.Components == nil {
		t.Error("Expected components to be set")
	}

	if health.Uptime <= 0 {
		t.Error("Expected uptime to be positive")
	}
}

func TestMetrics(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	// Execute some operations to generate metrics
	model := NewMockModel("test-model", "test")
	err = framework.RegisterModel("test-model", model)
	if err != nil {
		t.Fatalf("Expected model registration to succeed, got error: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	_, err = framework.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected generation to succeed, got error: %v", err)
	}

	// Get metrics
	metrics, err := framework.Metrics(ctx)
	if err != nil {
		t.Fatalf("Expected metrics retrieval to succeed, got error: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be non-nil")
	}

	if metrics.GeneratedRequests == nil {
		t.Error("Expected generation metrics to be set")
	}

	if metrics.GeneratedRequests.TotalRequests == 0 {
		t.Error("Expected at least one generation request")
	}

	if metrics.GeneratedRequests.SuccessCount == 0 {
		t.Error("Expected at least one successful generation")
	}

	if metrics.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

// Test error cases and edge conditions

func TestPonchoFrameworkNilConfig(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	framework := NewPonchoFramework(nil, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed with nil config, got error: %v", err)
	}

	// Should still be able to register components without config
	model := NewMockModel("test-model", "test")
	err = framework.RegisterModel("test-model", model)
	if err != nil {
		t.Fatalf("Expected model registration to succeed, got error: %v", err)
	}
}

func TestMetricsRecording(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	// Test error recording
	// This is a bit tricky since recordError is not exported
	// We'll trigger an error by trying to execute a non-existent model
	req := &interfaces.PonchoModelRequest{
		Model: "non-existent-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	_, err = framework.Generate(ctx, req)
	if err == nil {
		t.Error("Expected generation to fail for non-existent model")
	}

	// Check that error metrics were recorded
	metrics, err := framework.Metrics(ctx)
	if err != nil {
		t.Fatalf("Expected metrics retrieval to succeed, got error: %v", err)
	}

	if metrics.Errors == nil {
		t.Error("Expected error metrics to be set")
	}

	if metrics.Errors.TotalErrors == 0 {
		t.Error("Expected at least one error to be recorded")
	}
}

// Benchmark tests

func BenchmarkFrameworkGenerate(b *testing.B) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		b.Fatalf("Expected start to succeed, got error: %v", err)
	}

	model := NewMockModel("test-model", "test")
	err = framework.RegisterModel("test-model", model)
	if err != nil {
		b.Fatalf("Expected model registration to succeed, got error: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := framework.Generate(ctx, req)
		if err != nil {
			b.Fatalf("Expected generation to succeed, got error: %v", err)
		}
	}
}

func BenchmarkFrameworkExecuteTool(b *testing.B) {
	logger := interfaces.NewDefaultLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()
	err := framework.Start(ctx)
	if err != nil {
		b.Fatalf("Expected start to succeed, got error: %v", err)
	}

	tool := NewMockTool("test-tool", "Test tool", "1.0.0", "test")
	err = framework.RegisterTool("test-tool", tool)
	if err != nil {
		b.Fatalf("Expected tool registration to succeed, got error: %v", err)
	}

	input := map[string]interface{}{
		"test_param": "test_value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := framework.ExecuteTool(ctx, "test-tool", input)
		if err != nil {
			b.Fatalf("Expected tool execution to succeed, got error: %v", err)
		}
	}
}
