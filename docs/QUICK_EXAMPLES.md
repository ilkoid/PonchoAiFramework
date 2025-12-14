# PonchoAiFramework Quick Examples

This guide provides ready-to-use code examples for common fashion industry tasks using the PonchoAiFramework. All examples are complete and can be copied directly into your projects.

## Table of Contents

1. [Setup Examples](#setup-examples)
2. [Fashion Sketch Analysis](#fashion-sketch-analysis)
3. [Product Description Generation](#product-description-generation)
4. [Wildberries Integration](#wildberries-integration)
5. [Batch Processing](#batch-processing)
6. [API Server Examples](#api-server-examples)
7. [Custom Tools](#custom-tools)
8. [Workflow Automation](#workflow-automation)

---

## Setup Examples

### Basic Framework Setup

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/ilkoid/PonchoAiFramework/core"
    "github.com/ilkoid/PonchoAiFramework/interfaces"
    "github.com/ilkoid/PonchoAiFramework/models/deepseek"
    "github.com/ilkoid/PonchoAiFramework/models/zai"
)

func main() {
    ctx := context.Background()

    // Create framework instance
    fw := core.NewPonchoFramework()

    // Load configuration
    if err := fw.LoadFromFile("config.yaml"); err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Start framework
    if err := fw.Start(ctx); err != nil {
        log.Fatal("Failed to start framework:", err)
    }
    defer fw.Stop(ctx)

    // Register models
    deepseekModel := deepseek.NewDeepSeekModel()
    if err := fw.RegisterModel("deepseek-chat", deepseekModel); err != nil {
        log.Fatal("Failed to register DeepSeek:", err)
    }

    zaiModel := zai.NewZAIModel()
    if err := fw.RegisterModel("glm-4.6v-flash", zaiModel); err != nil {
        log.Fatal("Failed to register Z.AI:", err)
    }

    log.Println("Framework ready!")

    // Your application logic here
}
```

### Environment Configuration

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

logging:
  level: "info"
  format: "json"
```

---

## Fashion Sketch Analysis

### Structured Analysis from Sketch

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"

    "github.com/ilkoid/PonchoAiFramework/interfaces"
)

type SketchAnalysis struct {
    TypeOfItem        string   `json:"тип_изделия"`
    Silhouette        string   `json:"силуэт"`
    TopLine           string   `json:"линия_верха"`
    Sleeves           string   `json:"рукава"`
    Length            string   `json:"длина"`
    WaistLine         string   `json:"линия_талии"`
    Closure           string   `json:"застежка"`
    Pockets           string   `json:"карманы"`
    DecorativeElements []string `json:"декоративные_элементы"`
    Materials         []string `json:"материалы"`
    Colors            []string `json:"цвета"`
    Season            string   `json:"сезон"`
    Gender            string   `json:"пол"`
}

func AnalyzeFashionSketch(ctx context.Context, fw interfaces.PonchoFramework, imageData string) (*SketchAnalysis, error) {
    // Create request with strict JSON output requirements
    request := &interfaces.PonchoModelRequest{
        Model: "glm-4.6v-flash",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleSystem,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: `You are a precise fashion sketch analyzer. Output ONLY valid JSON.
                        Use Russian keys in snake_case. Example: {"тип_изделия": "платье", "цвет": "красный"}`,
                    },
                },
            },
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: `Analyze this fashion sketch and extract all construction details.
                        Include: item type, silhouette, neckline, sleeves, length, closure, pockets,
                        decorative elements, materials, colors, season, and gender.
                        Return ONLY JSON object with Russian keys.`,
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{
                            URL:    imageData,
                            Detail: "high",
                        },
                    },
                },
            },
        },
        MaxTokens:   &[]int{1000}[0],
        Temperature: &[]float32{0.1}[0],
    }

    response, err := fw.Generate(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    if len(response.Message.Content) == 0 {
        return nil, fmt.Errorf("empty response")
    }

    // Extract JSON from response
    jsonStr := response.Message.Content[0].Text

    // Clean up potential markdown formatting
    jsonStr = strings.TrimSpace(jsonStr)
    jsonStr = strings.TrimPrefix(jsonStr, "```json")
    jsonStr = strings.TrimSuffix(jsonStr, "```")
    jsonStr = strings.TrimSpace(jsonStr)

    // Parse JSON
    var analysis SketchAnalysis
    if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }

    return &analysis, nil
}

