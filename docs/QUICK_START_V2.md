# PonchoAiFramework v2.0 Quick Start

## Introduction

PonchoAiFramework v2.0 introduces revolutionary improvements for building complex AI workflows:

- **FlowContext**: Shared state management between workflow steps
- **MediaPipeline**: Automatic media conversion for vision models
- **FlowBuilder DSL**: Declarative workflow creation
- **Parallel Execution**: Native support for parallel processing
- **Type Safety**: Enhanced error handling and validation

## Installation

```bash
go get github.com/ilkoid/PonchoAiFramework@v2.0.0
```

## Quick Example: Fashion Product Analysis

Let's solve the original problem - analyzing fashion products with multiple AI models:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ilkoid/PonchoAiFramework/core/flow"
    "github.com/ilkoid/PonchoAiFramework/core/context"
    "github.com/ilkoid/PonchoAiFramework/interfaces"
)

func main() {
    // 1. Initialize your components (tools, models)
    articleImporter := createArticleImporter()
    visionModel := createVisionModel()
    categoryTool := createCategoryTool()
    textModel := createTextModel()

    // 2. Build the flow with FlowBuilder DSL
    fashionFlow := flow.NewFlowBuilder("fashion_enrichment").
        Description("AI-powered fashion product enrichment").
        Timeout(10 * time.Minute).
        RequiresVision().

        // Step 1: Import product data from S3
        Step("import").Tool(articleImporter, "s3_path").
            Output("product_data").
            Timeout(30 * time.Second).
            Continue().

        // Step 2: Analyze product images with vision model
        Step("analyze_images").Model(visionModel, "vision_prompt").
            Input("product_data.images").
            Output("visual_features").
            MaxTokens(500).
            Continue().

        // Step 3: Classify product category
        Step("classify").Tool(categoryTool, "product_data.name").
            Output("category_id").
            Timeout(15 * time.Second).
            Continue().

        // Step 4: Generate SEO description using ALL previous data
        Step("generate_description").Model(textModel, "description_prompt").
            Inputs("product_data", "visual_features", "category_id").
            Output("seo_description").
            Temperature(0.7).
            MaxTokens(1000).
            Continue().

        Build() // Build the flow

    // 3. Execute the flow
    ctx := context.Background()
    flowCtx := context.NewBaseFlowContext()

    result, err := fashionFlow.Execute(ctx, "s3://products/product-123.json", flowCtx)
    if err != nil {
        log.Fatal("Flow execution failed:", err)
    }

    // 4. Get results from context
    if categoryID, _ := flowCtx.GetString("category_id"); categoryID != "" {
        fmt.Printf("Product Category: %s\n", categoryID)
    }

    if description, _ := flowCtx.GetString("seo_description"); description != "" {
        fmt.Printf("SEO Description: %s\n", description)
    }
}
```

## Key Concepts

### 1. FlowContext - Shared State

FlowContext acts as shared memory for your workflow:

```go
// Any step can write to context
flowCtx.Set("product_name", "Summer Dress")
flowCtx.Set("images", imageBytes)

// Any step can read from context
name, _ := flowCtx.GetString("product_name")
images, _ := flowCtx.GetBytes("images")

// Type-safe operations
price, err := flowCtx.GetFloat("price")
if err != nil {
    return fmt.Errorf("price is required: %w", err)
}
```

### 2. FlowBuilder DSL - Declarative Workflows

Build complex workflows with a fluent API:

```go
flow := flow.NewFlowBuilder("my_workflow").
    Description("My AI workflow").
    Timeout(5 * time.Minute).

    // Tool execution
    Step("step1").Tool(myTool, "input_key").Output("result1").Continue().

    // Model execution
    Step("step2").Model(myModel, "prompt").
        Inputs("result1", "additional_data").
        Temperature(0.7).
        MaxTokens(1000).
        Output("result2").
        Continue().

    // Conditional execution
    Step("step3").
        Conditional(func(ctx context.FlowContext) bool {
            return ctx.Has("optional_data")
        }).
        True(stepIfTrue).
        False(stepIfFalse).
        Continue().

    // Parallel execution
    Step("parallel_analysis").
        Parallel().
        MaxConcurrency(3).
        AddSubStep(analyzeImages).
        AddSubStep(analyzeText).
        FailFast(false).
        Continue().

    Build()
