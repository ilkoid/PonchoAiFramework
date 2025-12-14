package articleflow

import (
	"context"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/wildberries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockModel is a mock implementation of PonchoModel
type MockModel struct {
	mock.Mock
	name      string
	supports  map[string]bool
}

func (m *MockModel) Name() string                     { return m.name }
func (m *MockModel) Provider() string                 { return "mock" }
func (m *MockModel) MaxTokens() int                  { return 4000 }
func (m *MockModel) DefaultTemperature() float32    { return 0.7 }
func (m *MockModel) SupportsStreaming() bool         { return m.supports["streaming"] }
func (m *MockModel) SupportsTools() bool            { return m.supports["tools"] }
func (m *MockModel) SupportsVision() bool            { return m.supports["vision"] }
func (m *MockModel) SupportsSystemRole() bool        { return m.supports["system"] }
func (m *MockModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}
func (m *MockModel) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *MockModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*interfaces.PonchoModelResponse), args.Error(1)
}
func (m *MockModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	args := m.Called(ctx, req, callback)
	return args.Error(0)
}

// MockTool is a mock implementation of PonchoTool
type MockTool struct {
	mock.Mock
	name string
}

func (m *MockTool) Name() string              { return m.name }
func (m *MockTool) Description() string       { return "Mock tool" }
func (m *MockTool) Version() string           { return "1.0.0" }
func (m *MockTool) Category() string          { return "mock" }
func (m *MockTool) Tags() []string            { return []string{"test"} }
func (m *MockTool) Dependencies() []string    { return []string{} }
func (m *MockTool) Initialize(ctx context.Context, config map[string]interface{}) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}
func (m *MockTool) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *MockTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}
func (m *MockTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
	}
}
func (m *MockTool) OutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
	}
}
func (m *MockTool) Validate(input interface{}) error {
	args := m.Called(input)
	return args.Error(0)
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	args := m.Called(append([]interface{}{msg}, fields...)...)
	_ = args
}
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	args := m.Called(append([]interface{}{msg}, fields...)...)
	_ = args
}
func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	args := m.Called(append([]interface{}{msg}, fields...)...)
	_ = args
}
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	args := m.Called(append([]interface{}{msg}, fields...)...)
	_ = args
}

func TestNewArticleFlow(t *testing.T) {
	s3Tool := &MockTool{name: "article_importer"}
	visionModel := &MockModel{name: "glm-4.6v", supports: map[string]bool{"vision": true}}
	textModel := &MockModel{name: "deepseek", supports: map[string]bool{}}
	logger := &MockLogger{}
	config := DefaultFlowConfig()
	cache := &MockWBCache{}

	flow := NewArticleFlow(s3Tool, visionModel, textModel, nil, nil, logger, config, cache)

	assert.NotNil(t, flow)
	assert.Equal(t, s3Tool, flow.s3Tool)
	assert.Equal(t, visionModel, flow.visionModel)
	assert.Equal(t, textModel, flow.textModel)
	assert.Equal(t, config, flow.config)
	assert.Equal(t, cache, flow.wbCache)
}

func TestArticleFlow_BuildVisionPrompt(t *testing.T) {
	s3Tool := &MockTool{}
	visionModel := &MockModel{}
	textModel := &MockModel{}
	logger := &MockLogger{}
	config := DefaultFlowConfig()
	cache := &MockWBCache{}

	flow := NewArticleFlow(s3Tool, visionModel, textModel, nil, nil, logger, config, cache)

	prompt := flow.buildVisionAnalysisPrompt()

	assert.Contains(t, prompt, "garment_type")
	assert.Contains(t, prompt, "silhouette")
	assert.Contains(t, prompt, "materials")
	assert.Contains(t, prompt, "colors")
}

func TestArticleFlow_BuildCreativePrompt(t *testing.T) {
	s3Tool := &MockTool{}
	visionModel := &MockModel{}
	textModel := &MockModel{}
	logger := &MockLogger{}
	config := DefaultFlowConfig()
	cache := &MockWBCache{}

	flow := NewArticleFlow(s3Tool, visionModel, textModel, nil, nil, logger, config, cache)

	plmJSON := []byte(`{"name": "Test Dress", "category": "Clothing"}`)
	prompt := flow.buildCreativePrompt(plmJSON)

	assert.Contains(t, prompt, "Test Dress")
	assert.Contains(t, prompt, "Clothing")
	assert.Contains(t, prompt, "fashion buyers")
	assert.Contains(t, prompt, "emotional connection")
}