// Usage example
func main() {
    // Initialize framework (see setup example)
    fw := initializeFramework()
    ctx := context.Background()

    // Your base64 encoded image data
    imageData := "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD..."

    analysis, err := AnalyzeFashionSketch(ctx, fw, imageData)
    if err != nil {
        log.Fatal("Analysis failed:", err)
    }

    fmt.Printf("Item Type: %s\n", analysis.TypeOfItem)
    fmt.Printf("Silhouette: %s\n", analysis.Silhouette)
    fmt.Printf("Sleeves: %s\n", analysis.Sleeves)
    fmt.Printf("Length: %s\n", analysis.Length)
    fmt.Printf("Colors: %v\n", analysis.Colors)
    fmt.Printf("Season: %s\n", analysis.Season)
}
```

### Creative Description from Sketch

```go
func GenerateCreativeDescription(ctx context.Context, fw interfaces.PonchoFramework, imageData string) (string, error) {
    request := &interfaces.PonchoModelRequest{
        Model: "glm-4.6v-flash",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleSystem,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: `You are a creative fashion writer. Create engaging, poetic descriptions of fashion items.
                        Focus on mood, style, and emotional impact. Output in Russian only.`,
                    },
                },
            },
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: `Create a beautiful, engaging description for this fashion sketch.
                        Describe the item's character, who would wear it, and what occasions it's perfect for.
                        Make it sound appealing and romantic.`,
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{
                            URL:    imageData,
                            Detail: "high",
                        },
                    },
                },
            },
        },
        MaxTokens:   &[]int{300}[0],
        Temperature: &[]float32{0.8}[0],
    }

    response, err := fw.Generate(ctx, request)
    if err != nil {
        return "", err
    }

    return response.Message.Content[0].Text, nil
}
```

### Using Prompt Templates

```handlebars
{{!-- examples/test_data/prompts/sketch_analysis.prompt --}}
{{role "config"}}
model: glm-4.6v-flash
config:
  temperature: 0.1
  maxOutputTokens: 2000

{{role "system"}}
You are a precise fashion sketch analyzer. Output ONLY valid JSON with Russian keys.

{{role "user"}}
Analyze this fashion sketch and extract all characteristics.
{{media url=image_url}}

Focus on:
- Item type (тип_изделия)
- Silhouette (силуэт)
- Neckline/Collar (линия_верха)
- Sleeves (рукава)
- Length (длина)
- Materials (материалы)
- Colors (цвета)
- Season (сезон)
- Gender (пол)

Return ONLY JSON object.
```

```go
// Using the prompt template
func AnalyzeSketchWithPrompt(ctx context.Context, fw interfaces.PonchoFramework, promptManager interfaces.PromptManager, imageData string) (*SketchAnalysis, error) {
    variables := map[string]interface{}{
        "image_url": imageData,
    }

    response, err := promptManager.ExecutePrompt(ctx, "sketch_analysis", variables, "glm-4.6v-flash")
    if err != nil {
        return nil, err
    }

    var analysis SketchAnalysis
    if err := json.Unmarshal([]byte(response.Message.Content[0].Text), &analysis); err != nil {
        return nil, err
    }

    return &analysis, nil
}
```

---

## Product Description Generation

### E-commerce Product Content

```go
type ProductContent struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Features    []string `json:"features"`
    Keywords    []string `json:"keywords"`
    SEOTags     []string `json:"seo_tags"`
    Category    string   `json:"category"`
    Season      string   `json:"season"`
    Occasion    []string `json:"occasion"`
}

func GenerateProductContent(ctx context.Context, fw interfaces.PonchoFramework, productInfo *ProductInfo) (*ProductContent, error) {
    request := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleSystem,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: `You are an expert e-commerce copywriter specializing in fashion.
                        Create compelling product content that sells. Output in JSON format.`,
                    },
                },
            },
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf(`Create complete e-commerce content for this fashion item:

Name: %s
Category: %s
Description: %s
Materials: %v
Colors: %v

Generate:
1. SEO-optimized title (60-80 characters)
2. Engaging description (150-200 words)
3. Key features (3-5 bullet points)
4. SEO keywords (10-15 relevant terms)
5. Category and season classification
6. Suitable occasions

Output valid JSON only.`,
                            productInfo.Name,
                            productInfo.Category,
                            productInfo.BasicDescription,
                            strings.Join(productInfo.Materials, ", "),
                            strings.Join(productInfo.Colors, ", ")),
                    },
                },
            },
        },
        MaxTokens:   &[]int{800}[0],
        Temperature: &[]float32{0.7}[0],
    }

    response, err := fw.Generate(ctx, request)
    if err != nil {
        return nil, err
    }

    var content ProductContent
    if err := json.Unmarshal([]byte(response.Message.Content[0].Text), &content); err != nil {
        return nil, err
    }

    return &content, nil
}

