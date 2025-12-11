# PonchoFramework - Core Interfaces Specification

## Overview

Документ определяет детальные спецификации всех core интерфейсов и структур PonchoFramework. Эти интерфейсы служат основой для всего фреймворка и обеспечивают унифицированную работу с AI-моделями, инструментами и workflow.

## Core Interfaces

### 1. PonchoFramework (Main Class)

**Назначение:** Центральный оркестратор и главный контейнер всего AI-фреймворка.

```go
type PonchoFramework struct {
    // Core registries
    models  *PonchoModelRegistry
    tools   *PonchoToolRegistry
    flows   *PonchoFlowRegistry
    prompts *PonchoPromptManager
    
    // Configuration
    config  *PonchoFrameworkConfig
    logger  *PonchoLogger
    
    // Runtime state
    started bool
    metrics *PonchoMetrics
}

// Constructor
func NewPonchoFramework(config *PonchoFrameworkConfig) *PonchoFramework

// Lifecycle management
func (pf *PonchoFramework) Start(ctx context.Context) error
func (pf *PonchoFramework) Stop(ctx context.Context) error
func (pf *PonchoFramework) IsStarted() bool

// Component registration
func (pf *PonchoFramework) RegisterModel(name string, model PonchoModel) error
func (pf *PonchoFramework) RegisterTool(name string, tool PonchoTool) error
func (pf *PonchoFramework) RegisterFlow(name string, flow PonchoFlow) error

// Core operations - замена genkit.Generate()
func (pf *PonchoFramework) Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
func (pf *PonchoFramework) GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error

// Component execution
func (pf *PonchoFramework) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
func (pf *PonchoFramework) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error)
func (pf *PonchoFramework) ExecutePrompt(ctx context.Context, promptName string, input map[string]interface{}) (interface{}, error)

// Component access
func (pf *PonchoFramework) GetModel(name string) (PonchoModel, error)
func (pf *PonchoFramework) GetTool(name string) (PonchoTool, error)
func (pf *PonchoFramework) GetFlow(name string) (PonchoFlow, error)
func (pf *PonchoFramework) GetPrompt(name string) (*PonchoPrompt, error)

// Utility methods
func (pf *PonchoFramework) ListModels() []string
func (pf *PonchoFramework) ListTools() []string
func (pf *PonchoFramework) ListFlows() []string
func (pf *PonchoFramework) ListPrompts() []string
```

### 2. PonchoModel (Interface)

**Назначение:** Унифицированный интерфейс для всех AI-моделей.

```go
type PonchoModel interface {
    // Core functionality
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error
    
    // Capabilities
    SupportsStreaming() bool
    SupportsTools() bool
    SupportsVision() bool
    SupportsSystemRole() bool
    
    // Metadata
    Name() string
    Provider() string
    MaxTokens() int
    DefaultTemperature() float32
    
    // Configuration
    Configure(config *PonchoModelConfig) error
    GetConfig() *PonchoModelConfig
    
    // Validation
    ValidateRequest(req *PonchoModelRequest) error
}

// Base implementation
type PonchoBaseModel struct {
    name     string
    provider string
    config   *PonchoModelConfig
    logger   *PonchoLogger
}

func (m *PonchoBaseModel) Name() string { return m.name }
func (m *PonchoBaseModel) Provider() string { return m.provider }
func (m *PonchoBaseModel) GetConfig() *PonchoModelConfig { return m.config }
```

### 3. PonchoTool (Interface)

**Назначение:** Унифицированный интерфейс для всех инструментов.

```go
type PonchoTool interface {
    // Core information
    Name() string
    Description() string
    Version() string
    
    // Execution
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // Schema
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
    
    // Validation
    Validate(input interface{}) error
    ValidateInput(input interface{}) error
    ValidateOutput(output interface{}) error
    
    // Metadata
    Category() string
    Tags() []string
    Dependencies() []string
    
    // Configuration
    Configure(config *PonchoToolConfig) error
    GetConfig() *PonchoToolConfig
}

// Base implementation
type PonchoBaseTool struct {
    name        string
    description string
    version     string
    category    string
    tags        []string
    config      *PonchoToolConfig
    logger      *PonchoLogger
}

func (t *PonchoBaseTool) Name() string { return t.name }
func (t *PonchoBaseTool) Description() string { return t.description }
func (t *PonchoBaseTool) Version() string { return t.version }
func (t *PonchoBaseTool) Category() string { return t.category }
func (t *PonchoBaseTool) Tags() []string { return t.tags }
func (t *PonchoBaseTool) GetConfig() *PonchoToolConfig { return t.config }
```

