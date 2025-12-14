# PonchoAiFramework API Reference

## Table of Contents

1. [Core Interfaces](#core-interfaces)
2. [Model API](#model-api)
3. [Prompt System API](#prompt-system-api)
4. [Tools API](#tools-api)
5. [Flows API](#flows-api)
6. [Wildberries API](#wildberries-api)
7. [Configuration API](#configuration-api)
8. [Error Handling](#error-handling)
9. [Types and Enums](#types-and-enums)

---

## Core Interfaces

### PonchoFramework

Main framework interface providing access to all framework capabilities.

```go
type PonchoFramework interface {
    // Lifecycle management
    Start(ctx context.Context) error
    Stop(ctx context.Context) error

    // Model management
    RegisterModel(name string, model PonchoModel) error
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error

    // Tool management
    RegisterTool(name string, tool PonchoTool) error
    ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)

    // Flow management
    RegisterFlow(name string, flow PonchoFlow) error
    ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error)
    ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback PonchoStreamCallback) error

    // Registries access
    GetModelRegistry() PonchoModelRegistry
    GetToolRegistry() PonchoToolRegistry
    GetFlowRegistry() PonchoFlowRegistry

    // Configuration and monitoring
    GetConfig() *PonchoFrameworkConfig
    ReloadConfig(ctx context.Context) error
    Health(ctx context.Context) (*PonchoHealthStatus, error)
    Metrics(ctx context.Context) (*PonchoMetrics, error)
}
```

**Methods:**

#### Start(ctx context.Context) error
Starts the framework and initializes all registered components.

**Parameters:**
- `ctx context.Context`: Context for the operation

**Returns:**
- `error`: Error if startup fails

**Example:**
```go
fw := core.NewPonchoFramework()
if err := fw.LoadFromFile("config.yaml"); err != nil {
    return err
}

ctx := context.Background()
if err := fw.Start(ctx); err != nil {
    return err
}
defer fw.Stop(ctx)
```

#### Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
Generates text using the specified model.

**Parameters:**
- `ctx context.Context`: Context for the operation
- `req *PonchoModelRequest`: Generation request

**Returns:**
- `*PonchoModelResponse`: Model response with generated content
- `error`: Error if generation fails

**Example:**
```go
request := &interfaces.PonchoModelRequest{
    Model: "deepseek-chat",
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

response, err := fw.Generate(ctx, request)
if err != nil {
    return err
}

fmt.Printf("Response: %s\n", response.Message.Content[0].Text)
```

### PonchoModel

Interface for AI model implementations.

```go
type PonchoModel interface {
    // Identification
    Name() string
    Provider() string
    Capabilities() *ModelCapabilities

    // Lifecycle
    Initialize(ctx context.Context, config map[string]interface{}) error
    Health(ctx context.Context) error

    // Generation
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error

    // Configuration
    GetConfig() map[string]interface{}
    UpdateConfig(config map[string]interface{}) error

    // Metrics
    GetMetrics() *ModelMetrics
    ResetMetrics()
}
```

**Methods:**

#### Capabilities() *ModelCapabilities
Returns the model's capabilities.

**Returns:**
- `*ModelCapabilities`: Model capabilities including supported features

**ModelCapabilities:**
```go
type ModelCapabilities struct {
    Streaming bool   `json:"streaming"` // Supports streaming responses
    Tools     bool   `json:"tools"`     // Supports function calling
    Vision    bool   `json:"vision"`    // Supports image analysis
    System    bool   `json:"system"`    // Supports system messages
    JSONMode  bool   `json:"json_mode"` // Supports JSON mode
    MaxTokens int    `json:"max_tokens"` // Maximum token limit
}
```

---

## Model API

### PonchoModelRequest

Request structure for model generation.

```go
type PonchoModelRequest struct {
    Model       string           `json:"model"`                 // Model identifier
    Messages    []*PonchoMessage `json:"messages"`              // Conversation messages
    MaxTokens   *int             `json:"max_tokens,omitempty"`  // Maximum tokens to generate
    Temperature *float32         `json:"temperature,omitempty"` // Sampling temperature (0.0-2.0)
    TopP        *float32         `json:"top_p,omitempty"`       // Nucleus sampling parameter
    Stop        []string         `json:"stop,omitempty"`        // Stop sequences
    Stream      bool             `json:"stream,omitempty"`      // Enable streaming
    Tools       []*PonchoTool    `json:"tools,omitempty"`       // Tools to use
    ToolChoice  interface{}      `json:"tool_choice,omitempty"` // Tool choice strategy
}
```

### PonchoModelResponse

Response structure from model generation.

```go
type PonchoModelResponse struct {
    ID      string         `json:"id"`       // Response identifier
    Object  string         `json:"object"`   // Response object type
    Created int64          `json:"created"`  // Creation timestamp
    Model   string         `json:"model"`    // Model used
    Message *PonchoMessage `json:"message"`  // Response message
    Usage   *UsageInfo     `json:"usage"`    // Token usage information
}

type UsageInfo struct {
    PromptTokens     int `json:"prompt_tokens"`     // Tokens in prompt
    CompletionTokens int `json:"completion_tokens"` // Tokens generated
    TotalTokens      int `json:"total_tokens"`      // Total tokens used
}
```

### PonchoMessage

Message structure in conversations.

```go
type PonchoMessage struct {
    Role      string              `json:"role"`       // Message role (system, user, assistant, tool)
    Content   []*PonchoContentPart `json:"content"`   // Message content
    ToolCalls []*PonchoToolCall   `json:"tool_calls,omitempty"` // Tool calls made
    ToolCallID string              `json:"tool_call_id,omitempty"` // ID of tool call being responded to
}
```

### PonchoContentPart

Content part within a message.

```go
type PonchoContentPart struct {
    Type     string            `json:"type"`       // Content type (text, image_url)
    Text     string            `json:"text,omitempty"` // Text content
    ImageURL *PonchoImageURL   `json:"image_url,omitempty"` // Image URL for vision
}

type PonchoImageURL struct {
    URL    string `json:"url"`     // Image URL (can be base64 data URL)
    Detail string `json:"detail"`  // Image detail level (low, high, auto)
}
```

**Example Usage:**

```go
// Text-only message
textMessage := &interfaces.PonchoMessage{
    Role: interfaces.PonchoRoleUser,
    Content: []*interfaces.PonchoContentPart{
        {
            Type: interfaces.PonchoContentTypeText,
            Text: "Analyze this fashion item",
        },
    },
}

// Multimodal message with text and image
multimodalMessage := &interfaces.PonchoMessage{
    Role: interfaces.PonchoRoleUser,
    Content: []*interfaces.PonchoContentPart{
        {
            Type: interfaces.PonchoContentTypeText,
            Text: "What do you see in this image?",
        },
        {
            Type: interfaces.PonchoContentTypeImageURL,
            ImageURL: &interfaces.PonchoImageURL{
                URL:    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ...",
                Detail: "high",
            },
        },
    },
}
```

---

## Prompt System API

### PromptManager

Interface for managing prompt templates.

```go
type PromptManager interface {
    // Template management
    LoadTemplate(name string) (*PromptTemplate, error)
    ListTemplates() ([]string, error)
    SaveTemplate(name string, template *PromptTemplate) error
    DeleteTemplate(name string) error

    // Execution
    ExecutePrompt(ctx context.Context, templateName string, variables map[string]interface{}, modelOverride string) (*PonchoModelResponse, error)
    ExecutePromptStreaming(ctx context.Context, templateName string, variables map[string]interface{}, modelOverride string, callback PonchoStreamCallback) error

    // Validation
    ValidateTemplate(template *PromptTemplate) error

    // Configuration
    GetConfig() *PromptConfig
    UpdateConfig(config *PromptConfig) error
}
```

### PromptTemplate

Template structure for prompts.

```go
type PromptTemplate struct {
    Name        string          `json:"name"`        // Template name
    Description string          `json:"description"` // Template description
    Parts       []*PromptPart   `json:"parts"`       // Template parts
    Variables   []PromptVariable `json:"variables"`  // Defined variables
    Metadata    map[string]interface{} `json:"metadata"` // Additional metadata
    CreatedAt   time.Time       `json:"created_at"`  // Creation time
    UpdatedAt   time.Time       `json:"updated_at"`  // Last update time
}

type PromptPart struct {
    Type    string `json:"type"`    // Part type (system, user, assistant, tool)
    Content string `json:"content"` // Part content with template variables
}

type PromptVariable struct {
    Name         string      `json:"name"`          // Variable name
    Type         string      `json:"type"`          // Variable type (string, number, boolean, object, array)
    Required     bool        `json:"required"`      // Whether variable is required
    DefaultValue interface{} `json:"default_value"` // Default value if not provided
    Description  string      `json:"description"`   // Variable description
}
```

### Template Syntax

Prompts use Handlebars syntax for variable substitution:

```handlebars
{{role "config"}}
model: glm-4.6v-flash
config:
  temperature: {{temperature}}
  maxOutputTokens: {{max_tokens}}

{{role "system"}}
You are a {{role_description}}.

{{role "user"}}
{{#if is_fashion}}
Please analyze this fashion item.
{{else}}
Please analyze this image.
{{/if}}

{{media url=image_url}}

{{#each additional_info}}
- {{name}}: {{value}}
{{/each}}
```

**Template Variables:**

- Simple variables: `{{variable_name}}`
- Conditional blocks: `{{#if condition}}...{{/if}}`
- Loops: `{{#each array}}...{{/each}}`
- Media insertion: `{{media url=image_url}}`
- Role specification: `{{role "config"}}` or `{{role "system"}}`

**Example Usage:**

```go
// Create a prompt template
template := &interfaces.PromptTemplate{
    Name:        "fashion_analysis",
    Description: "Analyzes fashion items",
    Parts: []*interfaces.PromptPart{
        {
            Type:    interfaces.PromptPartTypeSystem,
            Content: "You are a fashion expert. Analyze this {{item_type}}.",
        },
        {
            Type:    interfaces.PromptPartTypeUser,
            Content: "{{media url=image_url}}\n\nDescribe the key features.",
        },
    },
    Variables: []interfaces.PromptVariable{
        {
            Name:     "item_type",
            Type:     "string",
            Required: true,
        },
        {
            Name:     "image_url",
            Type:     "string",
            Required: true,
        },
    },
}

// Save template
err := promptManager.SaveTemplate("fashion_analysis", template)
if err != nil {
    return err
}

// Execute prompt with variables
variables := map[string]interface{}{
    "item_type": "dress",
    "image_url": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ...",
}

response, err := promptManager.ExecutePrompt(ctx, "fashion_analysis", variables, "")
if err != nil {
    return err
}

fmt.Printf("Analysis: %s\n", response.Message.Content[0].Text)
```

---

## Tools API

### PonchoTool

Interface for tool implementations.

```go
type PonchoTool interface {
    // Identification
    Name() string
    Description() string
    Category() string

    // Schema
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}

    // Execution
    Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)

    // Configuration
    GetConfig() map[string]interface{}
    UpdateConfig(config map[string]interface{}) error

    // Health and metrics
    Health(ctx context.Context) error
    GetMetrics() *ToolMetrics
}
```

### Tool Registration

```go
// Implement a custom tool
type WeatherTool struct {
    apiKey string
    client *http.Client
}

func (wt *WeatherTool) Name() string {
    return "weather"
}

func (wt *WeatherTool) Description() string {
    return "Gets current weather information for a location"
}

func (wt *WeatherTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "City name or coordinates",
            },
            "units": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"metric", "imperial"},
                "description": "Temperature units",
            },
        },
        "required": []string{"location"},
    }
}

func (wt *WeatherTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
    location := input["location"].(string)
    units := "metric"
    if u, ok := input["units"].(string); ok {
        units = u
    }

    // Call weather API
    weather, err := wt.getWeather(ctx, location, units)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "location":    location,
        "temperature": weather.Temperature,
        "humidity":    weather.Humidity,
        "description": weather.Description,
        "units":       units,
    }, nil
}

// Register the tool
weatherTool := &WeatherTool{
    apiKey: os.Getenv("WEATHER_API_KEY"),
    client: &http.Client{Timeout: 10 * time.Second},
}

err := fw.RegisterTool("weather", weatherTool)
if err != nil {
    log.Fatal(err)
}
```

### Tool Usage in Requests

```go
// Use tool in model request
request := &interfaces.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*interfaces.PonchoMessage{
        {
            Role: interfaces.PonchoRoleUser,
            Content: []*interfaces.PonchoContentPart{
                {
                    Type: interfaces.PonchoContentTypeText,
                    Text: "What's the weather like in Moscow?",
                },
            },
        },
    },
    Tools: []*interfaces.PonchoTool{
        {
            Type: "function",
            Function: &interfaces.PonchoFunction{
                Name:        "weather",
                Description: "Gets current weather information",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type": "string",
                        },
                        "units": map[string]interface{}{
                            "type": "string",
                            "enum": []string{"metric", "imperial"},
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
}

response, err := fw.Generate(ctx, request)
if err != nil {
    return err
}

// Handle tool calls
if len(response.Message.ToolCalls) > 0 {
    for _, toolCall := range response.Message.ToolCalls {
        // Execute tool
        result, err := fw.ExecuteTool(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
        if err != nil {
            log.Printf("Tool execution error: %v", err)
            continue
        }

        fmt.Printf("Tool result: %+v\n", result)
    }
}
```

---

## Flows API

### PonchoFlow

Interface for flow implementations.

```go
type PonchoFlow interface {
    // Identification
    Name() string
    Description() string
    Category() string

    // Schema
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}

    // Execution
    Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
    ExecuteStreaming(ctx context.Context, input map[string]interface{}, callback PonchoStreamCallback) error

    // Dependencies
    GetDependencies() []string // Names of required models/tools

    // Configuration
    GetConfig() map[string]interface{}
    UpdateConfig(config map[string]interface{}) error

    // Health and metrics
    Health(ctx context.Context) error
    GetMetrics() *FlowMetrics
}
```

### Flow Implementation Example

```go
type FashionAnalysisFlow struct {
    name         string
    fw           interfaces.PonchoFramework
    promptMgr    interfaces.PromptManager
    dependencies []string
}

func (faf *FashionAnalysisFlow) Name() string {
    return faf.name
}

func (faf *FashionAnalysisFlow) Description() string {
    return "Analyzes fashion items using vision and text models"
}

func (faf *FashionAnalysisFlow) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "image_url": map[string]interface{}{
                "type":        "string",
                "description": "Base64 encoded fashion image",
            },
            "analysis_type": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"description", "characteristics", "creative"},
                "description": "Type of analysis to perform",
            },
        },
        "required": []string{"image_url"},
    }
}

func (faf *FashionAnalysisFlow) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
    imageURL := input["image_url"].(string)
    analysisType, _ := input["analysis_type"].(string)
    if analysisType == "" {
        analysisType = "description"
    }

    // Step 1: Vision analysis
    variables := map[string]interface{}{
        "photoUrl": imageURL,
    }

    promptName := fmt.Sprintf("sketch_%s", analysisType)
    visionResponse, err := faf.promptMgr.ExecutePrompt(ctx, promptName, variables, "glm-4.6v-flash")
    if err != nil {
        return nil, err
    }

    // Step 2: Post-processing if needed
    if analysisType == "creative" {
        enhanceRequest := &interfaces.PonchoModelRequest{
            Model: "deepseek-chat",
            Messages: []*interfaces.PonchoMessage{
                {
                    Role: interfaces.PonchoRoleUser,
                    Content: []*interfaces.PonchoContentPart{
                        {
                            Type: interfaces.PonchoContentTypeText,
                            Text: fmt.Sprintf("Make this description more appealing: %s", visionResponse.Message.Content[0].Text),
                        },
                    },
                },
            },
            MaxTokens:   &[]int{300}[0],
            Temperature: &[]float32{0.8}[0],
        }

        enhancedResponse, err := faf.fw.Generate(ctx, enhanceRequest)
        if err != nil {
            // Return original if enhancement fails
            return visionResponse.Message.Content[0].Text, nil
        }

        return enhancedResponse.Message.Content[0].Text, nil
    }

    return visionResponse.Message.Content[0].Text, nil
}

func (faf *FashionAnalysisFlow) GetDependencies() []string {
    return []string{"glm-4.6v-flash", "deepseek-chat"}
}

// Register the flow
fashionFlow := &FashionAnalysisFlow{
    name:         "fashion_analysis",
    fw:           fw,
    promptMgr:    fw.GetPromptManager(),
    dependencies: []string{"glm-4.6v-flash", "deepseek-chat"},
}

err := fw.RegisterFlow("fashion_analysis", fashionFlow)
if err != nil {
    log.Fatal(err)
}
```

---

## Wildberries API

### WildberriesClient

Client for interacting with Wildberries marketplace API.

```go
type WildberriesClient interface {
    // Product information
    GetProductInfo(ctx context.Context, nmID string) (*ProductInfoResponse, error)
    GetCharacteristics(ctx context.Context, nmID string) (*CharacteristicsResponse, error)
    GetDescription(ctx context.Context, nmID string) (*DescriptionResponse, error)

    // Categories
    GetCategories(ctx context.Context) (*CategoriesResponse, error)
    GetCategoryTree(ctx context.Context) (*CategoryTreeResponse, error)

    // Search
    SearchProducts(ctx context.Context, query string, filters *SearchFilters) (*SearchResponse, error)

    // Analytics
    GetSalesData(ctx context.Context, nmID string, period *TimePeriod) (*SalesDataResponse, error)
    GetReviews(ctx context.Context, nmID string, pagination *Pagination) (*ReviewsResponse, error)

    // Content management
    UploadImage(ctx context.Context, nmID string, imageData []byte) (*UploadResponse, error)
    UpdateDescription(ctx context.Context, nmID string, description string) error

    // Health and metrics
    Health(ctx context.Context) error
    GetMetrics() *WildberriesMetrics
}
```

### FashionAnalyzer

Specialized analyzer for fashion items using Wildberries data.

```go
type FashionAnalyzer interface {
    // Analysis methods
    AnalyzeFashionItem(ctx context.Context, imageURL, description string) (*FashionAnalysis, error)
    AnalyzeSketch(ctx context.Context, imageURL string) (*SketchAnalysis, error)
    GenerateProductContent(ctx context.Context, item *FashionItem) (*ProductContent, error)

    // Market insights
    GetMarketTrends(ctx context.Context, category string) (*MarketTrends, error)
    AnalyzeCompetition(ctx context.Context, item *FashionItem) (*CompetitionAnalysis, error)

    // Recommendations
    GetStyleRecommendations(ctx context.Context, item *FashionItem) ([]*StyleRecommendation, error)
    GetSimilarItems(ctx context.Context, nmID string) ([]*SimilarItem, error)
}

type FashionAnalysis struct {
    Category       string                 `json:"category"`        // Product category
    Attributes     map[string]interface{} `json:"attributes"`      // Extracted attributes
    Style          string                 `json:"style"`           // Style classification
    Season         string                 `json:"season"`          // Season suitability
    Occasion       string                 `json:"occasion"`        // Suitable occasions
    Materials      []string               `json:"materials"`       // Identified materials
    Colors         []string               `json:"colors"`          // Identified colors
    Features       []string               `json:"features"`        // Key features
    MarketData     *MarketData            `json:"market_data"`     // Market insights
    Confidence     float64                `json:"confidence"`      // Analysis confidence
}
```

### Usage Example

```go
// Create Wildberries client
wbClient := wildberries.NewWildberriesClient(apiKey, &wildberries.Config{
    Timeout:    30 * time.Second,
    BaseURL:    "https://content-api.wildberries.ru",
    RateLimit:  &wildberries.RateLimitConfig{
        RequestsPerMinute: 100,
        BurstSize:        20,
    },
})

// Create fashion analyzer
analyzer := wildberries.NewFashionAnalyzer(wbClient, fw)

// Analyze fashion item
imageURL := "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ..."
description := "Элегантное вечернее платье из шелка"

analysis, err := analyzer.AnalyzeFashionItem(ctx, imageURL, description)
if err != nil {
    return err
}

fmt.Printf("Category: %s\n", analysis.Category)
fmt.Printf("Style: %s\n", analysis.Style)
fmt.Printf("Season: %s\n", analysis.Season)
fmt.Printf("Materials: %v\n", analysis.Materials)

// Get market trends
trends, err := analyzer.GetMarketTrends(ctx, "платья")
if err != nil {
    return err
}

for _, trend := range trends.Trends {
    fmt.Printf("Trend: %s (Popularity: %.1f%%)\n", trend.Name, trend.Popularity)
}
```

---

## Configuration API

### PonchoFrameworkConfig

Main configuration structure.

```go
type PonchoFrameworkConfig struct {
    Models map[string]*ModelConfig `yaml:"models" json:"models"`
    Tools  map[string]*ToolConfig  `yaml:"tools" json:"tools"`
    Flows  map[string]*FlowConfig  `yaml:"flows" json:"flows"`

    Logging *LoggingConfig   `yaml:"logging" json:"logging"`
    Cache   *CacheConfig     `yaml:"cache" json:"cache"`
    Metrics *MetricsConfig   `yaml:"metrics" json:"metrics"`
    Security *SecurityConfig `yaml:"security" json:"security"`

    S3           *S3Config           `yaml:"s3" json:"s3"`
    Wildberries  *WildberriesConfig  `yaml:"wildberries" json:"wildberries"`
}

type ModelConfig struct {
    Provider     string                  `yaml:"provider" json:"provider"`
    ModelName    string                  `yaml:"model_name" json:"model_name"`
    APIKey       string                  `yaml:"api_key" json:"api_key"`
    BaseURL      string                  `yaml:"base_url" json:"base_url"`
    MaxTokens    int                     `yaml:"max_tokens" json:"max_tokens"`
    Temperature  float32                 `yaml:"temperature" json:"temperature"`
    TopP         float32                 `yaml:"top_p" json:"top_p"`
    Timeout      string                  `yaml:"timeout" json:"timeout"`
    Supports     *ModelCapabilities      `yaml:"supports" json:"supports"`
    Retry        *RetryConfig            `yaml:"retry" json:"retry"`
    CustomParams map[string]interface{}   `yaml:"custom_params" json:"custom_params"`
}
```

### Configuration Management

```go
// Load configuration from file
err := fw.LoadFromFile("config.yaml")
if err != nil {
    return err
}

// Load configuration from environment
err := fw.LoadFromEnv()
if err != nil {
    return err
}

// Reload configuration
err := fw.ReloadConfig(ctx)
if err != nil {
    return err
}

// Get current configuration
config := fw.GetConfig()
fmt.Printf("Loaded %d models\n", len(config.Models))
```

### Environment Variable Substitution

Configuration supports environment variable substitution:

```yaml
models:
  deepseek-chat:
    api_key: "${DEEPSEEK_API_KEY}"
    timeout: "${TIMEOUT:30s}"  # With default value

  glm-4.6v-flash:
    api_key: "${ZAI_API_KEY}"
    base_url: "${ZAI_BASE_URL:https://open.bigmodel.cn/api/paas/v4}"
```

---

## Error Handling

### Error Types

```go
// Common error types
var (
    ErrModelNotFound       = errors.New("model not found")
    ErrToolNotFound        = errors.New("tool not found")
    ErrFlowNotFound        = errors.New("flow not found")
    ErrInvalidRequest      = errors.New("invalid request")
    ErrUnauthorized        = errors.New("unauthorized")
    ErrRateLimited         = errors.New("rate limited")
    ErrTimeout             = errors.New("request timeout")
    ErrInvalidResponse     = errors.New("invalid response")
)

// Model-specific errors
type ModelError struct {
    Model    string `json:"model"`
    Code     string `json:"code"`
    Message  string `json:"message"`
    Details  map[string]interface{} `json:"details,omitempty"`
}

func (e *ModelError) Error() string {
    return fmt.Sprintf("model '%s' error (%s): %s", e.Model, e.Code, e.Message)
}

// Tool execution errors
type ToolExecutionError struct {
    Tool     string `json:"tool"`
    Input    interface{} `json:"input"`
    Cause    error  `json:"cause"`
}

func (e *ToolExecutionError) Error() string {
    return fmt.Sprintf("tool '%s' execution error: %v", e.Tool, e.Cause)
}
```

### Error Handling Patterns

```go
// Handle model errors
response, err := fw.Generate(ctx, request)
if err != nil {
    var modelErr *interfaces.ModelError
    if errors.As(err, &modelErr) {
        switch modelErr.Code {
        case "rate_limited":
            // Implement backoff
            time.Sleep(time.Second)
            // Retry request
        case "invalid_api_key":
            // Log and alert
            log.Printf("Invalid API key for model %s", modelErr.Model)
        case "context_length_exceeded":
            // Truncate prompt and retry
            request.MaxTokens = &[]int{request.MaxTokens / 2}[0]
        }
    }
    return err
}

// Handle tool errors
result, err := fw.ExecuteTool(ctx, toolName, input)
if err != nil {
    var toolErr *interfaces.ToolExecutionError
    if errors.As(err, &toolErr) {
        log.Printf("Tool '%s' failed with input %+v: %v",
            toolErr.Tool, toolErr.Input, toolErr.Cause)
    }
    return err
}
```

---

## Types and Enums

### Role Types

```go
const (
    PonchoRoleSystem    = "system"
    PonchoRoleUser      = "user"
    PonchoRoleAssistant = "assistant"
    PonchoRoleTool      = "tool"
)
```

### Content Types

```go
const (
    PonchoContentTypeText      = "text"
    PonchoContentTypeImageURL  = "image_url"
)
```

### Prompt Part Types

```go
const (
    PromptPartTypeConfig    = "config"
    PromptPartTypeSystem    = "system"
    PromptPartTypeUser      = "user"
    PromptPartTypeAssistant = "assistant"
    PromptPartTypeTool      = "tool"
)
```

### Model Providers

```go
const (
    ProviderDeepSeek = "deepseek"
    ProviderZAI      = "zai"
    ProviderOpenAI   = "openai"
    ProviderAnthropic = "anthropic"
)
```

### Tool Categories

```go
const (
    ToolCategoryGeneral     = "general"
    ToolCategoryData        = "data"
    ToolCategoryAPI         = "api"
    ToolCategoryAnalysis    = "analysis"
    ToolCategoryFashion     = "fashion"
    ToolCategoryECommerce   = "ecommerce"
)
```

### Flow Categories

```go
const (
    FlowCategoryGeneration  = "generation"
    FlowCategoryAnalysis    = "analysis"
    FlowCategoryProcessing  = "processing"
    FlowCategoryWorkflow    = "workflow"
    FlowCategoryFashion     = "fashion"
)
```

---

## Streaming API

### Stream Callback

```go
type PonchoStreamCallback func(*PonchoModelResponse) error
```

### Streaming Usage

```go
// Model streaming
err := fw.GenerateStreaming(ctx, request, func(chunk *PonchoModelResponse) error {
    if chunk.Message != nil && len(chunk.Message.Content) > 0 {
        text := chunk.Message.Content[0].Text
        fmt.Print(text) // Stream output

        // Handle streaming errors
        if chunk.Error != nil {
            return chunk.Error
        }
    }
    return nil
})

// Flow streaming
err := fw.ExecuteFlowStreaming(ctx, "fashion_analysis", input, func(chunk *PonchoModelResponse) error {
    if chunk.Message != nil {
        // Process streaming chunk
        fmt.Printf("Chunk: %s\n", chunk.Message.Content[0].Text)
    }
    return nil
})

// Prompt streaming
err := promptManager.ExecutePromptStreaming(ctx, "fashion_analysis", variables, "", func(chunk *PonchoModelResponse) error {
    // Handle prompt streaming response
    return nil
})
```

---

## Metrics API

### Framework Metrics

```go
type PonchoMetrics struct {
    GeneratedRequests *GenerationMetrics `json:"generated_requests"`
    ToolExecutions    *ToolMetrics       `json:"tool_executions"`
    FlowExecutions    *FlowMetrics       `json:"flow_executions"`
    Errors            *ErrorMetrics      `json:"errors"`
    System            *SystemMetrics     `json:"system"`
    Timestamp         time.Time          `json:"timestamp"`
}

type GenerationMetrics struct {
    TotalRequests int64                        `json:"total_requests"`
    SuccessCount  int64                        `json:"success_count"`
    ErrorCount    int64                        `json:"error_count"`
    ByModel       map[string]*ModelMetrics     `json:"by_model"`
}

type ModelMetrics struct {
    RequestCount    int64   `json:"request_count"`
    SuccessCount    int64   `json:"success_count"`
    ErrorCount      int64   `json:"error_count"`
    TotalTokens     int64   `json:"total_tokens"`
    AvgResponseTime float64 `json:"avg_response_time"`
}
```

### Getting Metrics

```go
// Get framework metrics
metrics, err := fw.Metrics(ctx)
if err != nil {
    return err
}

fmt.Printf("Total requests: %d\n", metrics.GeneratedRequests.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n",
    float64(metrics.GeneratedRequests.SuccessCount)/float64(metrics.GeneratedRequests.TotalRequests)*100)

// Get model-specific metrics
for modelName, modelMetrics := range metrics.GeneratedRequests.ByModel {
    fmt.Printf("Model %s: %d requests, %.2fms avg\n",
        modelName, modelMetrics.RequestCount, modelMetrics.AvgResponseTime)
}
```

---

## Health Check API

### Health Status

```go
type PonchoHealthStatus struct {
    Status     string                       `json:"status"`
    Timestamp  time.Time                    `json:"timestamp"`
    Version    string                       `json:"version"`
    Components map[string]*ComponentHealth  `json:"components"`
    Uptime     time.Duration                `json:"uptime"`
}

type ComponentHealth struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### Health Check Usage

```go
// Check framework health
health, err := fw.Health(ctx)
if err != nil {
    return err
}

if health.Status != "healthy" {
    for componentName, componentHealth := range health.Components {
        if componentHealth.Status != "healthy" {
            log.Printf("Component %s unhealthy: %s",
                componentName, componentHealth.Message)
        }
    }
}

fmt.Printf("Framework status: %s (uptime: %v)\n", health.Status, health.Uptime)
```

This comprehensive API reference provides detailed documentation for all aspects of the PonchoAiFramework, making it easy for developers to understand and use the framework effectively.