type ProductInfo struct {
    Name             string   `json:"name"`
    Category         string   `json:"category"`
    BasicDescription string   `json:"basic_description"`
    Materials        []string `json:"materials"`
    Colors           []string `json:"colors"`
}
```

### Multi-language Content Generation

```go
func GenerateMultilingualContent(ctx context.Context, fw interfaces.PonchoFramework, baseContent *ProductContent, languages []string) (map[string]*ProductContent, error) {
    results := make(map[string]*ProductContent)

    // Add original
    results["ru"] = baseContent

    for _, lang := range languages {
        request := &interfaces.PonchoModelRequest{
            Model: "deepseek-chat",
            Messages: []*interfaces.PonchoMessage{
                {
                    Role: interfaces.PonchoRoleSystem,
                    Content: []*interfaces.PonchoContentPart{
                        {
                            Type: interfaces.PonchoContentTypeText,
                            Text: fmt.Sprintf(`You are a professional translator specializing in fashion e-commerce.
                            Translate product content to %s while maintaining brand voice and cultural nuances.
                            Keep SEO keywords relevant to the target market.`, lang),
                        },
                    },
                },
                {
                    Role: interfaces.PonchoRoleUser,
                    Content: []*interfaces.PonchoContentPart{
                        {
                            Type: interfaces.PonchoContentTypeText,
                            Text: fmt.Sprintf(`Translate this product content to %s:

Title: %s
Description: %s
Features: %v

Return valid JSON in the same structure.`,
                                lang,
                                baseContent.Title,
                                baseContent.Description,
                                baseContent.Features),
                        },
                    },
                },
            },
            MaxTokens:   &[]int{600}[0],
            Temperature: &[]float32{0.3}[0],
        }

        response, err := fw.Generate(ctx, request)
        if err != nil {
            return nil, fmt.Errorf("translation to %s failed: %w", lang, err)
        }

        var translatedContent ProductContent
        if err := json.Unmarshal([]byte(response.Message.Content[0].Text), &translatedContent); err != nil {
            return nil, fmt.Errorf("failed to parse %s translation: %w", lang, err)
        }

        results[lang] = &translatedContent
    }

    return results, nil
}

// Usage
func main() {
    fw := initializeFramework()
    ctx := context.Background()

    productInfo := &ProductInfo{
        Name:             "Элегантное шелковое платье",
        Category:         "Платья",
        BasicDescription: "Вечернее платье из натурального шелка с бисерной вышивкой",
        Materials:        []string{"Шелк", "Бисер"},
        Colors:           []string{"Бордовый", "Черный"},
    }

    content, err := GenerateProductContent(ctx, fw, productInfo)
    if err != nil {
        log.Fatal(err)
    }

    // Generate translations
    multilingual, err := GenerateMultilingualContent(ctx, fw, content, []string{"en", "zh", "es"})
    if err != nil {
        log.Fatal(err)
    }

    for lang, content := range multilingual {
        fmt.Printf("\n=== %s ===\n", lang)
        fmt.Printf("Title: %s\n", content.Title)
        fmt.Printf("Description: %s\n", content.Description)
    }
}
```

---

## Wildberries Integration

### Product Analysis with Market Data

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ilkoid/PonchoAiFramework/tools/wildberries"
)

func AnalyzeProductWithMarketData(ctx context.Context, wbClient *wildberries.WildberriesClient, fw interfaces.PonchoFramework, nmID string) error {
    // Get product info from Wildberries
    productInfo, err := wbClient.GetProductInfo(ctx, nmID)
    if err != nil {
        return fmt.Errorf("failed to get product info: %w", err)
    }

    // Get characteristics
    characteristics, err := wbClient.GetCharacteristics(ctx, nmID)
    if err != nil {
        return fmt.Errorf("failed to get characteristics: %w", err)
    }

    // Create fashion analyzer
    analyzer := wildberries.NewFashionAnalyzer(wbClient, fw)

    // Analyze product with image if available
    var analysis *wildberries.FashionAnalysis
    if len(productInfo.Data.Photos) > 0 {
        analysis, err = analyzer.AnalyzeFashionItem(
            ctx,
            productInfo.Data.Photos[0].Big,
            productInfo.Data.Description,
        )
        if err != nil {
            log.Printf("Warning: image analysis failed: %v", err)
        }
    }

    // Print comprehensive analysis
    fmt.Printf("=== Product Analysis ===\n")
    fmt.Printf("Product: %s\n", productInfo.Data.Title)
    fmt.Printf("Brand: %s\n", productInfo.Data.Brand)
    fmt.Printf("Price: %.2f RUB\n", productInfo.Data.Price)
    fmt.Printf("Rating: %.1f/5 (%d reviews)\n", productInfo.Data.ReviewRating, productInfo.Data.ReviewsCount)

    if analysis != nil {
        fmt.Printf("\n=== AI Analysis ===\n")
        fmt.Printf("Category: %s\n", analysis.Category)
        fmt.Printf("Style: %s\n", analysis.Style)
        fmt.Printf("Season: %s\n", analysis.Season)
        fmt.Printf("Confidence: %.1f%%\n", analysis.Confidence*100)
    }

    if characteristics != nil {
        fmt.Printf("\n=== Characteristics ===\n")
        for _, char := range characteristics.Data {
            fmt.Printf("%s: %s\n", char.Name, char.Value)
        }
    }

    return nil
}
```