### 4. PonchoFlow (Interface)

**Назначение:** Интерфейс для оркестрации многоэтапных процессов.

```go
type PonchoFlow interface {
    // Core information
    Name() string
    Description() string
    Version() string
    
    // Execution
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    ExecuteStreaming(ctx context.Context, input interface{}, callback PonchoStreamCallback) error
    
    // Schema
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
    
    // Metadata
    Category() string
    Tags() []string
    Dependencies() []string // Required tools/models
    
    // Configuration
    Configure(config *PonchoFlowConfig) error
    GetConfig() *PonchoFlowConfig
    
    // Validation
    Validate(input interface{}) error
    ValidateInput(input interface{}) error
    ValidateOutput(output interface{}) error
}

// Base implementation
type PonchoBaseFlow struct {
    name         string
    description  string
    version      string
    category     string
    tags         []string
    dependencies []string
    config       *PonchoFlowConfig
    framework    *PonchoFramework
    logger       *PonchoLogger
}

func (f *PonchoBaseFlow) Name() string { return f.name }
func (f *PonchoBaseFlow) Description() string { return f.description }
func (f *PonchoBaseFlow) Version() string { return f.version }
func (f *PonchoBaseFlow) Category() string { return f.category }
func (f *PonchoBaseFlow) Tags() []string { return f.tags }
func (f *PonchoBaseFlow) Dependencies() []string { return f.dependencies }
func (f *PonchoBaseFlow) GetConfig() *PonchoFlowConfig { return f.config }
```

## Core Data Structures

### 1. Request/Response Types

```go
// Model request - замена ai.ModelRequest
type PonchoModelRequest struct {
    // Core parameters
    Model       string                 `json:"model"`
    Messages    []*PonchoMessage       `json:"messages"`
    
    // Generation parameters
    Temperature float32                `json:"temperature,omitempty"`
    MaxTokens   int                    `json:"max_tokens,omitempty"`
    TopP        float32                `json:"top_p,omitempty"`
    
    // Tools
    Tools       []*PonchoTool          `json:"tools,omitempty"`
    ToolChoice  PonchoToolChoice       `json:"tool_choice,omitempty"`
    
    // Streaming
    Stream      bool                   `json:"stream,omitempty"`
    
    // Vision
    Media       []*PonchoMediaPart     `json:"media,omitempty"`
    
    // Configuration
    Config      *PonchoModelConfig     `json:"config,omitempty"`
    
    // Context
    Context     map[string]interface{}  `json:"context,omitempty"`
    Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

// Model response - замена ai.ModelResponse
type PonchoModelResponse struct {
    // Core response
    Message     *PonchoMessage          `json:"message"`
    FinishReason PonchoFinishReason       `json:"finish_reason"`
    
    // Usage statistics
    Usage       *PonchoUsage             `json:"usage"`
    
    // Metadata
    Model       string                   `json:"model"`
    Duration    time.Duration            `json:"duration"`
    Metadata    map[string]interface{}  `json:"metadata,omitempty"`
    
    // Error information
    Error       error                    `json:"error,omitempty"`
    ErrorCode   string                   `json:"error_code,omitempty"`
}

// Message structure - замена ai.Message
type PonchoMessage struct {
    Role    PonchoMessageRole       `json:"role"`
    Content []*PonchoContentPart    `json:"content"`
    Name    string                   `json:"name,omitempty"`
    Metadata map[string]interface{}   `json:"metadata,omitempty"`
}

// Content parts - замена ai.Part
type PonchoContentPart struct {
    Type       PonchoContentType       `json:"type"`
    Text       string                  `json:"text,omitempty"`
    Media      *PonchoMediaPart       `json:"media,omitempty"`
    ToolCall   *PonchoToolCall       `json:"tool_call,omitempty"`
    ToolResp   *PonchoToolResponse   `json:"tool_response,omitempty"`
    Metadata   map[string]interface{}  `json:"metadata,omitempty"`
}

// Media part for vision capabilities
type PonchoMediaPart struct {
    URL         string    `json:"url"`
    ContentType string    `json:"content_type"`
    Data        []byte    `json:"data,omitempty"`
    Size        int64     `json:"size"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Tool call structure
