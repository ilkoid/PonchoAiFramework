package prompts

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Type aliases for configuration structs
type TemplatesConfig struct {
	Directory     string   `yaml:"directory" json:"directory"`
	Extensions    []string `yaml:"extensions" json:"extensions"`
	AutoReload    bool     `yaml:"auto_reload" json:"auto_reload"`
	ReloadInterval string   `yaml:"reload_interval" json:"reload_interval"`
}

type CacheConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Size    int    `yaml:"size" json:"size"`
	TTL     string `yaml:"ttl" json:"ttl"`
	Type    string `yaml:"type" json:"type"`
}

type ValidationConfig struct {
	Strict          bool `yaml:"strict" json:"strict"`
	ValidateOnLoad  bool `yaml:"validate_on_load" json:"validate_on_load"`
	ValidateOnExec  bool `yaml:"validate_on_execute" json:"validate_on_execute"`
}

type ExecutionConfig struct {
	DefaultModel       string  `yaml:"default_model" json:"default_model"`
	DefaultMaxTokens   int     `yaml:"default_max_tokens" json:"default_max_tokens"`
	DefaultTemperature float32 `yaml:"default_temperature" json:"default_temperature"`
	Timeout           string  `yaml:"timeout" json:"timeout"`
	RetryAttempts     int     `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay        string  `yaml:"retry_delay" json:"retry_delay"`
}

// MockFramework is a mock implementation of PonchoFramework
type MockFramework struct {
	mock.Mock
}

func (m *MockFramework) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*interfaces.PonchoModelResponse), args.Error(1)
}

func (m *MockFramework) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	args := m.Called(ctx, req, callback)
	return args.Error(0)
}

func (m *MockFramework) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error) {
	args := m.Called(ctx, toolName, input)
	return args.Get(0), args.Error(1)
}

func (m *MockFramework) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error) {
	args := m.Called(ctx, flowName, input)
	return args.Get(0), args.Error(1)
}

func (m *MockFramework) RegisterModel(name string, model interfaces.PonchoModel) error {
	args := m.Called(name, model)
	return args.Error(0)
}

func (m *MockFramework) RegisterTool(name string, tool interfaces.PonchoTool) error {
	args := m.Called(name, tool)
	return args.Error(0)
}

func (m *MockFramework) RegisterFlow(name string, flow interfaces.PonchoFlow) error {
	args := m.Called(name, flow)
	return args.Error(0)
}

func (m *MockFramework) GetModel(name string) (interfaces.PonchoModel, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoModel), args.Error(1)
}

func (m *MockFramework) GetTool(name string) (interfaces.PonchoTool, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoTool), args.Error(1)
}

func (m *MockFramework) GetFlow(name string) (interfaces.PonchoFlow, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoFlow), args.Error(1)
}

func (m *MockFramework) ListModels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockFramework) ListTools() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockFramework) ListFlows() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockFramework) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockFramework) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockFramework) Health(ctx context.Context) (*interfaces.PonchoHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*interfaces.PonchoHealthStatus), args.Error(1)
}

func (m *MockFramework) Metrics(ctx context.Context) (*interfaces.PonchoMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*interfaces.PonchoMetrics), args.Error(1)
}

func (m *MockFramework) ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback interfaces.PonchoStreamCallback) error {
	args := m.Called(ctx, flowName, input, callback)
	return args.Error(0)
}

func (m *MockFramework) GetConfig() *interfaces.PonchoFrameworkConfig {
	args := m.Called()
	return args.Get(0).(*interfaces.PonchoFrameworkConfig)
}

func (m *MockFramework) GetModelRegistry() interfaces.PonchoModelRegistry {
	args := m.Called()
	return args.Get(0).(interfaces.PonchoModelRegistry)
}

func (m *MockFramework) GetToolRegistry() interfaces.PonchoToolRegistry {
	args := m.Called()
	return args.Get(0).(interfaces.PonchoToolRegistry)
}

func (m *MockFramework) GetFlowRegistry() interfaces.PonchoFlowRegistry {
	args := m.Called()
	return args.Get(0).(interfaces.PonchoFlowRegistry)
}

func (m *MockFramework) ReloadConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestNewPromptManager(t *testing.T) {
	mockFramework := &MockFramework{}
	mockLogger := &MockLogger{}
	config := &PromptConfig{
		Templates: TemplatesConfig{
			Directory:  "test_templates",
			Extensions: []string{".yaml", ".yml", ".json", ".prompt"},
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    100,
		},
		Validation: ValidationConfig{
			Strict:        false,
			ValidateOnLoad: false,
			ValidateOnExec: false,
		},
		Execution: ExecutionConfig{
			DefaultModel:       "test-model",
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
		},
	}

	pm := NewPromptManager(config, mockFramework, mockLogger)
	
	assert.NotNil(t, pm)
	assert.Implements(t, (*interfaces.PromptManager)(nil), pm)
}

func TestPromptManager_LoadTemplate(t *testing.T) {
	// Create temporary directory with test template
	tempDir, err := os.MkdirTemp("", "prompt_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	templateContent := `name: test_template
description: Test template for unit testing
version: 1.0.0
category: test
parts:
  - type: system
    content: "You are a helpful assistant for {{product_type}} analysis."
  - type: user
    content: "Analyze this {{product_type}}: {{description}}"
variables:
  - name: product_type
    type: string
    description: Type of product to analyze
    required: true
  - name: description
    type: string
    description: Product description
    required: true
`

	templatePath := filepath.Join(tempDir, "test_template.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	mockFramework := &MockFramework{}
	mockLogger := &MockLogger{}
	config := &PromptConfig{
		Templates: TemplatesConfig{
			Directory:  tempDir,
			Extensions: []string{".yaml", ".yml", ".json", ".prompt"},
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    100,
		},
		Validation: ValidationConfig{
			Strict:        false,
			ValidateOnLoad: false,
			ValidateOnExec: false,
		},
		Execution: ExecutionConfig{
			DefaultModel:       "test-model",
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
		},
	}

	// Setup mock logger expectations - allow any logger calls
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("Fatal", mock.Anything, mock.Anything).Return()

	pm := NewPromptManager(config, mockFramework, mockLogger)
	
	// Test loading template
	template, err := pm.LoadTemplate("test_template")
	assert.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "test_template", template.Name)
	assert.Equal(t, "Test template for unit testing", template.Description)
	assert.Equal(t, "1.0.0", template.Version)
	assert.Equal(t, "test", template.Category)
	assert.Len(t, template.Parts, 2)
	assert.Len(t, template.Variables, 2)
	
	mockLogger.AssertExpectations(t)
}

func TestPromptManager_ExecutePrompt(t *testing.T) {
	// Create temporary directory with test template
	tempDir, err := os.MkdirTemp("", "prompt_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	templateContent := `name: test_template
description: Test template
version: 1.0.0
category: test
parts:
  - type: user
    content: "Hello {{name}}, how are you?"
variables:
  - name: name
    type: string
    description: Name to greet
    required: true
`

	templatePath := filepath.Join(tempDir, "test_template.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	mockFramework := &MockFramework{}
	mockLogger := &MockLogger{}
	config := &PromptConfig{
		Templates: TemplatesConfig{
			Directory:  tempDir,
			Extensions: []string{".yaml", ".yml", ".json", ".prompt"},
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    100,
		},
		Validation: ValidationConfig{
			Strict:        false,
			ValidateOnLoad: false,
			ValidateOnExec: false,
		},
		Execution: ExecutionConfig{
			DefaultModel:       "test-model",
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
		},
	}

	pm := NewPromptManager(config, mockFramework, mockLogger)
	
	// Mock framework response
	expectedResponse := &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Hello John, I'm doing well!",
				},
			},
		},
		Usage: &interfaces.PonchoUsage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}

	// Setup mock logger expectations - allow any debug and error calls
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockFramework.On("Generate", mock.Anything, mock.Anything).Return(expectedResponse, nil)
	
	// Test execution
	ctx := context.Background()
	variables := map[string]interface{}{
		"name": "John",
	}
	
	response, err := pm.ExecutePrompt(ctx, "test_template", variables, "test-model")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Hello John, I'm doing well!", response.Message.Content[0].Text)
	
	mockFramework.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestPromptManager_ValidatePrompt(t *testing.T) {
	template := &interfaces.PromptTemplate{
		Name:        "test_template",
		Description: "Test template",
		Version:     "1.0.0",
		Category:    "test",
		Parts: []*interfaces.PromptPart{
			{
				Type:    interfaces.PromptPartTypeUser,
				Content: "Hello {{name}}",
			},
		},
		Variables: []*interfaces.PromptVariable{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockFramework := &MockFramework{}
	mockLogger := &MockLogger{}
	config := &PromptConfig{
		Templates: TemplatesConfig{
			Directory:  "test_templates",
			Extensions: []string{".yaml", ".yml", ".json", ".prompt"},
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    100,
		},
		Validation: ValidationConfig{
			Strict:        false,
			ValidateOnLoad: false,
			ValidateOnExec: false,
		},
		Execution: ExecutionConfig{
			DefaultModel:       "test-model",
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
		},
	}

	pm := NewPromptManager(config, mockFramework, mockLogger)
	
	// Test validation
	result, err := pm.ValidatePrompt(template)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Len(t, result.Errors, 0)
}

func TestPromptManager_ListTemplates(t *testing.T) {
	// Create temporary directory with test templates
	tempDir, err := os.MkdirTemp("", "prompt_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create multiple template files
	templates := map[string]string{
		"template1.yaml": `name: template1
description: First test template
version: 1.0.0
category: test
parts:
  - type: user
    content: "Template 1 content"
variables: []
`,
		"template2.yaml": `name: template2
description: Second test template
version: 1.0.0
category: test
parts:
  - type: user
    content: "Template 2 content"
variables: []
`,
	}

	for filename, content := range templates {
		templatePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(templatePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	mockFramework := &MockFramework{}
	mockLogger := &MockLogger{}
	config := &PromptConfig{
		Templates: TemplatesConfig{
			Directory:  tempDir,
			Extensions: []string{".yaml", ".yml", ".json", ".prompt"},
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    100,
		},
		Validation: ValidationConfig{
			Strict:        false,
			ValidateOnLoad: false,
			ValidateOnExec: false,
		},
		Execution: ExecutionConfig{
			DefaultModel:       "test-model",
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
		},
	}

	pm := NewPromptManager(config, mockFramework, mockLogger)
	
	// Test listing templates
	templateNames, err := pm.ListTemplates()
	assert.NoError(t, err)
	assert.Len(t, templateNames, 2)
	assert.Contains(t, templateNames, "template1")
	assert.Contains(t, templateNames, "template2")
}

func TestPromptManager_ReloadTemplates(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "prompt_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockFramework := &MockFramework{}
	mockLogger := &MockLogger{}
	config := &PromptConfig{
		Templates: TemplatesConfig{
			Directory:  tempDir,
			Extensions: []string{".yaml", ".yml", ".json", ".prompt"},
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    100,
		},
		Validation: ValidationConfig{
			Strict:        false,
			ValidateOnLoad: false,
			ValidateOnExec: false,
		},
		Execution: ExecutionConfig{
			DefaultModel:       "test-model",
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
		},
	}

	pm := NewPromptManager(config, mockFramework, mockLogger)
	
	// Test reload (should not fail even with empty directory)
	err = pm.ReloadTemplates()
	assert.NoError(t, err)
}

func TestPromptCache(t *testing.T) {
	mockLogger := &MockLogger{}
	// Setup mock logger expectations - allow any debug and error calls
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	
	cache := NewPromptCache(2, mockLogger)
	
	// Test cache operations
	template := &interfaces.PromptTemplate{
		Name:        "test_template",
		Description: "Test template",
		Version:     "1.0.0",
		Category:    "test",
		Parts:       []*interfaces.PromptPart{},
		Variables:   []*interfaces.PromptVariable{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test Set and Get
	cache.SetTemplate("test_template", template)
	retrieved, found := cache.GetTemplate("test_template")
	assert.True(t, found)
	assert.Equal(t, template, retrieved)

	// Test cache miss
	_, found = cache.GetTemplate("nonexistent")
	assert.False(t, found)

	// Test Stats
	stats := cache.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Size)
	assert.Equal(t, int64(2), stats.MaxSize)

	// Test Invalidate
	cache.InvalidateTemplate("test_template")
	_, found = cache.GetTemplate("test_template")
	assert.False(t, found)

	// Test Clear
	cache.SetTemplate("test_template", template)
	cache.Clear()
	_, found = cache.GetTemplate("test_template")
	assert.False(t, found)
	
	mockLogger.AssertExpectations(t)
}

func TestVariableProcessor(t *testing.T) {
	mockLogger := &MockLogger{}
	// Setup mock logger expectations - allow any debug and error calls
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	
	processor := NewVariableProcessor(mockLogger)

	template := &interfaces.PromptTemplate{
		Name: "test_template",
		Parts: []*interfaces.PromptPart{
			{
				Type:    interfaces.PromptPartTypeUser,
				Content: "Hello {{name}}, you are {{age}} years old.",
			},
		},
		Variables: []*interfaces.PromptVariable{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:         "age",
				Type:         "number",
				Required:     false,
				DefaultValue: 25,
			},
		},
	}

	// Test ProcessVariables
	content := "Hello {{name}}, you are {{age}} years old."
	variables := map[string]interface{}{
		"name": "John",
	}
	
	processed, err := processor.ProcessVariables(content, variables)
	assert.NoError(t, err)
	assert.Equal(t, "Hello John, you are {{age}} years old.", processed)

	// Test ExtractVariables
	extracted, err := processor.ExtractVariables(content)
	assert.NoError(t, err)
	assert.Len(t, extracted, 2)
	assert.Contains(t, extracted, "name")
	assert.Contains(t, extracted, "age")

	// Test ValidateVariables
	err = processor.ValidateVariables(template, variables)
	assert.NoError(t, err)

	// Test missing required variable
	err = processor.ValidateVariables(template, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required variable 'name' is missing")

	// Test SetDefaults
	variables = processor.SetDefaults(template, map[string]interface{}{})
	assert.Equal(t, "John", variables["name"])
	assert.Equal(t, 25, variables["age"])
	
	mockLogger.AssertExpectations(t)
}