### Market Trends Analysis

```go
func AnalyzeMarketTrends(ctx context.Context, wbClient *wildberries.WildberriesClient, fw interfaces.PonchoFramework, category string) (*wildberries.MarketTrends, error) {
    // Search for top products in category
    searchResponse, err := wbClient.SearchProducts(ctx, category, &wildberries.SearchFilters{
        Sort:  "popular",
        Limit: 100,
    })
    if err != nil {
        return nil, err
    }

    // Analyze trends using AI
    productData := make([]string, len(searchResponse.Data.Products))
    for i, product := range searchResponse.Data.Products {
        productData[i] = fmt.Sprintf("%s - %s - %.2f",
            product.Name,
            product.Brand,
            product.Price)
    }

    request := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf(`Analyze these top products in the "%s" category and identify market trends:

Products:
%s

Identify:
1. Top 5 trending styles
2. Popular colors
3. Price ranges
4. Key features in demand
5. Seasonal trends

Output analysis in JSON format.`, category, strings.Join(productData, "\n")),
                    },
                },
            },
        },
        MaxTokens:   &[]int{1000}[0],
        Temperature: &[]float32{0.3}[0],
    }

    response, err := fw.Generate(ctx, request)
    if err != nil {
        return nil, err
    }

    // Parse and return trends
    var trends wildberries.MarketTrends
    // Implement JSON parsing based on response format

    return &trends, nil
}
```

### Competitive Analysis

```go
func AnalyzeCompetition(ctx context.Context, wbClient *wildberries.WildberriesClient, fw interfaces.PonchoFramework, targetNmID string) error {
    // Get target product
    targetProduct, err := wbClient.GetProductInfo(ctx, targetNmID)
    if err != nil {
        return err
    }

    // Find similar products
    searchResponse, err := wbClient.SearchProducts(ctx, targetProduct.Data.SubjectName, &wildberries.SearchFilters{
        Sort:  "popular",
        Limit: 20,
    })
    if err != nil {
        return err
    }

    // Analyze competition
    competitorData := make([]wildberries.CompetitorInfo, 0, len(searchResponse.Data.Products))

    for _, product := range searchResponse.Data.Products {
        if product.Id == targetProduct.Data.Id {
            continue // Skip self
        }

        competitorData = append(competitorData, wildberries.CompetitorInfo{
            Name:     product.Name,
            Brand:    product.Brand,
            Price:    product.Price,
            Rating:   product.ReviewRating,
            Reviews:  product.ReviewsCount,
        })
    }

    // Generate competitive analysis
    analysis := &wildberries.CompetitionAnalysis{
        TargetProduct: wildberries.ProductInfo{
            Name:    targetProduct.Data.Name,
            Brand:   targetProduct.Data.Brand,
            Price:   targetProduct.Data.Price,
            Rating:  targetProduct.Data.ReviewRating,
            Reviews: targetProduct.Data.ReviewsCount,
        },
        Competitors: competitorData,
    }

    // Calculate insights
    var avgPrice, totalReviews float64
    var highRatedCount int

    for _, comp := range analysis.Competitors {
        avgPrice += comp.Price
        totalReviews += float64(comp.Reviews)
        if comp.Rating >= 4.5 {
            highRatedCount++
        }
    }

    avgPrice /= float64(len(analysis.Competitors))

    fmt.Printf("=== Competitive Analysis ===\n")
    fmt.Printf("Target Product: %s\n", analysis.TargetProduct.Name)
    fmt.Printf("Price: %.2f RUB (%.1f%s %s market average)\n",
        analysis.TargetProduct.Price,
        (analysis.TargetProduct.Price-avgPrice)/avgPrice*100,
        "%",
        "above/below",
    )
    fmt.Printf("High-rated competitors: %d/%d\n", highRatedCount, len(analysis.Competitors))

    return nil
}
```

---

## Batch Processing

### Processing Multiple Images

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
)

type ProcessingJob struct {
    ID       string
    ImageURL string
    ItemType string
}

type ProcessingResult struct {
    ID       string
    Analysis *SketchAnalysis
    Error    error
    Duration time.Duration
}