type PonchoToolCall struct {
    ID       string                 `json:"id"`
    Name     string                 `json:"name"`
    Input    map[string]interface{} `json:"input"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Tool response structure
type PonchoToolResponse struct {
    ID       string                 `json:"id"`
    Output   interface{}            `json:"output"`
    Error    error                  `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

### 2. Enums and Constants

```go
// Message roles
type PonchoMessageRole string

const (
    PonchoRoleSystem    PonchoMessageRole = "system"
    PonchoRoleUser      PonchoMessageRole = "user"
    PonchoRoleModel     PonchoMessageRole = "model"
    PonchoRoleTool      PonchoMessageRole = "tool"
)

// Content types
type PonchoContentType string

const (
    PonchoContentTypeText       PonchoContentType = "text"
    PonchoContentTypeMedia      PonchoContentType = "media"
    PonchoContentTypeToolCall   PonchoContentType = "tool_call"
    PonchoContentTypeToolResp   PonchoContentType = "tool_response"
)

// Finish reasons
type PonchoFinishReason string

const (
    PonchoFinishReasonStop      PonchoFinishReason = "stop"
    PonchoFinishReasonLength    PonchoFinishReason = "length"
    PonchoFinishReasonToolCalls PonchoFinishReason = "tool_calls"
    PonchoFinishReasonError     PonchoFinishReason = "error"
)

// Tool choice types
type PonchoToolChoice string

const (
    PonchoToolChoiceNone   PonchoToolChoice = "none"
    PonchoToolChoiceAuto   PonchoToolChoice = "auto"
    PonchoToolChoiceRequired PonchoToolChoice = "required"
)
```

### 3. Configuration Structures

```go
// Framework configuration
type PonchoFrameworkConfig struct {
    // Model configurations
    Models map[string]*PonchoModelConfig `json:"models" yaml:"models"`
    
    // Tool configurations
    Tools map[string]*PonchoToolConfig `json:"tools" yaml:"tools"`
    
    // Flow configurations
    Flows map[string]*PonchoFlowConfig `json:"flows" yaml:"flows"`
    
    // Prompt configurations
    Prompts *PonchoPromptConfig `json:"prompts" yaml:"prompts"`
    
    // Global settings
    Logging   *PonchoLoggingConfig   `json:"logging" yaml:"logging"`
    Metrics   *PonchoMetricsConfig   `json:"metrics" yaml:"metrics"`
    Cache     *PonchoCacheConfig     `json:"cache" yaml:"cache"`
    Security  *PonchoSecurityConfig  `json:"security" yaml:"security"`
}

// Model configuration
type PonchoModelConfig struct {
    Provider     string            `json:"provider" yaml:"provider"`
    ModelName    string            `json:"model_name" yaml:"model_name"`
    APIKey       string            `json:"api_key" yaml:"api_key"`
    Endpoint     string            `json:"endpoint" yaml:"endpoint"`
    MaxTokens    int               `json:"max_tokens" yaml:"max_tokens"`
    Temperature  float32           `json:"temperature" yaml:"temperature"`
    Timeout      time.Duration     `json:"timeout" yaml:"timeout"`
    Supports     struct {
        Vision   bool `json:"vision" yaml:"vision"`
        Tools    bool `json:"tools" yaml:"tools"`
        Stream   bool `json:"stream" yaml:"stream"`
        System   bool `json:"system" yaml:"system"`
    } `json:"supports" yaml:"supports"`
    Custom       map[string]interface{} `json:"custom" yaml:"custom"`
}

// Tool configuration
type PonchoToolConfig struct {
    Enabled      bool                   `json:"enabled" yaml:"enabled"`
    Timeout      time.Duration          `json:"timeout" yaml:"timeout"`
    Retry        *PonchoRetryConfig     `json:"retry" yaml:"retry"`
    Cache        *PonchoCacheConfig     `json:"cache" yaml:"cache"`
    Dependencies []string               `json:"dependencies" yaml:"dependencies"`
    Custom       map[string]interface{}  `json:"custom" yaml:"custom"`
}

// Flow configuration
type PonchoFlowConfig struct {
    Enabled      bool                   `json:"enabled" yaml:"enabled"`
    Timeout      time.Duration          `json:"timeout" yaml:"timeout"`
    Retry        *PonchoRetryConfig     `json:"retry" yaml:"retry"`
    Parallel     bool                   `json:"parallel" yaml:"parallel"`
    Dependencies []string               `json:"dependencies" yaml:"dependencies"`
    Custom       map[string]interface{}  `json:"custom" yaml:"custom"`
}
```

### 4. Registry Interfaces

```go
// Model registry
type PonchoModelRegistry interface {
    Register(name string, model PonchoModel) error
    Get(name string) (PonchoModel, error)
    List() []string
    Unregister(name string) error
    Clear() error
}

// Tool registry
type PonchoToolRegistry interface {
    Register(name string, tool PonchoTool) error
    Get(name string) (PonchoTool, error)
    List() []string
    ListByCategory(category string) []string
    Unregister(name string) error
    Clear() error
}

// Flow registry
type PonchoFlowRegistry interface {
    Register(name string, flow PonchoFlow) error
    Get(name string) (PonchoFlow, error)
    List() []string
    ListByCategory(category string) []string
    Unregister(name string) error
    Clear() error
}
```

### 5. Callback Interfaces

```go
// Streaming callback
type PonchoStreamCallback func(chunk *PonchoStreamChunk) error

// Stream chunk
type PonchoStreamChunk struct {
    Type      PonchoStreamChunkType  `json:"type"`
    Content   string                 `json:"content"`
    Delta     string                 `json:"delta,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Error     error                  `json:"error,omitempty"`
    Done      bool                   `json:"done"`
}

