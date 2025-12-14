# PonchoAiFramework User Guide

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration](#configuration)
3. [Model Usage](#model-usage)
4. [Prompt System](#prompt-system)
5. [Tools Integration](#tools-integration)
6. [Wildberries API](#wildberries-api)
7. [Flow Orchestration](#flow-orchestration)
8. [Fashion Industry Features](#fashion-industry-features)
9. [Testing](#testing)
10. [Production Deployment](#production-deployment)

---

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/ilkoid/PonchoAiFramework.git
cd PonchoAiFramework

# Install dependencies
go mod tidy

# Build the framework
go build -o poncho-framework ./cmd/poncho-framework
```

### Basic Setup

1. **Set Environment Variables**

```bash
export DEEPSEEK_API_KEY="your_deepseek_api_key"
export ZAI_API_KEY="your_zai_api_key"
export S3_ACCESS_KEY="your_s3_access_key"
export S3_SECRET_KEY="your_s3_secret_key"
export WB_API_CONTENT_KEY="your_wildberries_api_key"
```

2. **Create Configuration**

```yaml
# config.yaml
models:
  deepseek-chat:
    provider: deepseek
    model_name: deepseek-chat
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com/v1"
    max_tokens: 4000
    temperature: 0.7
    timeout: "30s"

  glm-4.6v-flash:
    provider: zai
    model_name: glm-4.6v-flash
    api_key: "${ZAI_API_KEY}"
    base_url: "https://open.bigmodel.cn/api/paas/v4"
    max_tokens: 1500
    temperature: 0.3
    timeout: "30s"
    supports:
      vision: true
      json_mode: true

tools:
  wildberries:
    type: wildberries
    config:
      api_key: "${WB_API_CONTENT_KEY}"
      timeout: 30
      rate_limit:
        requests_per_minute: 100

logging:
  level: "info"
  format: "json"

cache:
  type: "memory"
  ttl: "1h"
  max_size: 100
```

3. **Run Your First Request**

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ilkoid/PonchoAiFramework/core"
    "github.com/ilkoid/PonchoAiFramework/interfaces"
    "github.com/ilkoid/PonchoAiFramework/models/deepseek"
    "github.com/ilkoid/PonchoAiFramework/models/zai"
)

func main() {
    ctx := context.Background()

    // Create framework
    fw := core.NewPonchoFramework()

    // Load configuration
    if err := fw.LoadFromFile("config.yaml"); err != nil {
        log.Fatal(err)
    }

    // Start framework
    if err := fw.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer fw.Stop(ctx)

    // Generate text
    request := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: "Опиши модное платье для летнего сезона",
                    },
                },
            },
        },
    }

    response, err := fw.Generate(ctx, request)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Message.Content[0].Text)
}
```

---

## Configuration

### Environment Variables

The framework supports environment variable substitution using `${VARIABLE_NAME}` syntax:

```yaml
models:
  my-model:
    api_key: "${API_KEY}"  # Will be replaced with actual env variable
    timeout: "${TIMEOUT:30s}"  # With default value
```

### Model Configuration

Each model requires specific configuration:

```yaml
models:
  # DeepSeek Model
  deepseek-chat:
    provider: deepseek
    model_name: deepseek-chat
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com/v1"
    max_tokens: 4000
    temperature: 0.7
    timeout: "30s"
    supports:
      streaming: true
      tools: true
      json_mode: true

  # Z.AI GLM Model with Vision
  glm-4.6v-flash:
    provider: zai
    model_name: glm-4.6v-flash
    api_key: "${ZAI_API_KEY}"
    base_url: "https://open.bigmodel.cn/api/paas/v4"
    max_tokens: 1500
    temperature: 0.3
    timeout: "30s"
    supports:
      streaming: true
      tools: true
      vision: true
      system: true
      json_mode: true
    custom_params:
      top_p: 0.9
```

### Tool Configuration

```yaml
tools:
  # Wildberries Integration
  wildberries:
    type: wildberries
    config:
      api_key: "${WB_API_CONTENT_KEY}"
      timeout: 30
      base_url: "https://content-api.wildberries.ru"
      rate_limit:
        requests_per_minute: 100
        burst_size: 20

  # S3 Storage
  s3_storage:
    type: s3
    config:
      access_key: "${S3_ACCESS_KEY}"
      secret_key: "${S3_SECRET_KEY}"
      url: "https://storage.yandexcloud.net"
      region: "ru-central1"
      bucket: "plm-ai"
```

---

## Model Usage

### Text Generation with DeepSeek

```go
// Simple text generation
request := &interfaces.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*interfaces.PonchoMessage{
        {
            Role: interfaces.PonchoRoleUser,
            Content: []*interfaces.PonchoContentPart{
                {
                    Type: interfaces.PonchoContentTypeText,
                    Text: "Напиши описание модной куртки для осени",
                },
            },
        },
    },
    MaxTokens: &[]int{500}[0],
    Temperature: &[]float32{0.8}[0],
}

response, err := fw.Generate(ctx, request)
if err != nil {
    return err
}

fmt.Printf("Generated text: %s\n", response.Message.Content[0].Text)
```

### Vision Analysis with GLM-4.6V

```go
// Analyze fashion sketch
request := &interfaces.PonchoModelRequest{
    Model: "glm-4.6v-flash",
    Messages: []*interfaces.PonchoMessage{
        {
            Role: interfaces.PonchoRoleUser,
            Content: []*interfaces.PonchoContentPart{
                {
                    Type: interfaces.PonchoContentTypeText,
                    Text: "Проанализируй этот эскиз одежды и опиши все элементы конструкции",
                },
                {
                    Type: interfaces.PonchoContentTypeImageURL,
                    ImageURL: &interfaces.PonchoImageURL{
                        URL: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...",
                    },
                },
            },
        },
    },
    MaxTokens: &[]int{1000}[0],
    Temperature: &[]float32{0.1}[0],
}

response, err := fw.Generate(ctx, request)
if err != nil {
    return err
}

fmt.Printf("Analysis result: %s\n", response.Message.Content[0].Text)
```

### Streaming Generation

```go
// Streaming text generation
request := &interfaces.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*interfaces.PonchoMessage{
        {
            Role: interfaces.PonchoRoleUser,
            Content: []*interfaces.PonchoContentPart{
                {
                    Type: interfaces.PonchoContentTypeText,
                    Text: "Расскажи о модных трендах этой осени",
                },
            },
        },
    },
}

err := fw.GenerateStreaming(ctx, request, func(chunk *interfaces.PonchoModelResponse) error {
    if chunk.Message != nil && len(chunk.Message.Content) > 0 {
        text := chunk.Message.Content[0].Text
        fmt.Print(text) // Stream output
    }
    return nil
})

if err != nil {
    log.Printf("Streaming error: %v", err)
}
```

---

## Prompt System

### Creating Prompt Templates

Create `.prompt` files in your prompts directory:

```handlebars
{{role "config"}}
model: glm-4.6v-flash
config:
  temperature: 0.1
  maxOutputTokens: 2000

{{role "system"}}
You are a precise fashion sketch analyzer. Convert visual information into structured JSON.
Output ONLY valid JSON with Russian keys and values.

{{role "user"}}
ЗАДАЧА:
Проанализируй эскиз одежды и извлеки все характеристики товара.

ИЗОБРАЖЕНИЕ ДЛЯ АНАЛИЗА:
{{media url=photoUrl}}

ОБЯЗАТЕЛЬНЫЕ АСПЕКТЫ АНАЛИЗА:
- Тип изделия (платье, юбка, брюки и т.п.)
- Силуэт и форма
- Линия верха (воротник, горловина)
- Рукава (длина, форма)
- Длина изделия
- Цвет и материал
- Декоративные элементы

ВЕРНИ ТОЛЬКО JSON-ОБЪЕКТ:
```

### Using Prompts

```go
// Execute prompt with variables
variables := map[string]interface{}{
    "photoUrl": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ...",
}

response, err := fw.promptManager.ExecutePrompt(
    ctx,
    "sketch_description", // template name
    variables,
    "glm-4.6v-flash",     // model override (optional)
)
if err != nil {
    return err
}

fmt.Printf("Prompt result: %s\n", response.Message.Content[0].Text)
```

### Advanced Prompt Features

```handlebars
{{#if isFashionSketch}}
{{role "system"}}
You are analyzing a fashion sketch, focus on construction details.
{{else}}
{{role "system"}}
You are analyzing a fashion photo, focus on style and mood.
{{/if}}

{{#each productDetails}}
- {{name}}: {{value}}
{{/each}}
```

---

## Tools Integration

### Registering Custom Tools

```go
// Implement the PonchoTool interface
type FashionTool struct {
    name string
}

func (ft *FashionTool) Name() string {
    return ft.name
}

func (ft *FashionTool) Description() string {
    return "Analyzes fashion items and returns characteristics"
}

func (ft *FashionTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "image_url": map[string]interface{}{
                "type": "string",
                "description": "URL of fashion item image",
            },
            "item_type": map[string]interface{}{
                "type": "string",
                "description": "Type of fashion item (dress, pants, etc.)",
            },
        },
        "required": []string{"image_url"},
    }
}

func (ft *FashionTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
    imageURL := input["image_url"].(string)
    itemType := input["item_type"].(string)

    // Your custom logic here
    result := map[string]interface{}{
        "category": itemType,
        "style":    "casual",
        "season":   "summer",
        "features": []string{"lightweight", "breathable"},
    }

    return result, nil
}

// Register the tool
fashionTool := &FashionTool{name: "fashion_analyzer"}
err := fw.RegisterTool("fashion_analyzer", fashionTool)
if err != nil {
    log.Fatal(err)
}
```

### Using Tools in Model Requests

```go
// Generate with tool calling
request := &interfaces.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*interfaces.PonchoMessage{
        {
            Role: interfaces.PonchoRoleUser,
            Content: []*interfaces.PonchoContentPart{
                {
                    Type: interfaces.PonchoContentTypeText,
                    Text: "Проанализируй это платье и определи его характеристики",
                },
            },
        },
    },
    Tools: []*interfaces.PonchoTool{
        {
            Type: "function",
            Function: &interfaces.PonchoFunction{
                Name:        "fashion_analyzer",
                Description: "Analyzes fashion items",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "image_url": map[string]interface{}{
                            "type": "string",
                        },
                        "item_type": map[string]interface{}{
                            "type": "string",
                        },
                    },
                },
            },
        },
    },
}

response, err := fw.Generate(ctx, request)
if err != nil {
    return err
}

// Handle tool calls in response
if len(response.Message.ToolCalls) > 0 {
    for _, toolCall := range response.Message.ToolCalls {
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

## Wildberries API

### Setting up Wildberries Integration

```go
// Configure Wildberries tool in config.yaml
tools:
  wildberries_client:
    type: wildberries
    config:
      api_key: "${WB_API_CONTENT_KEY}"
      timeout: 30
      rate_limit:
        requests_per_minute: 100
```

### Fashion Item Analysis

```go
// Analyze fashion item with Wildberries data
analyzer := wildberries.NewFashionAnalyzer(wbClient)

analysis, err := analyzer.AnalyzeFashionItem(ctx, imageURL, description)
if err != nil {
    return err
}

fmt.Printf("Analysis Results:\n")
fmt.Printf("Category: %s\n", analysis.Category)
fmt.Printf("Attributes: %+v\n", analysis.Attributes)
fmt.Printf("Market Data: %+v\n", analysis.MarketData)
```

### Wildberries Product Data

```go
// Get product information
client := wildberries.NewWildberriesClient(apiKey)

productInfo, err := client.GetProductInfo(ctx, "12345678")
if err != nil {
    return err
}

fmt.Printf("Product: %s\n", productInfo.Data.Title)
fmt.Printf("Price: %.2f\n", productInfo.Data.Price)
fmt.Printf("Category: %s\n", productInfo.Data.SubjectName)

// Get characteristics
characteristics, err := client.GetCharacteristics(ctx, "12345678")
if err != nil {
    return err
}

for _, char := range characteristics.Data {
    fmt.Printf("%s: %s\n", char.Name, char.Value)
}
```

### Batch Processing

```go
// Process multiple items
items := []wildberries.FashionItem{
    {ImageURL: "url1", Description: "платье вечернее"},
    {ImageURL: "url2", Description: "костюм деловой"},
    {ImageURL: "url3", Description: "футболка спортивная"},
}

results := make([]*wildberries.FashionAnalysis, len(items))
var wg sync.WaitGroup
semaphore := make(chan struct{}, 5) // Limit concurrent requests

for i, item := range items {
    wg.Add(1)
    go func(index int, fashionItem wildberries.FashionItem) {
        defer wg.Done()
        semaphore <- struct{}{}
        defer func() { <-semaphore }()

        analysis, err := analyzer.AnalyzeFashionItem(ctx, fashionItem.ImageURL, fashionItem.Description)
        if err != nil {
            log.Printf("Error processing item %d: %v", index, err)
            return
        }

        results[index] = analysis
    }(i, item)
}

wg.Wait()
```

---

## Flow Orchestration

### Creating Custom Flows

```go
// Implement the PonchoFlow interface
type FashionAnalysisFlow struct {
    name    string
    fw      interfaces.PonchoFramework
    prompts map[string]string
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
                "type": "string",
                "description": "Base64 encoded fashion image",
            },
            "analysis_type": map[string]interface{}{
                "type": "string",
                "enum": []string{"description", "characteristics", "creative"},
                "description": "Type of analysis to perform",
            },
        },
        "required": []string{"image_url"},
    }
}

func (faf *FashionAnalysisFlow) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
    imageURL := input["image_url"].(string)
    analysisType := input["analysis_type"].(string)

    // Step 1: Vision analysis
    variables := map[string]interface{}{
        "photoUrl": imageURL,
    }

    promptName := "sketch_" + analysisType
    visionResponse, err := faf.fw.GetPromptManager().ExecutePrompt(ctx, promptName, variables, "glm-4.6v-flash")
    if err != nil {
        return nil, err
    }

    // Step 2: Post-processing if needed
    if analysisType == "creative" {
        // Add creative enhancement
        enhanceRequest := &interfaces.PonchoModelRequest{
            Model: "deepseek-chat",
            Messages: []*interfaces.PonchoMessage{
                {
                    Role: interfaces.PonchoRoleUser,
                    Content: []*interfaces.PonchoContentPart{
                        {
                            Type: interfaces.PonchoContentTypeText,
                            Text: fmt.Sprintf("Сделай описание более продающим: %s", visionResponse.Message.Content[0].Text),
                        },
                    },
                },
            },
        }

        enhancedResponse, err := faf.fw.Generate(ctx, enhanceRequest)
        if err != nil {
            return visionResponse.Message.Content[0].Text, nil // Return original if enhancement fails
        }

        return enhancedResponse.Message.Content[0].Text, nil
    }

    return visionResponse.Message.Content[0].Text, nil
}

// Register the flow
fashionFlow := &FashionAnalysisFlow{
    name:    "fashion_analysis",
    fw:      fw,
    prompts: map[string]string{"description": "sketch_description"},
}

err := fw.RegisterFlow("fashion_analysis", fashionFlow)
if err != nil {
    log.Fatal(err)
}
```

### Using Flows

```go
// Execute a flow
input := map[string]interface{}{
    "image_url":     "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ...",
    "analysis_type": "characteristics",
}

result, err := fw.ExecuteFlow(ctx, "fashion_analysis", input)
if err != nil {
    return err
}

fmt.Printf("Flow result: %v\n", result)
```

### Streaming Flows

```go
// Execute flow with streaming
err := fw.ExecuteFlowStreaming(ctx, "fashion_analysis", input, func(chunk *interfaces.PonchoModelResponse) error {
    if chunk.Message != nil && len(chunk.Message.Content) > 0 {
        fmt.Print(chunk.Message.Content[0].Text)
    }
    return nil
})
```

---

## Fashion Industry Features

### Sketch Analysis

```go
// Analyze fashion sketches with structured output
sketchAnalyzer := flows.NewFashionSketchAnalyzer(fw, promptManager)

request := &fashion.SketchAnalysisRequest{
    ImageURL:     "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ...",
    OutputFormat: "json", // or "text"
    Language:     "russian",
}

result, err := sketchAnalyzer.Analyze(ctx, request)
if err != nil {
    return err
}

switch result := result.(type) {
case *fashion.SketchCharacteristics:
    fmt.Printf("Item Type: %s\n", result.ItemType)
    fmt.Printf("Silhouette: %s\n", result.Silhouette)
    fmt.Printf("Sleeves: %s\n", result.Sleeves)
    fmt.Printf("Length: %s\n", result.Length)

case string:
    fmt.Printf("Creative Description: %s\n", result)
}
```

### Product Content Generation

```go
// Generate product content for e-commerce
contentGenerator := flows.NewProductContentGenerator(fw)

product := &fashion.Product{
    Name:        "Вечернее платье",
    Category:    "Платья",
    Description: "Элегантное платье для особых случаев",
    ImageURL:    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ...",
}

content, err := contentGenerator.Generate(ctx, product)
if err != nil {
    return err
}

fmt.Printf("Title: %s\n", content.Title)
fmt.Printf("Description: %s\n", content.Description)
fmt.Printf("Keywords: %v\n", content.Keywords)
fmt.Printf("SEO Tags: %v\n", content.SEOTags)
```

### Trend Analysis

```go
// Analyze fashion trends
trendAnalyzer := flows.NewTrendAnalyzer(fw)

trends, err := trendAnalyzer.Analyze(ctx, &fashion.TrendRequest{
    Season:     "осень-зима 2024",
    Categories: []string{"платья", "пальто", "акессуары"},
    Region:     "Россия",
})
if err != nil {
    return err
}

for _, trend := range trends.Trends {
    fmt.Printf("Trend: %s\n", trend.Name)
    fmt.Printf("Description: %s\n", trend.Description)
    fmt.Printf("Popularity: %.1f%%\n", trend.Popularity)
}
```

---

## Testing

### Unit Testing

```go
func TestFashionAnalysis(t *testing.T) {
    // Setup test framework
    fw := setupTestFramework(t)
    ctx := context.Background()

    // Test prompt execution
    variables := map[string]interface{}{
        "photoUrl": "data:image/jpeg;base64,test_image_data",
    }

    response, err := fw.GetPromptManager().ExecutePrompt(ctx, "sketch_description", variables, "glm-4.6v-flash")
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.NotEmpty(t, response.Message.Content[0].Text)
}
```

### Integration Testing

```go
func TestRealAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Test with real API keys
    if os.Getenv("ZAI_API_KEY") == "" {
        t.Skip("ZAI_API_KEY not set")
    }

    fw := setupRealFramework(t)
    ctx := context.Background()

    // Test actual vision analysis
    request := &interfaces.PonchoModelRequest{
        Model: "glm-4.6v-flash",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: "Что изображено на этой фотографии?",
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{
                            URL: "https://example.com/fashion-item.jpg",
                        },
                    },
                },
            },
        },
    }

    response, err := fw.Generate(ctx, request)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Message.Content[0].Text)
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Run specific test
go test -v ./flows -run TestFashionAnalysis

