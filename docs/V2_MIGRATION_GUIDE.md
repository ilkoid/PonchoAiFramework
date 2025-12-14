# PonchoAiFramework v2.0 Migration Guide

## Overview

PonchoAiFramework v2.0 introduces major improvements for building complex AI workflows:

1. **FlowContext**: Shared state management between workflow steps
2. **MediaPipeline**: Automatic media conversion and processing
3. **FlowBuilder DSL**: Declarative workflow creation
4. **Enhanced Type Safety**: Generic result types and validation
5. **Parallel Execution**: Native support for parallel and conditional steps

## Key Improvements

### Problem Solved: State Management

**v1 (Manual State Passing)**:
```go
func (f *FashionFlow) Execute(input interface{}) (interface{}, error) {
    // Step 1: Import
    result1 := f.tool1.Execute(input)
    product := result1.(map[string]interface{})["product"]
    images := result1.(map[string]interface{})["images"]

    // Step 2: Analyze (needs images from step 1)
    features := f.analyzeImages(images)

    // Step 3: Classify (needs product from step 1)
    category := f.classify(product["name"])

    // Step 4: Generate (needs ALL previous results)
    prompt := fmt.Sprintf("Product: %v\nFeatures: %v\nCategory: %v",
                         product, features, category)
    return f.generate(prompt)
}
```

**v2 (FlowContext)**:
```go
flow := flow.NewFlowBuilder("fashion_enrichment").
    Step("import").Tool(importer, "s3_path").Output("product_data").Continue().
    Step("analyze").Model(vision, "prompt").Inputs("product_data.images").Output("features").Continue().
    Step("classify").Tool(classifier, "product_data.name").Output("category").Continue().
    Step("generate").Model(text, "prompt").
        Inputs("product_data", "features", "category").Output("description").Continue().
    Build()

// FlowContext automatically manages state between steps!
```

### Problem Solved: Media Processing

**v1 (Manual Conversion)**:
```go
// Manual base64 encoding for vision models
base64 := base64.StdEncoding.EncodeToString(imageBytes)
dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64)
```

**v2 (Automatic)**:
```go
// MediaPipeline handles conversion automatically
mediaPipeline := media.NewMediaPipeline(nil)
preparedMedia, _ := mediaPipeline.PrepareForModel(ctx, mediaList, model)
```

## Migration Steps

### Step 1: Update Flows to Use FlowContext

1. Import new packages:
```go
import (
    "github.com/ilkoid/PonchoAiFramework/core/context"
    "github.com/ilkoid/PonchoAiFramework/core/flow"
)
```

2. Replace old Execute method:
```go
// OLD
func (f *MyFlow) Execute(input interface{}) (interface{}, error) {
    // ... implementation
}

// NEW
func (f *MyFlow) Execute(ctx context.Context, input interface{}, flowCtx context.FlowContext) (interface{}, error) {
    // ... implementation
}
```

3. Or use FlowBuilder for new flows:
```go
flow := flow.NewFlowBuilder("my_flow").
    Step("step1").Tool(tool1, "input").Output("output1").Continue().
    Step("step2").Tool(tool2, "output1").Output("final").Continue().
    Build()
```

### Step 2: Update Tool Integration

**OLD**:
```go
result, err := tool.Execute(input)
processedData := result.(map[string]interface{})["data"]
```

**NEW**:
```go
// Store in FlowContext
result, err := tool.Execute(input)
if err == nil {
    flowCtx.Set("tool_result", result)
}

// Retrieve from FlowContext
toolResult, has := flowCtx.Get("tool_result")
```

### Step 3: Add Media Processing Support

1. Create MediaPipeline:
```go
mediaPipeline := media.NewMediaPipeline(nil)
```

2. Use MediaHelper for vision models:
```go
helper := common.NewMediaHelper(mediaPipeline, logger)

// Prepare media for model
content, err := helper.PrepareMediaContent(ctx, mediaList, model)
```

### Step 4: Implement Parallel Execution

**NEW**:
```go
flow := flow.NewFlowBuilder("parallel_flow").
    Step("parallel_step").
    Parallel().
    MaxConcurrency(3).
    AddSubStep(analyzeImageStep).
    AddSubStep(analyzeTextStep).
    AddSubStep(classifyStep).
    Continue().
    Build()
```

## API Changes

### Core Interfaces

#### FlowContext
```go
type FlowContext interface {
    Set(key string, value interface{}) error
    Get(key string) (interface{}, bool)
    SetBytes(key string, value []byte) error
    GetBytes(key string) ([]byte, error)
    SetMedia(key string, media *MediaData) error
    GetMedia(key string) (*MediaData, error)
    // ... more methods
}
```

#### PonchoFlowV2
```go
type PonchoFlowV2 interface {
    Execute(ctx context.Context, input interface{}, flowCtx FlowContext) (interface{}, error)
    CreateContext() FlowContext
    ValidateContext(flowCtx FlowContext) error
    // ... more methods
}
```

### Media Pipeline