```

### 3. MediaPipeline - Automatic Media Processing

No more manual base64 encoding:

```go
// Create media pipeline
mediaPipeline := media.NewMediaPipeline(nil)

// Automatically convert images for vision models
preparedMedia, err := mediaPipeline.PrepareForModel(ctx, mediaList, visionModel)

// Use in model request
message, _ := helper.CreateVisionMessage(ctx, "Analyze this image", mediaList, visionModel)
```

## Common Patterns

### Pattern 1: Sequential Processing

```go
flow := flow.NewFlowBuilder("sequential").
    Step("validate").Tool(validator, "input").Output("validated").Continue().
    Step("process").Tool(processor, "validated").Output("processed").Continue().
    Step("store").Tool(storage, "processed").Continue().
    Build()
```

### Pattern 2: Data Accumulation

```go
flow := flow.NewFlowBuilder("data_accumulation").
    Step("collect_data").Tool(collector, "source").Output("data_batch").Continue().
    Step("process_batch").
        Custom(func(ctx context.Context, flowCtx context.FlowContext) error {
            // Accumulate multiple data points
            data, _ := flowCtx.Get("data_batch")
            for _, item := range data.([]interface{}) {
                processed := processItem(item)
                flowCtx.AppendToArray("processed_items", processed)
            }
            return nil
        }).
        Continue().
    Build()
```

### Pattern 3: Parallel Analysis

```go
flow := flow.NewFlowBuilder("parallel_analysis").
    Step("analyze").Parallel().
        MaxConcurrency(5).
        AddSubStep(analyzeImage).
        AddSubStep(analyzeText).
        AddSubStep(analyzeMetadata).
        FailFast(false).
        Continue().
    Build()
```

### Pattern 4: Conditional Logic

```go
flow := flow.NewFlowBuilder("conditional").
    Step("check_premium").
        Conditional(func(ctx context.FlowContext) bool {
            isPremium, _ := ctx.GetBool("is_premium")
            return isPremium
        }).
        True(premiumAnalysis).
        False(basicAnalysis).
        Continue().
    Build()
```

## Working with Tools

### Tool Implementation

```go
type MyTool struct {
    name string
}

func (t *MyTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Process input
    result := process(input)

    // Return result
    return map[string]interface{}{
        "success": true,
        "data":    result,
    }, nil
}

func (t *MyTool) Name() string { return t.name }
func (t *MyTool) Description() string { return "My custom tool" }
// ... other interface methods
```

### Using Tools in Flows

```go
// Simple usage
Step("process").Tool(myTool, "input_data").Output("result").Continue()

// With error handling
Step("process").Tool(myTool, "input_data").
    Output("result").
    CanFail(true). // Don't fail entire flow if tool fails
    Timeout(30 * time.Second).
    Continue().

// With multiple inputs
Step("process").Tool(myTool, "primary_input").
    Input("secondary_input").
    Output("combined_result").
    Continue()
```

## Working with Models

### Model Integration

```go
// Text generation
Step("generate").Model(llm, "prompt_key").
    Output("generated_text").
    Temperature(0.7).
    MaxTokens(1000).
    Continue().

// Vision model with images
Step("analyze").Model(visionModel, "prompt").
    Input("images").
    WithMedia("image1", "image2").
    Output("analysis").
    Continue().

// Multiple inputs
Step("summarize").Model(llm, "prompt").
    Inputs("data1", "data2", "data3").
    Output("summary").
    Continue()
```

## Error Handling

### Step-Level Error Handling

```go
// Allow step to fail without stopping flow
Step("optional_step").Tool(optionalTool, "input").
    CanFail(true).
    Continue().

