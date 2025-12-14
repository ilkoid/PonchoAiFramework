package prompts

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFramework for security testing
type MockSecurityFramework struct {
	mock.Mock
}

func (m *MockSecurityFramework) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*interfaces.PonchoModelResponse), args.Error(1)
}

func (m *MockSecurityFramework) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	args := m.Called(ctx, req, callback)
	return args.Error(0)
}

func (m *MockSecurityFramework) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error) {
	args := m.Called(ctx, toolName, input)
	return args.Get(0), args.Error(1)
}

func (m *MockSecurityFramework) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error) {
	args := m.Called(ctx, flowName, input)
	return args.Get(0), args.Error(1)
}

func (m *MockSecurityFramework) RegisterModel(name string, model interfaces.PonchoModel) error {
	args := m.Called(name, model)
	return args.Error(0)
}

func (m *MockSecurityFramework) RegisterTool(name string, tool interfaces.PonchoTool) error {
	args := m.Called(name, tool)
	return args.Error(0)
}

func (m *MockSecurityFramework) RegisterFlow(name string, flow interfaces.PonchoFlow) error {
	args := m.Called(name, flow)
	return args.Error(0)
}

func (m *MockSecurityFramework) GetModel(name string) (interfaces.PonchoModel, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoModel), args.Error(1)
}

func (m *MockSecurityFramework) GetTool(name string) (interfaces.PonchoTool, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoTool), args.Error(1)
}

func (m *MockSecurityFramework) GetFlow(name string) (interfaces.PonchoFlow, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoFlow), args.Error(1)
}

func (m *MockSecurityFramework) ListModels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockSecurityFramework) ListTools() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockSecurityFramework) ListFlows() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockSecurityFramework) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSecurityFramework) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSecurityFramework) Health(ctx context.Context) (*interfaces.PonchoHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*interfaces.PonchoHealthStatus), args.Error(1)
}

func (m *MockSecurityFramework) Metrics(ctx context.Context) (*interfaces.PonchoMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*interfaces.PonchoMetrics), args.Error(1)
}

func (m *MockSecurityFramework) ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback interfaces.PonchoStreamCallback) error {
	args := m.Called(ctx, flowName, input, callback)
	return args.Error(0)
}

func (m *MockSecurityFramework) GetConfig() *interfaces.PonchoFrameworkConfig {
	args := m.Called()
	return args.Get(0).(*interfaces.PonchoFrameworkConfig)
}

func (m *MockSecurityFramework) GetModelRegistry() interfaces.PonchoModelRegistry {
	args := m.Called()
	return args.Get(0).(interfaces.PonchoModelRegistry)
}

func (m *MockSecurityFramework) GetToolRegistry() interfaces.PonchoToolRegistry {
	args := m.Called()
	return args.Get(0).(interfaces.PonchoToolRegistry)
}

func (m *MockSecurityFramework) GetFlowRegistry() interfaces.PonchoFlowRegistry {
	args := m.Called()
	return args.Get(0).(interfaces.PonchoFlowRegistry)
}

func (m *MockSecurityFramework) ReloadConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockLogger for security testing
type MockSecurityLogger struct {
	mock.Mock
}