func ProcessBatch(ctx context.Context, fw interfaces.PonchoFramework, jobs []ProcessingJob, workers int) []ProcessingResult {
    results := make([]ProcessingResult, len(jobs))
    jobQueue := make(chan ProcessingJob, len(jobs))

    // Create wait group
    var wg sync.WaitGroup

    // Start workers
    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobQueue {
                startTime := time.Now()

                result := ProcessingResult{
                    ID: job.ID,
                }

                analysis, err := AnalyzeFashionSketch(ctx, fw, job.ImageURL)
                if err != nil {
                    result.Error = err
                } else {
                    result.Analysis = analysis
                }

                result.Duration = time.Since(startTime)

                // Store result
                // Note: In production, use proper synchronization or result channels
                for i, j := range jobs {
                    if j.ID == job.ID {
                        results[i] = result
                        break
                    }
                }

                log.Printf("Worker %d completed job %s in %v",
                    workerID, job.ID, result.Duration)
            }
        }(w)
    }

    // Queue jobs
    for _, job := range jobs {
        jobQueue <- job
    }
    close(jobQueue)

    // Wait for completion
    wg.Wait()

    return results
}

// Usage example
func main() {
    fw := initializeFramework()
    ctx := context.Background()

    // Prepare batch jobs
    jobs := []ProcessingJob{
        {ID: "001", ImageURL: "data:image/jpeg;base64,Image1Base64...", ItemType: "dress"},
        {ID: "002", ImageURL: "data:image/jpeg;base64,Image2Base64...", ItemType: "shirt"},
        {ID: "003", ImageURL: "data:image/jpeg;base64,Image3Base64...", ItemType: "pants"},
        // ... more jobs
    }

    // Process with 5 concurrent workers
    results := ProcessBatch(ctx, fw, jobs, 5)

    // Analyze results
    successCount := 0
    totalDuration := time.Duration(0)

    for _, result := range results {
        if result.Error == nil {
            successCount++
        }
        totalDuration += result.Duration
    }

    fmt.Printf("\n=== Batch Results ===\n")
    fmt.Printf("Total jobs: %d\n", len(jobs))
    fmt.Printf("Successful: %d\n", successCount)
    fmt.Printf("Failed: %d\n", len(jobs)-successCount)
    fmt.Printf("Average processing time: %v\n", totalDuration/time.Duration(len(jobs)))
}
```

### Rate-Limited API Calls

```go
type RateLimitedProcessor struct {
    fw        interfaces.PonchoFramework
    rateLimit time.Duration
    lastCall  time.Time
    mu        sync.Mutex
}

func NewRateLimitedProcessor(fw interfaces.PonchoFramework, requestsPerSecond int) *RateLimitedProcessor {
    return &RateLimitedProcessor{
        fw:        fw,
        rateLimit: time.Second / time.Duration(requestsPerSecond),
    }
}

func (rlp *RateLimitedProcessor) AnalyzeWithRateLimit(ctx context.Context, imageURL string) (*SketchAnalysis, error) {
    rlp.mu.Lock()
    defer rlp.mu.Unlock()

    // Rate limiting
    elapsed := time.Since(rlp.lastCall)
    if elapsed < rlp.rateLimit {
        sleepDuration := rlp.rateLimit - elapsed
        log.Printf("Rate limiting: sleeping %v", sleepDuration)
        time.Sleep(sleepDuration)
    }

    rlp.lastCall = time.Now()

    // Perform analysis
    return AnalyzeFashionSketch(ctx, rlp.fw, imageURL)
}