# Run benchmarks
go test -bench=. ./models/...
```

---

## Production Deployment

### Docker Configuration

```dockerfile
# Dockerfile.production
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o poncho-framework ./cmd/poncho-framework

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/poncho-framework .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./poncho-framework", "server", "--config", "configs/production.yaml"]
```

### Production Configuration

```yaml
# configs/production.yaml
models:
  deepseek-chat:
    provider: deepseek
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com/v1"
    timeout: "60s"
    max_tokens: 4000
    retry:
      attempts: 3
      delay: "1s"
      backoff: "exponential"

logging:
  level: "warn"
  format: "json"
  file: "/var/log/poncho-framework.log"

metrics:
  enabled: true
  endpoint: "http://prometheus:9090"
  interval: "30s"

health_check:
  enabled: true
  endpoint: "/health"
  interval: "10s"

security:
  api_key_required: true
  rate_limiting:
    requests_per_minute: 1000
    burst_size: 100
```

### Monitoring Setup

```go
// Add custom metrics
func setupMetrics(fw interfaces.PonchoFramework) {
    // Request counter
    prometheus.Register(prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "poncho_requests_total",
            Help: "Total number of requests",
        },
        []string{"model", "status"},
    ))

    // Response time histogram
    prometheus.Register(prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "poncho_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"model"},
    ))
}
```

### Health Checks

```go
// Custom health check
func (fw *PonchoFramework) Health(ctx context.Context) (*interfaces.PonchoHealthStatus, error) {
    components := make(map[string]*interfaces.ComponentHealth)

    // Check models
    modelsHealth := &interfaces.ComponentHealth{Status: "healthy"}
    for modelName := range fw.models {
        if err := fw.checkModelHealth(ctx, modelName); err != nil {
            modelsHealth.Status = "unhealthy"
            modelsHealth.Message = fmt.Sprintf("Model %s failed: %v", modelName, err)
            break
        }
    }
    components["models"] = modelsHealth

    // Check external services
    if wbHealth := fw.checkWildberriesHealth(ctx); wbHealth != nil {
        components["wildberries"] = wbHealth
    }

    overallStatus := "healthy"
    for _, component := range components {
        if component.Status != "healthy" {
            overallStatus = "unhealthy"
            break
        }
    }

    return &interfaces.PonchoHealthStatus{
        Status:     overallStatus,
        Timestamp:  time.Now(),
        Version:    fw.version,
        Components: components,
        Uptime:     time.Since(fw.startTime),
    }, nil
}
```

### Rate Limiting

```go
// Implement rate limiting for API endpoints
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(s.config.RateLimit.RequestsPerMinute), s.config.RateLimit.BurstSize)

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Advanced Usage Examples

