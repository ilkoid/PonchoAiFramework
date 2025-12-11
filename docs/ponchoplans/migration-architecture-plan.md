# Migration Architecture Plan: GenKit → PonchoFramework

## Executive Summary

Документ описывает детальный план миграции с Firebase GenKit на собственный PonchoFramework. План обеспечивает плавный переход с минимальными рисками и полной сохранностью функциональности.

## Current State Analysis

### Existing GenKit Dependencies

#### Core Components Using GenKit
```
mini-agent/
├── mini_agent.go           # MiniAgent struct with *genkit.Genkit
├── flow.go                # MiniAgentFlow with genkit.Generate()
├── prompt_executor.go      # PromptExecutor with AI integration
└── dotprompt_manager.go    # DotPromptManager with genkit.LookupPrompt()

models/
├── deepseek.go            # DeepSeek model registration
├── glm.go                 # Z.AI model registration
└── vision.go              # Vision analyzer with AI interfaces

tools/
├── article_importer.go     # genkit.DefineTool()
├── bucket_browser.go       # genkit.DefineTool()
├── parent_categories.go    # genkit.DefineTool()
├── category_subjects.go    # genkit.DefineTool()
└── subject_characteristics.go # genkit.DefineTool()

flows/
└── article_importer.go     # Flow with genkit.Generate()

cmd/
├── poncho-mini/           # CLI apps using GenKit
├── test_multimodal_prompt/ # Testing with GenKit
└── [other CLI apps]       # Various GenKit integrations
```

#### Key GenKit Patterns Used
1. **Model Registration**: `genkit.DefineModel(g, name, options, handler)`
2. **Tool Registration**: `genkit.DefineTool(g, name, description, handler)`
3. **Generation**: `genkit.Generate(ctx, g, ai.WithPrompt(), ai.WithTools())`
4. **Prompt Management**: `genkit.LookupPrompt(g, name)`
5. **Flow Definition**: `genkit.DefineFlow(g, name, handler)`

## Migration Architecture

### Phase 1: Foundation (Week 1-2)

#### 1.1 Core Framework Creation
```
poncho-framework/
├── core/
│   ├── framework.go          # PonchoFramework main class
│   ├── interfaces.go         # Core interfaces (Model, Tool, Flow)
│   ├── registry.go           # Component registries
│   └── types.go             # Core types and structures
├── config/
│   ├── framework_config.go    # Configuration management
│   └── model_config.go      # Model-specific configs
└── errors/
    └── framework_errors.go   # Custom error types
```

#### 1.2 Basic Interfaces
```go
// Core interfaces that will replace GenKit abstractions
type PonchoModel interface {
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    SupportsStreaming() bool
    SupportsTools() bool
    SupportsVision() bool
}

type PonchoTool interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    Name() string
    Description() string
    InputSchema() map[string]interface{}
}

type PonchoFlow interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    Name() string
    InputSchema() map[string]interface{}
}
```

### Phase 2: Model Adapters (Week 3-4)

#### 2.1 Model Registry Implementation
```
poncho-framework/models/
├── registry.go              # Model registry
├── deepseek_adapter.go      # DeepSeek model adapter
├── zai_adapter.go          # Z.AI model adapter
├── base_model.go           # Base model implementation
└── request_builder.go      # Request/response builders
```

#### 2.2 Migration Strategy for Models
```go
// Before (GenKit):
genkit.DefineModel(g, "deepseek/chat", options, handler)

// After (PonchoFramework):
ponchoFramework.RegisterModel("deepseek/chat", &PonchoDeepSeekModel{
    apiKey: config.DeepSeek.APIKey,
    baseURL: "https://api.deepseek.com",
})
```

#### 2.3 Model Adapter Pattern
```go
type PonchoDeepSeekModel struct {
    apiKey   string
    client   *deepseek.Client
    config   *PonchoModelConfig
}

func (m *PonchoDeepSeekModel) Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error) {
    // Convert PonchoModelRequest to DeepSeek format
    // Use existing logic from models/deepseek.go
    // Convert response back to PonchoModelResponse
}
```