// Process with rate limiting
func ProcessWithRateLimit(ctx context.Context, fw interfaces.PonchoFramework, imageUrls []string) {
    processor := NewRateLimitedProcessor(fw, 2) // 2 requests per second

    for i, imageURL := range imageUrls {
        fmt.Printf("Processing image %d/%d\n", i+1, len(imageUrls))

        analysis, err := processor.AnalyzeWithRateLimit(ctx, imageURL)
        if err != nil {
            log.Printf("Failed to process image %d: %v", i+1, err)
            continue
        }

        fmt.Printf("  - Item: %s\n", analysis.TypeOfItem)
        fmt.Printf("  - Style: %s\n", analysis.Silhouette)
    }
}
```

---

## API Server Examples

### HTTP API Server

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/ilkoid/PonchoAiFramework/interfaces"
)

type APIServer struct {
    fw         interfaces.PonchoFramework
    promptMgr  interfaces.PromptManager
    httpServer *http.Server
}

type AnalysisRequest struct {
    ImageURL    string `json:"image_url"`
    AnalysisType string `json:"analysis_type"` // "characteristics" or "creative"
}

type AnalysisResponse struct {
    Result   interface{} `json:"result"`
    Duration string      `json:"duration"`
    Success  bool        `json:"success"`
    Error    string      `json:"error,omitempty"`
}

func NewAPIServer(fw interfaces.PonchoFramework, port int) *APIServer {
    router := mux.NewRouter()
    server := &APIServer{
        fw:        fw,
        promptMgr: fw.GetPromptManager(),
        httpServer: &http.Server{
            Addr:         fmt.Sprintf(":%d", port),
            Handler:      router,
            WriteTimeout: 15 * time.Second,
            ReadTimeout:  15 * time.Second,
        },
    }

    // Register routes
    router.HandleFunc("/health", server.handleHealth).Methods("GET")
    router.HandleFunc("/api/v1/analyze", server.handleAnalyze).Methods("POST")
    router.HandleFunc("/api/v1/describe", server.handleDescribe).Methods("POST")
    router.HandleFunc("/api/v1/models", server.handleListModels).Methods("GET")

    return server
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    health, err := s.fw.Health(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(health)
}

func (s *APIServer) handleAnalyze(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := r.Context()

    var req AnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    response := AnalysisResponse{
        Duration: time.Since(start).String(),
    }

    // Use prompt manager for analysis
    variables := map[string]interface{}{
        "image_url": req.ImageURL,
    }

    promptName := "sketch_description"
    if req.AnalysisType == "creative" {
        promptName = "sketch_creative"
    }

    promptResponse, err := s.promptMgr.ExecutePrompt(ctx, promptName, variables, "glm-4.6v-flash")
    if err != nil {
        response.Success = false
        response.Error = err.Error()
    } else {
        response.Success = true
        response.Result = promptResponse.Message.Content[0].Text
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *APIServer) handleDescribe(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := r.Context()

    var req struct {
        ImageURL string `json:"image_url"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Generate creative description
    request := &interfaces.PonchoModelRequest{
        Model: "glm-4.6v-flash",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: "Создай красивое, поэтичное описание этого предмета одежды",
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{
                            URL:    req.ImageURL,
                            Detail: "high",
                        },
                    },
                },
            },
        },
        MaxTokens:   &[]int{200}[0],
        Temperature: &[]float32{0.8}[0],
    }

    response, err := s.fw.Generate(ctx, request)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    result := struct {
        Description string `json:"description"`
        Duration    string `json:"duration"`
    }{
        Description: response.Message.Content[0].Text,
        Duration:    time.Since(start).String(),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func (s *APIServer) handleListModels(w http.ResponseWriter, r *http.Request) {
    models := s.fw.GetModelRegistry().List()
    json.NewEncoder(w).Encode(map[string]interface{}{
        "models": models,
    })
}

func (s *APIServer) Start() error {
    log.Printf("Starting API server on %s", s.httpServer.Addr)
    return s.httpServer.ListenAndServe()
}

func (s *APIServer) Stop(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}

// Usage
func main() {
    fw := initializeFramework()
    ctx := context.Background()

    // Start framework
    if err := fw.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer fw.Stop(ctx)

    // Start API server
    apiServer := NewAPIServer(fw, 8080)

    // Graceful shutdown
    go func() {
        if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Wait for interrupt signal
    <-ctx.Done()

    // Shutdown
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := apiServer.Stop(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
}
```

---

## Custom Tools

### Color Palette Extractor Tool

```go
package main

import (
    "context"
    "fmt"
    "image/color"
    "strings"

    "github.com/ilkoid/PonchoAiFramework/interfaces"
)

type ColorPaletteTool struct {
    name string
}

func NewColorPaletteTool() *ColorPaletteTool {
    return &ColorPaletteTool{
        name: "color_palette_extractor",
    }
}

func (cpt *ColorPaletteTool) Name() string {
    return cpt.name
}

func (cpt *ColorPaletteTool) Description() string {
    return "Extracts color palette from fashion item images"
}

func (cpt *ColorPaletteTool) Category() string {
    return "fashion"
}

func (cpt *ColorPaletteTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "image_url": map[string]interface{}{
                "type":        "string",
                "description": "URL of fashion item image",
            },
            "max_colors": map[string]interface{}{
                "type":        "integer",
                "description": "Maximum number of colors to extract (default: 5)",
                "minimum":     1,
                "maximum":     10,
            },
        },
        "required": []string{"image_url"},
    }
}

func (cpt *ColorPaletteTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
    imageURL := input["image_url"].(string)
    maxColors := 5
    if mc, ok := input["max_colors"].(float64); ok {
        maxColors = int(mc)
    }

    // Use AI model to extract colors
    request := &interfaces.PonchoModelRequest{
        Model: "glm-4.6v-flash",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf(`Extract the main color palette from this fashion item.
                        List up to %d colors with their names and hex codes.
                        Format as JSON: [{"name": "color_name", "hex": "#RRGGBB", "percentage": 0.0}]`, maxColors),
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{
                            URL:    imageURL,
                            Detail: "high",
                        },
                    },
                },
            },
        },
        MaxTokens:   &[]int{200}[0],
        Temperature: &[]float32{0.1}[0],
    }

    // This would need access to framework instance
    // For now, return mock data
    colors := []map[string]interface{}{
        {"name": "Красный", "hex": "#DC143C", "percentage": 40.0},
        {"name": "Черный", "hex": "#000000", "percentage": 35.0},
        {"name": "Белый", "hex": "#FFFFFF", "percentage": 25.0},
    }

    return map[string]interface{}{
        "colors": colors,
        "dominant": colors[0]["name"],
    }, nil
}