// Stream chunk types
type PonchoStreamChunkType string

const (
    PonchoStreamChunkStart   PonchoStreamChunkType = "start"
    PonchoStreamChunkDelta   PonchoStreamChunkType = "delta"
    PonchoStreamChunkContent PonchoStreamChunkType = "content"
    PonchoStreamChunkEnd    PonchoStreamChunkType = "end"
    PonchoStreamChunkError  PonchoStreamChunkType = "error"
)
```

## Usage Examples

### Basic Model Generation
```go
// Initialize framework
config := &PonchoFrameworkConfig{
    Models: map[string]*PonchoModelConfig{
        "deepseek-chat": {
            Provider:  "deepseek",
            ModelName: "deepseek-chat",
            APIKey:    os.Getenv("DEEPSEEK_API_KEY"),
        },
    },
}

framework := NewPonchoFramework(config)
framework.Start(ctx)

// Generate response
response, err := framework.Generate(ctx, &PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*PonchoMessage{
        {
            Role: PonchoRoleUser,
            Content: []*PonchoContentPart{
                {
                    Type: PonchoContentTypeText,
                    Text: "Проанализируй товар с артикулом 12611516",
                },
            },
        },
    },
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### Tool Execution
```go
// Execute tool
result, err := framework.ExecuteTool(ctx, "importArticleData", map[string]interface{}{
    "article_id":     "12611516",
    "include_images": true,
    "max_images":     3,
})
```

### Flow Execution
```go
// Execute flow
result, err := framework.ExecuteFlow(ctx, "articleImporter", map[string]interface{}{
    "article_id": "12611516",
    "mode":       "full",
})
```

---

*Эта спецификация служит основой для реализации всех компонентов PonchoFramework и обеспечивает полную совместимость с существующим функционалом.*