# PonchoFramework - API Documentation

## Overview

PonchoFramework предоставляет унифицированный API для работы с AI-моделями, инструментами и workflow. Документация содержит полное описание всех методов, примеры использования и руководства по миграции с GenKit.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Core API](#core-api)
3. [Model API](#model-api)
4. [Tool API](#tool-api)
5. [Flow API](#flow-api)
6. [Prompt API](#prompt-api)
7. [Streaming API](#streaming-api)
8. [Configuration API](#configuration-api)
9. [Migration Guide](#migration-guide)
10. [Error Handling](#error-handling)
11. [Best Practices](#best-practices)

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/poncho-tools/poncho-framework"
)

func main() {
    ctx := context.Background()
    
    // 1. Create configuration
    config := &poncho.PonchoFrameworkConfig{
        Models: map[string]*poncho.PonchoModelConfig{
            "deepseek-chat": {
                Provider:     "deepseek",
                ModelName:    "deepseek-chat",
                APIKey:       os.Getenv("DEEPSEEK_API_KEY"),
                MaxTokens:    4000,
                Temperature:  0.7,
                Timeout:      30 * time.Second,
                Supports: struct {
                    Vision bool `json:"vision" yaml:"vision"`
                    Tools  bool `json:"tools" yaml:"tools"`
                    Stream bool `json:"stream" yaml:"stream"`
                    System bool `json:"system" yaml:"system"`
                }{
                    Vision: false,
                    Tools:  true,
                    Stream: true,
                    System: true,
                },
            },
        },
        Logging: &poncho.PonchoLoggingConfig{
            Level:  "info",
            Format: "json",
        },
    }
    
    // 2. Initialize framework
    framework := poncho.NewPonchoFramework(config)
    
    // 3. Start framework
    if err := framework.Start(ctx); err != nil {
        log.Fatal("Failed to start framework:", err)
    }
    defer framework.Stop(ctx)
    
    // 4. Use framework
    response, err := framework.Generate(ctx, &poncho.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*poncho.PonchoMessage{
            {
                Role: poncho.PonchoRoleUser,
                Content: []*poncho.PonchoContentPart{
                    {
                        Type: poncho.PonchoContentTypeText,
                        Text: "Привет! Расскажи о PonchoFramework.",
                    },
                },
            },
        },
    })
    
    if err != nil {
        log.Fatal("Generation failed:", err)
    }
    
    fmt.Println("Response:", response.Message.Content[0].Text)
}
```

## Core API

### PonchoFramework Methods

#### Constructor

```go
func NewPonchoFramework(config *PonchoFrameworkConfig) *PonchoFramework
```

**Description:** Создает новый экземпляр PonchoFramework с указанной конфигурацией.

**Parameters:**
- `config`: Конфигурация фреймворка

**Returns:**
- `*PonchoFramework`: Новый экземпляр фреймворка

**Example:**
```go
config := &poncho.PonchoFrameworkConfig{
    Models: map[string]*poncho.PonchoModelConfig{
        "glm-chat": {
            Provider:  "zai",
            ModelName: "glm-4.6",
            APIKey:    os.Getenv("ZAI_API_KEY"),
        },
    },
}

framework := poncho.NewPonchoFramework(config)
```

#### Lifecycle Management

```go
func (pf *PonchoFramework) Start(ctx context.Context) error
func (pf *PonchoFramework) Stop(ctx context.Context) error
func (pf *PonchoFramework) IsStarted() bool
```

**Description:** Управление жизненным циклом фреймворка.

**Example:**
```go
ctx := context.Background()

// Start framework
if err := framework.Start(ctx); err != nil {
    return fmt.Errorf("failed to start: %w", err)
}
defer framework.Stop(ctx)

// Check status
if framework.IsStarted() {
    fmt.Println("Framework is running")
}
```

#### Core Generation Methods

```go
func (pf *PonchoFramework) Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
func (pf *PonchoFramework) GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error
```

**Description:** Основные методы генерации ответов от AI-моделей.

**Parameters:**
- `ctx`: Контекст выполнения
- `req`: Запрос к модели
- `callback`: Функция обратного вызова для стриминга

**Returns:**
- `*PonchoModelResponse`: Ответ модели
- `error`: Ошибка выполнения

**Example:**
```go
// Simple generation
response, err := framework.Generate(ctx, &poncho.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*poncho.PonchoMessage{
        {
            Role: poncho.PonchoRoleUser,
            Content: []*poncho.PonchoContentPart{
                {
                    Type: poncho.PonchoContentTypeText,
                    Text: "Проанализируй товар с артикулом 12611516",
                },
            },
        },
    },
    Temperature: 0.7,
    MaxTokens:   1000,
})

// Streaming generation
err = framework.GenerateStreaming(ctx, &poncho.PonchoModelRequest{
    Model: "glm-chat",
    Messages: []*poncho.PonchoMessage{
        {
            Role: poncho.PonchoRoleUser,
            Content: []*poncho.PonchoContentPart{
                {
                    Type: poncho.PonchoContentTypeText,
                    Text: "Напиши описание товара",
                },
            },
        },
    },
}, func(chunk *poncho.PonchoStreamChunk) error {
    if chunk.Type == poncho.PonchoStreamChunkDelta {
        fmt.Print(chunk.Delta)
    }
    return nil
})
```

#### Component Registration

```go
func (pf *PonchoFramework) RegisterModel(name string, model PonchoModel) error
func (pf *PonchoFramework) RegisterTool(name string, tool PonchoTool) error
func (pf *PonchoFramework) RegisterFlow(name string, flow PonchoFlow) error
```

**Description:** Регистрация компонентов в фреймворке.

**Example:**
```go
// Register custom model
customModel := &MyCustomModel{
    name:     "my-model",
    provider: "custom",
}

if err := framework.RegisterModel("my-model", customModel); err != nil {
    return fmt.Errorf("failed to register model: %w", err)
}

// Register custom tool
articleTool := &ArticleImporterTool{
    name: "articleImporter",
}

if err := framework.RegisterTool("articleImporter", articleTool); err != nil {
    return fmt.Errorf("failed to register tool: %w", err)
}
```

#### Component Execution

```go
func (pf *PonchoFramework) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
func (pf *PonchoFramework) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error)
func (pf *PonchoFramework) ExecutePrompt(ctx context.Context, promptName string, input map[string]interface{}) (interface{}, error)
```

**Description:** Выполнение зарегистрированных компонентов.

**Example:**
```go
// Execute tool
result, err := framework.ExecuteTool(ctx, "importArticleData", map[string]interface{}{
    "article_id":     "12611516",
    "include_images": true,
    "max_images":     3,
})

// Execute flow
result, err := framework.ExecuteFlow(ctx, "articleImporter", map[string]interface{}{
    "article_id": "12611516",
    "mode":       "full",
})

// Execute prompt
result, err := framework.ExecutePrompt(ctx, "sketchAnalysis", map[string]interface{}{
    "imageUrl":     "https://example.com/image.jpg",
    "customPrompt":  "Проанализируй эскиз",
})
```

## Model API

### PonchoModel Interface

```go
type PonchoModel interface {
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error
    
    SupportsStreaming() bool
    SupportsTools() bool
    SupportsVision() bool
    SupportsSystemRole() bool
    
    Name() string
    Provider() string
    MaxTokens() int
    DefaultTemperature() float32
    
    Configure(config *PonchoModelConfig) error
    GetConfig() *PonchoModelConfig
    
    ValidateRequest(req *PonchoModelRequest) error
}
```

### Custom Model Implementation

```go
type MyCustomModel struct {
    *poncho.PonchoBaseModel
    client *http.Client
    apiKey string
}

func NewMyCustomModel(config *poncho.PonchoModelConfig) *MyCustomModel {
    return &MyCustomModel{
        PonchoBaseModel: &poncho.PonchoBaseModel{
            name:     config.ModelName,
            provider: config.Provider,
            config:   config,
        },
        client: &http.Client{Timeout: config.Timeout},
        apiKey: config.APIKey,
    }
}

func (m *MyCustomModel) Generate(ctx context.Context, req *PonchoModelRequest) (*poncho.PonchoModelResponse, error) {
    // Custom implementation
    return &poncho.PonchoModelResponse{
        Message: &poncho.PonchoMessage{
            Role: poncho.PonchoRoleModel,
            Content: []*poncho.PonchoContentPart{
                {
                    Type: poncho.PonchoContentTypeText,
                    Text: "Custom model response",
                },
            },
        },
        Usage: &poncho.PonchoUsage{
            PromptTokens: 10,
            CompletionTokens: 5,
            TotalTokens: 15,
        },
    }, nil
}

func (m *MyCustomModel) SupportsStreaming() bool { return true }
func (m *MyCustomModel) SupportsTools() bool { return false }
func (m *MyCustomModel) SupportsVision() bool { return false }
func (m *MyCustomModel) SupportsSystemRole() bool { return true }
```

## Tool API

### PonchoTool Interface

```go
type PonchoTool interface {
    Name() string
    Description() string
    Version() string
    
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
    
    Validate(input interface{}) error
    
    Category() string
    Tags() []string
    Dependencies() []string
    
    Configure(config *PonchoToolConfig) error
    GetConfig() *PonchoToolConfig
}
```

### Custom Tool Implementation

```go
type ArticleImporterTool struct {
    *poncho.PonchoBaseTool
    s3Client *s3.Client
    storage  *storage.MemoryStorage
}

func NewArticleImporterTool(s3Client *s3.Client, storage *storage.MemoryStorage) *ArticleImporterTool {
    return &ArticleImporterTool{
        PonchoBaseTool: &poncho.PonchoBaseTool{
            name:        "articleImporter",
            description: "Imports article data from S3 storage",
            version:     "1.0.0",
            category:    "data-import",
            tags:        []string{"s3", "articles", "import"},
        },
        s3Client: s3Client,
        storage:  storage,
    }
}

func (t *ArticleImporterTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Parse input
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid input type")
    }
    
    articleID, ok := inputMap["article_id"].(string)
    if !ok {
        return nil, fmt.Errorf("article_id is required")
    }
    
    // Import logic
    data, err := t.s3Client.DownloadArticleData(ctx, articleID)
    if err != nil {
        return nil, fmt.Errorf("failed to download article: %w", err)
    }
    
    // Store in memory
    if err := t.storage.StoreArticle(articleID, data); err != nil {
        return nil, fmt.Errorf("failed to store article: %w", err)
    }
    
    return map[string]interface{}{
        "article_id": articleID,
        "status":     "imported",
        "data":       data,
    }, nil
}

func (t *ArticleImporterTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "article_id": map[string]interface{}{
                "type":        "string",
                "description": "Article identifier",
            },
            "include_images": map[string]interface{}{
                "type":        "boolean",
                "description": "Include images in import",
                "default":     false,
            },
        },
        "required": []string{"article_id"},
    }
}

func (t *ArticleImporterTool) OutputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "article_id": map[string]interface{}{
                "type": "string",
            },
            "status": map[string]interface{}{
                "type": "string",
            },
            "data": map[string]interface{}{
                "type": "object",
            },
        },
    }
}
```

## Flow API

### PonchoFlow Interface

```go
type PonchoFlow interface {
    Name() string
    Description() string
    Version() string
    
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    ExecuteStreaming(ctx context.Context, input interface{}, callback PonchoStreamCallback) error
    
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
    
    Category() string
    Tags() []string
    Dependencies() []string
    
    Configure(config *PonchoFlowConfig) error
    GetConfig() *PonchoFlowConfig
    
    Validate(input interface{}) error
}
```

### Custom Flow Implementation

```go
type ArticleProcessingFlow struct {
    *poncho.PonchoBaseFlow
    framework *poncho.PonchoFramework
}