#### MediaData
```go
type MediaData struct {
    URL      string      `json:"url,omitempty"`
    Bytes    []byte      `json:"-"`
    MimeType string      `json:"mime_type"`
    Format   MediaFormat `json:"format"`
    Size     int64       `json:"size"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

## Best Practices

### 1. Use Type-Safe Context Operations
```go
// Good
productName, err := flowCtx.GetString("product_name")
if err != nil {
    return fmt.Errorf("product name required: %w", err)
}

// Avoid (runtime error prone)
productName := flowCtx.Get("product_name").(string)
```

### 2. Organize Context Keys
```go
// Use prefixes for related data
flowCtx.Set("product_basic_info", productInfo)
flowCtx.Set("product_images", images)
flowCtx.Set("product_analysis", analysis)

// Use clear keys for results
flowCtx.Set("vision_analysis_result", features)
flowCtx.Set("categorization_result", category)
```

### 3. Handle Media Efficiently
```go
// Use MediaPipeline for automatic conversion
preparedMedia, err := mediaPipeline.PrepareForModel(ctx, rawMedia, model)

// Cache expensive operations
if cached, found := mediaCache.Get(cacheKey); found {
    return cached, nil
}
```

### 4. Use FlowBuilder for Complex Workflows
```go
flow := flow.NewFlowBuilder("complex_workflow").
    Description("Processes fashion products with AI analysis").
    Timeout(10 * time.Minute).
    RequiresVision().
    EnableParallel(true).
    MaxConcurrency(5).
    Step("import").Tool(importer, "path").Output("data").Continue().
    Step("analyze").Parallel().
        AddSubStep(analyzeImages).
        AddSubStep(analyzeText).
        MaxConcurrency(3).
        Continue().
    Step("generate").Model(llm, "prompt").Inputs("data", "analysis").Output("result").Continue().
    Build()
```

## Migration Checklist

- [ ] Update flow imports to include `core/context` and `core/flow`
- [ ] Replace `Execute(input)` with `Execute(ctx, input, flowCtx)`
- [ ] Add FlowContext creation and management
- [ ] Update tool result handling to use FlowContext
- [ ] Add MediaPipeline for vision model integrations
- [ ] Consider using FlowBuilder for new flows
- [ ] Add type-safe context operations
- [ ] Update error handling for FlowContext operations
- [ ] Add context validation where appropriate
- [ ] Update tests to use FlowContext

## Backward Compatibility

v2.0 maintains backward compatibility through adapters:

```go
// Adapter to use v1 flows in v2 system
type V1FlowAdapter struct {
    v1Flow interfaces.PonchoFlow
}

func (adapter *V1FlowAdapter) Execute(ctx context.Context, input interface{}, flowCtx FlowContext) (interface{}, error) {
    return adapter.v1Flow.Execute(input)
}
```

## Example: Complete Migration

### Before (v1):
```go
type FashionProcessor struct {
    importer interfaces.PonchoTool
    vision  interfaces.PonchoModel
    classifier interfaces.PonchoTool
    generator interfaces.PonchoModel
}

func (fp *FashionProcessor) Execute(input interface{}) (interface{}, error) {
    // Import
    result, _ := fp.importer.Execute(input)
    data := result.(map[string]interface{})

    // Vision analysis
    images := data["images"].([]interface{})
    var analyses []string
    for _, img := range images {
        // Manual image processing...
        analysis := fp.analyzeImage(img)
        analyses = append(analyses, analysis)
    }

    // Classification
    name := data["name"].(string)
    category, _ := fp.classifier.Execute(name)

    // Generation
    prompt := fmt.Sprintf("Product: %s\nImages: %v\nCategory: %v",
                         name, analyses, category)
    return fp.generator.Execute(prompt)
}
```

### After (v2):
```go
func NewFashionProductFlow(importer interfaces.PonchoTool, vision interfaces.PonchoModel,
                         classifier interfaces.PonchoTool, generator interfaces.PonchoModel) interfaces.PonchoFlowV2 {

    return flow.NewFlowBuilder("fashion_enrichment").
        Description("Enriches fashion products with AI analysis").
        Version("2.0.0").
        Category("fashion").
        RequiresVision().
        Timeout(10 * time.Minute).

        Step("import").Tool(importer, "s3_path").
            Output("product_data").
            Timeout(30 * time.Second).
            Continue().

        Step("analyze_vision").Model(vision, "vision_prompt").
            Inputs("product_data.images").
            Output("visual_features").
            MaxTokens(500).
            Timeout(60 * time.Second).
            Continue().

        Step("classify").Tool(classifier, "product_data.name").
            Output("category_id").
            Timeout(15 * time.Second).
            Continue().

        Step("generate_description").Model(generator, "description_prompt").
            Inputs("product_data", "visual_features", "category_id").
            Output("seo_description").
            Temperature(0.7).
            MaxTokens(1000).
            Continue().

        Build()
}
```

## Support

For migration assistance:
1. Check the examples in `/examples/` directory
2. Review API documentation
3. Open an issue with specific migration questions
4. Use the adapter pattern for gradual migration

The v2.0 architecture provides significant improvements for complex workflows while maintaining the simplicity of the original framework for simple use cases.