### Multi-Model Pipeline

```go
func analyzeFashionItem(ctx context.Context, fw interfaces.PonchoFramework, imageURL string) (*FashionAnalysisResult, error) {
    // Step 1: Vision analysis with GLM-4.6V
    visionRequest := &interfaces.PonchoModelRequest{
        Model: "glm-4.6v-flash",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: "Извлечи все характеристики этого предмета одежды",
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{URL: imageURL},
                    },
                },
            },
        },
        MaxTokens: &[]int{800}[0],
        Temperature: &[]float32{0.1}[0],
    }

    visionResponse, err := fw.Generate(ctx, visionRequest)
    if err != nil {
        return nil, err
    }

    // Step 2: Structured data extraction with DeepSeek
    structureRequest := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf("Преобразуй в JSON: %s", visionResponse.Message.Content[0].Text),
                    },
                },
            },
        },
        MaxTokens: &[]int{400}[0],
        Temperature: &[]float32{0.0}[0],
    }

    structureResponse, err := fw.Generate(ctx, structureRequest)
    if err != nil {
        return nil, err
    }

    // Step 3: Creative description generation
    creativeRequest := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf("Напиши продающее описание для: %s", visionResponse.Message.Content[0].Text),
                    },
                },
            },
        },
        MaxTokens: &[]int{300}[0],
        Temperature: &[]float32{0.8}[0],
    }

    creativeResponse, err := fw.Generate(ctx, creativeRequest)
    if err != nil {
        return nil, err
    }

    // Compile result
    result := &FashionAnalysisResult{
        Characteristics: structureResponse.Message.Content[0].Text,
        Description:     creativeResponse.Message.Content[0].Text,
        RawAnalysis:     visionResponse.Message.Content[0].Text,
    }

    return result, nil
}
```

### Batch Processing with Workers

```go
func processBatch(ctx context.Context, fw interfaces.PonchoFramework, items []FashionItem) []ProcessResult {
    const numWorkers = 5
    jobs := make(chan FashionItem, len(items))
    results := make(chan ProcessResult, len(items))

    // Start workers
    for w := 0; w < numWorkers; w++ {
        go func() {
            for item := range jobs {
                result := ProcessResult{
                    ID: item.ID,
                }

                analysis, err := analyzeFashionItem(ctx, fw, item.ImageURL)
                if err != nil {
                    result.Error = err
                } else {
                    result.Analysis = analysis
                }

                results <- result
            }
        }()
    }

    // Send jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    // Collect results
    var processResults []ProcessResult
    for i := 0; i < len(items); i++ {
        result := <-results
        processResults = append(processResults, result)
    }

    return processResults
}
```

This comprehensive user guide provides everything needed to effectively use the PonchoAiFramework, from basic setup to advanced production deployments with a focus on fashion industry applications.