func NewArticleProcessingFlow(framework *poncho.PonchoFramework) *ArticleProcessingFlow {
    return &ArticleProcessingFlow{
        PonchoBaseFlow: &poncho.PonchoBaseFlow{
            name:         "articleProcessor",
            description:  "Processes article data with AI analysis",
            version:      "1.0.0",
            category:     "processing",
            tags:         []string{"articles", "ai", "processing"},
            dependencies: []string{"articleImporter", "visionAnalyzer"},
            framework:    framework,
        },
    }
}

func (f *ArticleProcessingFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid input type")
    }
    
    articleID, ok := inputMap["article_id"].(string)
    if !ok {
        return nil, fmt.Errorf("article_id is required")
    }
    
    // Step 1: Import article data
    importResult, err := f.framework.ExecuteTool(ctx, "articleImporter", map[string]interface{}{
        "article_id": articleID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to import article: %w", err)
    }
    
    // Step 2: Analyze with AI
    analysisResult, err := f.framework.Generate(ctx, &poncho.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*poncho.PonchoMessage{
            {
                Role: poncho.PonchoRoleSystem,
                Content: []*poncho.PonchoContentPart{
                    {
                        Type: poncho.PonchoContentTypeText,
                        Text: "Ты - эксперт по анализу товаров для маркетплейсов.",
                    },
                },
            },
            {
                Role: poncho.PonchoRoleUser,
                Content: []*poncho.PonchoContentPart{
                    {
                        Type: poncho.PonchoContentTypeText,
                        Text: fmt.Sprintf("Проанализируй товар: %+v", importResult),
                    },
                },
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to analyze article: %w", err)
    }
    
    return map[string]interface{}{
        "article_id": articleID,
        "import":     importResult,
        "analysis":   analysisResult.Message.Content[0].Text,
        "status":     "processed",
    }, nil
}
```

## Prompt API

### Prompt Management

```go
type PonchoPromptManager interface {
    LoadPrompt(name string) (*PonchoPrompt, error)
    ExecutePrompt(ctx context.Context, name string, input map[string]interface{}) (interface{}, error)
    ListPrompts() []string
    RegisterPrompt(name string, prompt *PonchoPrompt) error
}

type PonchoPrompt struct {
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    Version      string                 `json:"version"`
    Model        string                 `json:"model"`
    Config       *PonchoModelConfig     `json:"config"`
    Template     string                 `json:"template"`
    InputSchema  map[string]interface{} `json:"input_schema"`
    OutputSchema map[string]interface{} `json:"output_schema"`
}
```

### Prompt Usage Example

```go
// Load and execute prompt
prompt, err := framework.GetPrompt("sketchAnalysis")
if err != nil {
    return fmt.Errorf("prompt not found: %w", err)
}

result, err := framework.ExecutePrompt(ctx, "sketchAnalysis", map[string]interface{}{
    "imageUrl":    "https://example.com/sketch.jpg",
    "customPrompt": "Проанализируй детали эскиза",
})
```

## Streaming API

### Streaming Callback

```go
type PonchoStreamCallback func(chunk *PonchoStreamChunk) error

type PonchoStreamChunk struct {
    Type      PonchoStreamChunkType  `json:"type"`
    Content   string                 `json:"content"`
    Delta     string                 `json:"delta,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Error     error                  `json:"error,omitempty"`
    Done      bool                   `json:"done"`
}
```

### Streaming Example

```go
// Streaming with progress tracking
var totalTokens int

err = framework.GenerateStreaming(ctx, &poncho.PonchoModelRequest{
    Model: "glm-chat",
    Messages: []*poncho.PonchoMessage{
        {
            Role: poncho.PonchoRoleUser,
            Content: []*poncho.PonchoContentPart{
                {
                    Type: poncho.PonchoContentTypeText,
                    Text: "Напиши подробное описание товара",
                },
            },
        },
    },
}, func(chunk *poncho.PonchoStreamChunk) error {
    switch chunk.Type {
    case poncho.PonchoStreamChunkStart:
        fmt.Println("Generation started...")
        
    case poncho.PonchoStreamChunkDelta:
        fmt.Print(chunk.Delta)
        totalTokens += len(strings.Fields(chunk.Delta))
        
    case poncho.PonchoStreamChunkEnd:
        fmt.Printf("\nGeneration completed. Total tokens: %d\n", totalTokens)
        
    case poncho.PonchoStreamChunkError:
        return fmt.Errorf("streaming error: %w", chunk.Error)
    }
    return nil
})
```

## Configuration API

### Configuration Structures

```go
type PonchoFrameworkConfig struct {
    Models map[string]*PonchoModelConfig `json:"models" yaml:"models"`
    Tools  map[string]*PonchoToolConfig  `json:"tools" yaml:"tools"`
    Flows  map[string]*PonchoFlowConfig  `json:"flows" yaml:"flows"`
    Prompts *PonchoPromptConfig          `json:"prompts" yaml:"prompts"`
    Logging *PonchoLoggingConfig         `json:"logging" yaml:"logging"`
    Metrics *PonchoMetricsConfig         `json:"metrics" yaml:"metrics"`
    Cache   *PonchoCacheConfig           `json:"cache" yaml:"cache"`
    Security *PonchoSecurityConfig        `json:"security" yaml:"security"`
}
```

### Configuration Example

```go
config := &poncho.PonchoFrameworkConfig{
    Models: map[string]*poncho.PonchoModelConfig{
        "deepseek-chat": {
            Provider:     "deepseek",
            ModelName:    "deepseek-chat",
            APIKey:       os.Getenv("DEEPSEEK_API_KEY"),
            Endpoint:     "https://api.deepseek.com/v1",
            MaxTokens:    4000,
            Temperature:  0.7,
            Timeout:      30 * time.Second,
            Supports: struct {
                Vision bool `json:"vision" yaml:"vision"`
                Tools  bool `json:"tools" yaml:"tools"`
                Stream bool `json:"stream" yaml:"stream"`
                System bool `json:"system" yaml:"system"`
            }{
                Vision: false,
                Tools:  true,
                Stream: true,
                System: true,
            },
        },
        "glm-vision": {
            Provider:     "zai",
            ModelName:    "glm-4.5v",
            APIKey:       os.Getenv("ZAI_API_KEY"),
            Endpoint:     "https://api.z.ai/v1",
            MaxTokens:    2000,
            Temperature:  0.5,
            Timeout:      60 * time.Second,
            Supports: struct {
                Vision bool `json:"vision" yaml:"vision"`
                Tools  bool `json:"tools" yaml:"tools"`
                Stream bool `json:"stream" yaml:"stream"`
                System bool `json:"system" yaml:"system"`
            }{
                Vision: true,
                Tools:  true,
                Stream: true,
                System: true,
            },
        },
    },
    Tools: map[string]*poncho.PonchoToolConfig{
        "articleImporter": {
            Enabled: true,
            Timeout: 30 * time.Second,
            Retry: &poncho.PonchoRetryConfig{
                MaxAttempts: 3,
                Backoff:     "exponential",
                BaseDelay:   1 * time.Second,
            },
        },
    },
    Logging: &poncho.PonchoLoggingConfig{
        Level:  "info",
        Format: "json",
        Output: "stdout",
    },
    Metrics: &poncho.PonchoMetricsConfig{
        Enabled: true,
        Interval: 30 * time.Second,
    },
}
```

## Migration Guide

### From GenKit to PonchoFramework

#### 1. Model Registration Migration

**GenKit:**
```go
// Old GenKit approach
model := genkit.DefineModel("deepseek", "chat", &ai.ModelDefinition{
    Provider: "deepseek",
    // ... configuration
})
```

**PonchoFramework:**
```go
// New PonchoFramework approach
config := &poncho.PonchoModelConfig{
    Provider:  "deepseek",
    ModelName: "deepseek-chat",
    APIKey:    os.Getenv("DEEPSEEK_API_KEY"),
}

framework := poncho.NewPonchoFramework(config)
framework.Start(ctx)
```

#### 2. Tool Definition Migration

**GenKit:**
```go
// Old GenKit approach
tool := genkit.DefineTool("importArticle", &ai.ToolDefinition{
    Description: "Import article data",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "article_id": {"type": "string"},
        },
    },
    Handler: func(ctx context.Context, input interface{}) (interface{}, error) {
        // Tool logic
        return result, nil
    },
})
```

**PonchoFramework:**
```go
// New PonchoFramework approach
type ArticleImporterTool struct {
    *poncho.PonchoBaseTool
    // Tool dependencies
}

func (t *ArticleImporterTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Tool logic
    return result, nil
}

tool := &ArticleImporterTool{
    PonchoBaseTool: &poncho.PonchoBaseTool{
        name:        "importArticle",
        description: "Import article data",
    },
}

framework.RegisterTool("importArticle", tool)
```

#### 3. Flow Migration

**GenKit:**
```go
// Old GenKit approach
flow := genkit.DefineFlow("articleProcessor", func(ctx context.Context, input interface{}) (interface{}, error) {
    // Flow logic
    return result, nil
})
```

**PonchoFramework:**
```go
// New PonchoFramework approach
type ArticleProcessingFlow struct {
    *poncho.PonchoBaseFlow
    framework *poncho.PonchoFramework
}

func (f *ArticleProcessingFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Flow logic
    return result, nil
}

flow := &ArticleProcessingFlow{
    PonchoBaseFlow: &poncho.PonchoBaseFlow{
        name:        "articleProcessor",
        description: "Process article data",
    },
    framework: framework,
}

framework.RegisterFlow("articleProcessor", flow)
```

#### 4. Generation Migration

**GenKit:**
```go
// Old GenKit approach
response, err := genkit.Generate(ctx, &ai.ModelRequest{
    Model: "deepseek/chat",
    Messages: []*ai.Message{
        ai.NewUserMessage("Analyze this article"),
    },
})
```

**PonchoFramework:**
```go
// New PonchoFramework approach
response, err := framework.Generate(ctx, &poncho.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*poncho.PonchoMessage{
        {
            Role: poncho.PonchoRoleUser,
            Content: []*poncho.PonchoContentPart{
                {
                    Type: poncho.PonchoContentTypeText,
                    Text: "Analyze this article",
                },
            },
        },
    },
})
```

## Error Handling

### Error Types

```go
type PonchoError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
    Cause   error                  `json:"cause,omitempty"`
}

func (e *PonchoError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *PonchoError) Unwrap() error {
    return e.Cause
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `MODEL_NOT_FOUND` | Модель не найдена |
| `TOOL_EXECUTION_FAILED` | Ошибка выполнения инструмента |
| `FLOW_VALIDATION_FAILED` | Ошибка валидации flow |
| `PROMPT_NOT_FOUND` | Промпт не найден |
| `CONFIGURATION_ERROR` | Ошибка конфигурации |
| `TIMEOUT` | Таймаут выполнения |
| `RATE_LIMIT_EXCEEDED` | Превышен лимит запросов |

### Error Handling Example

```go
response, err := framework.Generate(ctx, req)
if err != nil {
    var ponchoErr *poncho.PonchoError
    if errors.As(err, &ponchoErr) {
        switch ponchoErr.Code {
        case "MODEL_NOT_FOUND":
            log.Printf("Model not found: %s", ponchoErr.Message)
            // Fallback logic
        case "TIMEOUT":
            log.Printf("Request timeout: %s", ponchoErr.Message)
            // Retry logic
        default:
            log.Printf("Unknown error: %s", ponchoErr.Message)
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
    return err
}
```

## Best Practices

### 1. Configuration Management

```go
// Use environment variables for sensitive data
config := &poncho.PonchoFrameworkConfig{
    Models: map[string]*poncho.PonchoModelConfig{
        "deepseek-chat": {
            APIKey: os.Getenv("DEEPSEEK_API_KEY"),
        },
    },
}

// Validate configuration
if err := validateConfig(config); err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}
```

### 2. Context Management

```go
// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := framework.Generate(ctx, req)
```

### 3. Error Handling

```go
// Always handle errors properly
response, err := framework.Generate(ctx, req)
if err != nil {
    // Log error with context
    log.Printf("Generation failed for model %s: %v", req.Model, err)
    
    // Return structured error
    return fmt.Errorf("generation failed: %w", err)
}
```

### 4. Resource Management

```go
// Always cleanup resources
framework.Start(ctx)
defer framework.Stop(ctx)

// Use defer for cleanup
client := &http.Client{Timeout: 30 * time.Second}
defer client.CloseIdleConnections()
```

### 5. Testing

```go
// Mock interfaces for testing
type MockModel struct {
    *poncho.PonchoBaseModel
    response *poncho.PonchoModelResponse
    err      error
}

func (m *MockModel) Generate(ctx context.Context, req *poncho.PonchoModelRequest) (*poncho.PonchoModelResponse, error) {
    return m.response, m.err
}

// Use in tests
func TestArticleProcessing(t *testing.T) {
    mockModel := &MockModel{
        response: &poncho.PonchoModelResponse{
            Message: &poncho.PonchoMessage{
                Role: poncho.PonchoRoleModel,
                Content: []*poncho.PonchoContentPart{
                    {
                        Type: poncho.PonchoContentTypeText,
                        Text: "Test response",
                    },
                },
            },
        },
    }
    
    framework := poncho.NewPonchoFramework(config)
    framework.RegisterModel("test-model", mockModel)
    
    // Test logic
}
```

---

*Эта документация предоставляет полное руководство по использованию PonchoFramework и миграции с GenKit.*