// Retry on failure
Step("important_step").Tool(importantTool, "input").
    Timeout(60 * time.Second).
    Continue().

// Handle errors in custom steps
Step("custom_step").
    Custom(func(ctx context.Context, flowCtx context.FlowContext) error {
        if err := riskyOperation(); err != nil {
            // Log error but don't fail step
            flowCtx.Set("step_error", err.Error())
            return nil
        }
        return nil
    }).
    Continue()
```

### Context Validation

```go
// Require specific context keys
Step("dependent_step").Tool(tool, "input").
    Requires("required_data", "config").
    Continue().

// Validate context before execution
if err := flow.ValidateContext(flowCtx); err != nil {
    return fmt.Errorf("invalid context: %w", err)
}
```

## Testing Flows

### Unit Testing Steps

```go
func TestToolStep(t *testing.T) {
    // Create mock flow context
    flowCtx := context.NewBaseFlowContext()
    flowCtx.Set("test_input", "test_data")

    // Create step
    step := &flow.ToolStep{
        name:     "test_step",
        tool:     mockTool,
        inputKey: "test_input",
        outputKey: "result",
    }

    // Execute step
    err := step.Execute(context.Background(), flowCtx)

    // Assert results
    assert.NoError(t, err)
    result, exists := flowCtx.Get("result")
    assert.True(t, exists)
    assert.NotNil(t, result)
}
```

### Integration Testing

```go
func TestFashionFlow(t *testing.T) {
    // Build flow with test components
    flow := buildTestFlow()

    // Create context with test data
    flowCtx := context.NewBaseFlowContext()
    flowCtx.Set("s3_path", "test://product.json")

    // Execute flow
    result, err := flow.Execute(context.Background(), "test_input", flowCtx)

    // Verify results
    assert.NoError(t, err)

    // Check context state
    category, _ := flowCtx.GetString("category_id")
    assert.NotEmpty(t, category)

    description, _ := flowCtx.GetString("seo_description")
    assert.NotEmpty(t, description)
}
```

## Best Practices

### 1. Use Type-Safe Operations
```go
// Good
price, err := flowCtx.GetFloat("price")
if err != nil {
    return fmt.Errorf("invalid price: %w", err)
}

// Avoid
price := flowCtx.Get("price").(float64) // Runtime panic risk
```

### 2. Organize Context Keys
```go
// Use prefixes for organization
flowCtx.Set("product_basic_info", info)
flowCtx.Set("product_images", images)
flowCtx.Set("product_analysis", analysis)

// Use descriptive keys
flowCtx.Set("vision_analysis_result", features)
flowCtx.Set("category_classification", category)
```

### 3. Handle Resources Properly
```go
// Set appropriate timeouts
Step("vision_analysis").Model(visionModel, "prompt").
    Timeout(60 * time.Second). // Vision models can be slow
    Continue().

// Use parallel execution for I/O bound tasks
Step("parallel_downloads").
    Parallel().
    MaxConcurrency(10).
    Continue()
```

### 4. Add Context Validation
```go
func (f *MyFlow) ValidateContext(flowCtx context.FlowContext) error {
    requiredKeys := []string{"product_data", "images", "config"}
    for _, key := range requiredKeys {
        if !flowCtx.Has(key) {
            return fmt.Errorf("required key '%s' missing", key)
        }
    }
    return nil
}
```

## Next Steps

1. **Explore Examples**: Check `/examples/` for complete implementations
2. **Read Migration Guide**: If upgrading from v1.0, see `V2_MIGRATION_GUIDE.md`
3. **API Documentation**: Detailed API docs available in `/docs/api/`
4. **Community**: Join discussions for patterns and best practices

## Need Help?

- Check the examples directory for complete use cases
- Review the migration guide for v1.0 → v2.0 transitions
- Open an issue for specific questions
- Join our community discussions

Welcome to PonchoAiFramework v2.0 - where building complex AI workflows is simple and elegant!