func (m *MockSecurityLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockSecurityLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockSecurityLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockSecurityLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func TestVariableProcessor_InjectionProtection(t *testing.T) {
	logger := &MockSecurityLogger{}
	logger.On("Warn", mock.Anything, mock.Anything).Maybe().Return()
	
	processor := NewVariableProcessor(logger)

	tests := []struct {
		name          string
		content       string
		variables     map[string]interface{}
		expectedResult string
		shouldWarn     bool
	}{
		{
			name:      "normal variable substitution",
			content:   "Hello {{name}}, you are {{age}} years old",
			variables: map[string]interface{}{"name": "John", "age": 25},
			expectedResult: "Hello John, you are 25 years old",
			shouldWarn: false,
		},
		{
			name:      "script injection attempt",
			content:   "Hello {{name}}",
			variables: map[string]interface{}{"name": "<script>alert('xss')</script>"},
			expectedResult: "Hello <script>alert(&#x27;xss&#x27;)</script>",
			shouldWarn: false,
		},
		{
			name:      "javascript injection attempt",
			content:   "URL: {{url}}",
			variables: map[string]interface{}{"url": "javascript:alert('xss')"},
			expectedResult: "URL: ",
			shouldWarn: false,
		},
		{
			name:      "data URI injection attempt",
			content:   "Image: {{image}}",
			variables: map[string]interface{}{"image": "data:text/html,<script>alert('xss')</script>"},
			expectedResult: "Image: ",
			shouldWarn: false,
		},
		{
			name:      "vbscript injection attempt",
			content:   "Content: {{content}}",
			variables: map[string]interface{}{"content": "vbscript:msgbox('xss')"},
			expectedResult: "Content: ",
			shouldWarn: false,
		},
		{
			name:      "onload injection attempt",
			content:   "Image: {{image}}",
			variables: map[string]interface{}{"image": "<img onload=alert('xss')>"},
			expectedResult: "Image: <img onload=alert(&#x27;xss&#x27;)>",
			shouldWarn: false,
		},
		{
			name:      "onerror injection attempt",
			content:   "Image: {{image}}",
			variables: map[string]interface{}{"image": "<img onerror=alert('xss')>"},
			expectedResult: "Image: <img onerror=alert(&#x27;xss&#x27;)>",
			shouldWarn: false,
		},
		{
			name:      "onclick injection attempt",
			content:   "Button: {{button}}",
			variables: map[string]interface{}{"button": "<button onclick=alert('xss')>Click me</button>"},
			expectedResult: "Button: <button onclick=alert(&#x27;xss&#x27;)>Click me</button>",
			shouldWarn: false,
		},
		{
			name:      "HTML injection attempt",
			content:   "Content: {{content}}",
			variables: map[string]interface{}{"content": "<div>malicious content</div>"},
			expectedResult: "Content: <div>malicious content</div>",
			shouldWarn: false,
		},
		{
			name:      "missing variable",
			content:   "Hello {{missing}} and {{name}}",
			variables: map[string]interface{}{"name": "John"},
			expectedResult: "Hello  and John",
			shouldWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessVariables(tt.content, tt.variables)
			
			assert.NoError(t, err, "ProcessVariables should not return error")
			assert.Equal(t, tt.expectedResult, result, "Result should match expected sanitized output")
			
			if tt.shouldWarn {
				logger.AssertCalled(t, "Warn", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestSanitizeVariableValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "string with script tag",
			input:    "<script>alert('xss')</script>",
			expected: "<script>alert(&#x27;xss&#x27;)</script>",
		},
		{
			name:     "string with javascript protocol",
			input:    "javascript:alert('xss')",
			expected: "",
		},
		{
			name:     "string with data protocol",
			input:    "data:text/html,<script>alert('xss')</script>",
			expected: "",
		},
		{
			name:     "string with vbscript protocol",
			input:    "vbscript:msgbox('xss')",
			expected: "",
		},
		{
			name:     "string with onload handler",
			input:    "<img onload=alert('xss')>",
			expected: "<img onload=alert(&#x27;xss&#x27;)>",
		},
		{
			name:     "string with onerror handler",
			input:    "<img onerror=alert('xss')>",
			expected: "<img onerror=alert(&#x27;xss&#x27;)>",
		},
		{
			name:     "string with onclick handler",
			input:    "<button onclick=alert('xss')>Click</button>",
			expected: "<button onclick=alert(&#x27;xss&#x27;)>Click</button>",
		},
		{
			name:     "string with HTML entities",
			input:    "<div>content</div>",
			expected: "<div>content</div>",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "number input",
			input:    42,
			expected: "42",
		},
		{
			name:     "boolean input",
			input:    true,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeVariableValue(tt.input)
			assert.Equal(t, tt.expected, result, "Sanitized value should match expected")
		})
	}
}