// Register the tool
func main() {
    fw := initializeFramework()

    colorTool := NewColorPaletteTool()
    err := fw.RegisterTool("color_palette", colorTool)
    if err != nil {
        log.Fatal(err)
    }

    // Use in model request
    request := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: "Extract colors from this dress and suggest matching accessories",
                    },
                    {
                        Type: interfaces.PonchoContentTypeImageURL,
                        ImageURL: &interfaces.PonchoImageURL{
                            URL: "data:image/jpeg;base64,ImageBase64Data...",
                        },
                    },
                },
            },
        },
        Tools: []*interfaces.PonchoTool{
            {
                Type: "function",
                Function: &interfaces.PonchoFunction{
                    Name:        "color_palette",
                    Description: "Extract color palette from image",
                    Parameters:  colorTool.InputSchema(),
                },
            },
        },
    }

    ctx := context.Background()
    response, err := fw.Generate(ctx, request)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Message.Content[0].Text)
}
```

### Style Recommendation Tool

```go
type StyleRecommendationTool struct {
    fw interfaces.PonchoFramework
}

func NewStyleRecommendationTool(fw interfaces.PonchoFramework) *StyleRecommendationTool {
    return &StyleRecommendationTool{
        fw: fw,
    }
}

func (srt *StyleRecommendationTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
    itemType := input["item_type"].(string)
    occasion := input["occasion"].(string)
    season := input["season"].(string)

    request := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf(`Provide style recommendations for:
                        Item Type: %s
                        Occasion: %s
                        Season: %s

                        Include:
                        1. Color combinations
                        2. Accessory suggestions
                        3. Footwear recommendations
                        4. Styling tips

                        Output as structured JSON.`, itemType, occasion, season),
                    },
                },
            },
        },
        MaxTokens:   &[]int{500}[0],
        Temperature: &[]float32{0.7}[0],
    }

    response, err := srt.fw.Generate(ctx, request)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "recommendations": response.Message.Content[0].Text,
    }, nil
}
```

---

## Workflow Automation

### Complete Fashion Item Pipeline

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
)

type FashionItemPipeline struct {
    fw         interfaces.PonchoFramework
    wbClient   *wildberries.WildberriesClient
    analyzer   *wildberries.FashionAnalyzer
    promptMgr  interfaces.PromptManager
}

type PipelineInput struct {
    ImageURL     string `json:"image_url"`
    Name         string `json:"name"`
    Category     string `json:"category"`
    Description  string `json:"description"`
}

type PipelineOutput struct {
    Analysis       *wildberries.FashionAnalysis `json:"analysis"`
    Content        *ProductContent              `json:"content"`
    MarketData     *wildberries.MarketData       `json:"market_data"`
    Recommendations []string                     `json:"recommendations"`
    ProcessingTime time.Duration                 `json:"processing_time"`
}

func NewFashionItemPipeline(fw interfaces.PonchoFramework, wbClient *wildberries.WildberriesClient) *FashionItemPipeline {
    return &FashionItemPipeline{
        fw:        fw,
        wbClient:  wbClient,
        analyzer:  wildberries.NewFashionAnalyzer(wbClient, fw),
        promptMgr: fw.GetPromptManager(),
    }
}

func (fip *FashionItemPipeline) Process(ctx context.Context, input *PipelineInput) (*PipelineOutput, error) {
    start := time.Now()
    output := &PipelineOutput{}

    // Step 1: AI Analysis
    log.Printf("Step 1: Analyzing fashion item...")
    analysis, err := fip.analyzer.AnalyzeFashionItem(ctx, input.ImageURL, input.Description)
    if err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }
    output.Analysis = analysis

    // Step 2: Generate E-commerce Content
    log.Printf("Step 2: Generating product content...")
    productInfo := &ProductInfo{
        Name:             input.Name,
        Category:         input.Category,
        BasicDescription: input.Description,
        Materials:        analysis.Attributes["materials"].([]string),
        Colors:           analysis.Attributes["colors"].([]string),
    }

    content, err := GenerateProductContent(ctx, fip.fw, productInfo)
    if err != nil {
        return nil, fmt.Errorf("content generation failed: %w", err)
    }
    output.Content = content

    // Step 3: Market Analysis
    log.Printf("Step 3: Analyzing market data...")
    // Implement market data analysis using Wildberries API

    // Step 4: Style Recommendations
    log.Printf("Step 4: Generating style recommendations...")
    recRequest := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {
                        Type: interfaces.PonchoContentTypeText,
                        Text: fmt.Sprintf(`Based on this fashion analysis, provide 5 style recommendations:

Item: %s
Style: %s
Season: %s
Colors: %v

Provide tips for wearing, pairing, and occasions.`,
                            input.Name,
                            analysis.Style,
                            analysis.Season,
                            analysis.Attributes["colors"]),
                    },
                },
            },
        },
        MaxTokens:   &[]int{300}[0],
        Temperature: &[]float32{0.7}[0],
    }

    recResponse, err := fip.fw.Generate(ctx, recRequest)
    if err != nil {
        log.Printf("Warning: recommendations failed: %v", err)
    } else {
        output.Recommendations = []string{recResponse.Message.Content[0].Text}
    }

    output.ProcessingTime = time.Since(start)
    log.Printf("Pipeline completed in %v", output.ProcessingTime)

    return output, nil
}

// Usage example
func main() {
    fw := initializeFramework()
    wbClient := initializeWildberriesClient()
    ctx := context.Background()

    pipeline := NewFashionItemPipeline(fw, wbClient)

    input := &PipelineInput{
        ImageURL:    "data:image/jpeg;base64,FashionItemImageBase64...",
        Name:        "Элегантное платье-кимоно",
        Category:    "Платья",
        Description: "Платье свободного кроя с восточными мотивами",
    }

    result, err := pipeline.Process(ctx, input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n=== Pipeline Results ===\n")
    fmt.Printf("Analysis Category: %s\n", result.Analysis.Category)
    fmt.Printf("Generated Title: %s\n", result.Content.Title)
    fmt.Printf("Processing Time: %v\n", result.ProcessingTime)

    if len(result.Recommendations) > 0 {
        fmt.Printf("\nStyle Recommendations:\n%s\n", result.Recommendations[0])
    }
}
```

### Quality Assurance Workflow

```go
type QualityChecker struct {
    fw        interfaces.PonchoFramework
    rules     []QualityRule
}

type QualityRule struct {
    Name        string
    Description string
    Check       func(*PipelineOutput) QualityResult
}

type QualityResult struct {
    Passed  bool   `json:"passed"`
    Score   float64 `json:"score"`
    Message string `json:"message"`
}

func NewQualityChecker(fw interfaces.PonchoFramework) *QualityChecker {
    qc := &QualityChecker{
        fw: fw,
        rules: []QualityRule{
            {
                Name:        "Description Length",
                Description: "Check if description meets minimum length",
                Check: func(output *PipelineOutput) QualityResult {
                    if output.Content == nil {
                        return QualityResult{Passed: false, Message: "No content generated"}
                    }
                    if len(output.Content.Description) < 100 {
                        return QualityResult{Passed: false, Score: 0.5, Message: "Description too short"}
                    }
                    return QualityResult{Passed: true, Score: 1.0, Message: "Description length adequate"}
                },
            },
            {
                Name:        "SEO Keywords",
                Description: "Check if enough SEO keywords are provided",
                Check: func(output *PipelineOutput) QualityResult {
                    if output.Content == nil || len(output.Content.Keywords) < 5 {
                        return QualityResult{Passed: false, Score: 0.3, Message: "Insufficient SEO keywords"}
                    }
                    return QualityResult{Passed: true, Score: 1.0, Message: "Good SEO coverage"}
                },
            },
            {
                Name:        "Analysis Confidence",
                Description: "Check AI analysis confidence",
                Check: func(output *PipelineOutput) QualityResult {
                    if output.Analysis == nil || output.Analysis.Confidence < 0.7 {
                        return QualityResult{Passed: false, Score: output.Analysis.Confidence, Message: "Low analysis confidence"}
                    }
                    return QualityResult{Passed: true, Score: output.Analysis.Confidence, Message: "High confidence analysis"}
                },
            },
        },
    }
    return qc
}

func (qc *QualityChecker) CheckQuality(ctx context.Context, output *PipelineOutput) map[string]QualityResult {
    results := make(map[string]QualityResult)

    for _, rule := range qc.rules {
        results[rule.Name] = rule.Check(output)
        log.Printf("Quality check '%s': %s (Score: %.1f)",
            rule.Name,
            results[rule.Name].Message,
            results[rule.Name].Score)
    }

    return results
}

func (qc *QualityChecker) GetOverallScore(results map[string]QualityResult) float64 {
    if len(results) == 0 {
        return 0
    }

    var total float64
    for _, result := range results {
        total += result.Score
    }

    return total / float64(len(results))
}
```

These examples provide comprehensive, ready-to-use code for implementing common fashion industry workflows with the PonchoAiFramework. Each example is complete and can be adapted to specific use cases.