### Phase 3: Tool Migration (Week 5-6)

#### 3.1 Tool Registry Implementation
```
poncho-framework/tools/
├── registry.go              # Tool registry
├── base_tool.go            # Base tool implementation
├── article_importer.go     # Migrated article importer
├── bucket_browser.go       # Migrated bucket browser
├── wildberries_tools.go    # Migrated WB tools
└── validation.go           # Tool validation
```

#### 3.2 Tool Migration Pattern
```go
// Before (GenKit):
genkit.DefineTool(g, "importArticleData", description, func(ctx *ai.ToolContext, input ArticleImporterInput) (ArticleImporterOutput, error) {
    // Tool implementation
})

// After (PonchoFramework):
type PonchoArticleImporterTool struct {
    s3Client *s3.S3Client
    visionAnalyzer PonchoModel
}

func (t *PonchoArticleImporterTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Convert input to ArticleImporterInput
    // Use existing tool logic
    // Return ArticleImporterOutput
}

// Registration:
ponchoFramework.RegisterTool("importArticleData", &PonchoArticleImporterTool{...})
```

### Phase 4: Flow Integration (Week 7-8)

#### 4.1 Flow Manager Implementation
```
poncho-framework/flows/
├── manager.go               # Flow manager
├── base_flow.go            # Base flow implementation
├── article_importer.go     # Migrated article importer flow
├── mini_agent.go           # Migrated mini-agent flow
└── orchestration.go        # Flow orchestration
```

#### 4.2 Flow Migration Pattern
```go
// Before (GenKit):
func DefineMiniAgentFlow(g *genkit.Genkit, ...) *MiniAgentFlow {
    return &MiniAgentFlow{
        genkit: g,
        // ... other dependencies
    }
}

// After (PonchoFramework):
type PonchoMiniAgentFlow struct {
    framework *PonchoFramework
    // ... other dependencies
}

func (f *PonchoMiniAgentFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Use PonchoFramework.Generate() instead of genkit.Generate()
    // Use registered tools instead of GenKit tool references
}
```

### Phase 5: Prompt System (Week 9-10)

#### 5.1 Prompt Manager Implementation
```
poncho-framework/prompts/
├── manager.go               # Prompt manager
├── loader.go                # Prompt file loader
├── parser.go                # Enhanced prompt parser
├── builder.go               # Prompt builder
└── cache.go                 # Prompt cache
```

#### 5.2 Prompt Migration Strategy
```go
// Before (GenKit):
prompt := genkit.LookupPrompt(g, "article_analysis/data_import")
response, err := prompt.Execute(ctx, ai.WithInput(input))

// After (PonchoFramework):
prompt, err := ponchoFramework.GetPrompt("article_analysis/data_import")
response, err := ponchoFramework.ExecutePrompt(ctx, prompt.Name, input)
```

## Risk Mitigation

### Technical Risks
1. **API Compatibility**: Differences in request/response formats
   - *Mitigation*: Comprehensive adapter testing
2. **Performance Degradation**: New framework might be slower
   - *Mitigation*: Performance benchmarks and optimization
3. **Feature Gaps**: Some GenKit features might be missing
   - *Mitigation*: Feature parity analysis and implementation

### Project Risks
1. **Timeline Delays**: Complex migration might take longer
   - *Mitigation*: Phased approach with incremental delivery
2. **Resource Allocation**: Team might be stretched thin
   - *Mitigation*: Prioritize critical path features
3. **Business Impact**: Migration might affect production
   - *Mitigation*: Parallel development and gradual rollout

## Testing Strategy

### Unit Testing
- core components only
- Mock external dependencies
- Test edge cases

### Integration Testing
- Test component interactions by small test utilities using real data

---

*Этот документ будет обновляться по мере прогресса миграции и должен служить руководством для всей команды разработки.*