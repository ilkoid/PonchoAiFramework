# Architecture

## System Architecture

PonchoFramework follows a clean architecture pattern with clear separation of concerns between core framework, model/tool/flow implementations, and infrastructure layers.

## Directory Structure

```
/home/ilkoid/go-workspace/src/PonchoAiFramework/
├── core/                    # Core framework components (✅ IMPLEMENTED)
│   ├── framework.go         # Main PonchoFramework class ✅
│   ├── interfaces.go        # Core interfaces ✅
│   ├── base/                # Base implementations ✅
│   │   ├── model.go         # PonchoBaseModel ✅
│   │   ├── tool.go          # PonchoBaseTool ✅
│   │   └── flow.go          # PonchoBaseFlow ✅
│   ├── registry/            # Registry implementations ✅
│   │   ├── model_registry.go ✅
│   │   ├── tool_registry.go ✅
│   │   └── flow_registry.go ✅
│   ├── config/              # Configuration system ✅
│   │   ├── config.go        # Config manager ✅
│   │   ├── loader.go        # Config loader ✅
│   │   └── validator.go    # Config validator ✅
│   ├── logger.go           # Logging system ✅
│   └── framework_test.go   # Framework tests ✅
├── interfaces/               # Core interfaces and types (✅ IMPLEMENTED)
│   ├── interfaces.go        # Interface aliases ✅
│   ├── types.go           # Type definitions ✅
│   └── logger.go          # Logger interface ✅
├── models/                  # Model implementations (NOT IMPLEMENTED YET)
│   ├── deepseek/            # DeepSeek integration
│   │   ├── client.go
│   │   ├── model.go
│   │   └── streaming.go
│   ├── zai/                 # Z.AI integration
│   │   ├── client.go
│   │   ├── model.go
│   │   ├── vision.go
│   │   └── streaming.go
│   └── common/              # Shared model utilities
│       ├── request.go
│       ├── response.go
│       └── converter.go
├── tools/                   # Tool implementations (NOT IMPLEMENTED YET)
│   ├── article_importer/
│   ├── bucket_browser/
│   ├── wildberries/
│   └── common/              # Shared tool utilities
├── flows/                   # Flow implementations (NOT IMPLEMENTED YET)
│   ├── article_importer/
│   ├── mini_agent/
│   └── common/              # Shared flow utilities
├── prompts/                 # Prompt management (✅ IMPLEMENTED - Phase 5 Complete)
│   ├── manager.go          # Main prompt manager ✅
│   ├── types.go           # Prompt system types and extensions ✅
│   ├── parser.go           # Template loader with V1 format support ✅
│   ├── executor.go         # Template execution engine ✅
│   ├── validator.go        # Template validation system ✅
│   └── cache.go           # LRU cache implementation ✅
├── docs/                    # Documentation (EXISTS)
│   ├── README.md
│   ├── poncho-framework-design.md
│   ├── core-interfaces-specification.md
│   ├── implementation-strategy.md
│   └── api-documentation.md
├── sample/                  # Sample reference code from old Poncho Tools (EXISTS)
│   └── poncho-tools/        # Legacy implementation for reference
│       ├── mini-agent/      # Mini-agent flow examples
│       ├── s3/              # S3 integration patterns
│       ├── wildberries/     # Wildberries API tools
│       └── cmd/             # Example applications
├── config.yaml              # Main configuration file (EXISTS)
├── go.mod                   # Go module definition (EXISTS)
└── .env.example            # Environment variables template (EXISTS)
```

## Core Components

### 1. PonchoFramework (Main Orchestrator)

**Location:** `core/framework.go` (✅ IMPLEMENTED)

**Responsibilities:**
- Central registry for models, tools, and flows
- Request routing and orchestration
- Configuration management
- Lifecycle management (Start/Stop)
- Metrics collection

**Key Methods:**
- `NewPonchoFramework(config *PonchoFrameworkConfig) *PonchoFramework`
- `Start(ctx context.Context) error`
- `Stop(ctx context.Context) error`
- `Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)`
- `ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)`
- `ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error)`

### 2. Model Registry

**Location:** `core/registry/model_registry.go` (✅ IMPLEMENTED)

**Responsibilities:**
- Store and manage registered models
- Model lifecycle management
- Model lookup and validation
- Thread-safe access to models

**Interface:**
```go
type PonchoModelRegistry interface {
    Register(name string, model PonchoModel) error
    Get(name string) (PonchoModel, error)
    List() []string
    Unregister(name string) error
    Clear() error
}
```

### 3. Tool Registry

**Location:** `core/registry/tool_registry.go` (✅ IMPLEMENTED)

**Responsibilities:**
- Store and manage registered tools
- Tool dependency resolution
- Category-based tool listing
- Input/output schema validation

**Interface:**
```go
type PonchoToolRegistry interface {
    Register(name string, tool PonchoTool) error
    Get(name string) (PonchoTool, error)
    List() []string
    ListByCategory(category string) []string
    Unregister(name string) error
    Clear() error
}
```

### 4. Flow Registry

**Location:** `core/registry/flow_registry.go` (✅ IMPLEMENTED)

**Responsibilities:**
- Store and manage registered flows
- Flow dependency validation
- Flow orchestration support
- Parallel execution coordination

## Key Design Patterns

### 1. Registry Pattern
All components (models, tools, flows) are registered in centralized registries for easy lookup and management.

### 2. Strategy Pattern
Models, tools, and flows implement common interfaces, allowing them to be swapped without changing client code.

### 3. Builder Pattern
Complex objects like `PonchoModelRequest` use builder pattern for cleaner construction.

### 4. Observer Pattern
Metrics collection and logging use observer pattern to monitor framework events without tight coupling.

### 5. Factory Pattern
Model, tool, and flow creation uses factory functions for consistent initialization.

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                      │
│  ┌─────────────┬─────────────┬─────────────────────┐ │
│  │   CLI Apps  │   Web API   │   Background Jobs   │ │
│  └─────────────┴─────────────┴─────────────────────┘ │
│                         │                           │
│                         ▼                           │
├─────────────────────────────────────────────────────────────┤
│                PonchoFramework Core                    │
│  ┌──────────────────────────────────────────────────┐ │
│  │         Request Router & Orchestrator          │ │
│  └──────────────────────────────────────────────────┘ │
│         │              │              │              │
│         ▼              ▼              ▼              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  Model   │  │   Tool   │  │   Flow   │  │  Prompt  │ │
│  │ Registry │  │ Registry │  │ Registry │  │ Manager  │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
├─────────────────────────────────────────────────────────────┤
│                Implementation Layer                    │
│  ┌─────────────┬─────────────┬─────────────────────┬─────────────┐ │
│  │   Models    │    Tools    │       Flows         │   Prompts    │ │
│  │             │             │                     │             │ │
│  │ DeepSeek    │ S3 Import   │ Article Importer    │ V1 Parser   │ │
│  │ Z.AI (GLM)  │ WB API      │ Mini-Agent          │ Templates   │ │
│  │ Vision      │ Vision      │ Description Gen     │ Validator   │ │
│  │ Custom      │ Custom      │ Custom              │ Cache       │ │
│  └─────────────┴─────────────┴─────────────────────┴─────────────┘ │
│         │              │              │              │              │
│         ▼              ▼              ▼              ▼              │
├─────────────────────────────────────────────────────────────┤
│              Infrastructure Layer                      │
│  ┌─────────────┬─────────────┬─────────────────────┐ │
│  │     S3      │ Wildberries │     Config          │ │
│  │   Storage   │    API      │   Management        │ │
│  │  (Yandex)   │             │    (YAML)           │ │
│  └─────────────┴─────────────┴─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Interface Specifications

### PonchoModel Interface
**Purpose:** Unified interface for all AI models
- Generate/GenerateStreaming methods
- Capability methods (Streaming, Tools, Vision, SystemRole)
- Metadata methods (Name, Provider, MaxTokens, Temperature)

### PonchoTool Interface
**Purpose:** Unified interface for all tools
- Identity methods (Name, Description, Version)
- Execute method with input/output
- Schema validation methods
- Metadata methods (Category, Tags, Dependencies)

### PonchoFlow Interface
**Purpose:** Unified interface for workflows
- Identity methods (Name, Description, Version)
- Execute/ExecuteStreaming methods
- Schema validation methods
- Metadata methods (Category, Tags, Dependencies)

## Configuration Architecture

### Configuration File Structure

**Location:** `config.yaml`

```yaml
models:
  deepseek-chat:
    provider: "deepseek"
    model_name: "deepseek-chat"
    api_key: "${DEEPSEEK_API_KEY}"
    max_tokens: 4000
    temperature: 0.7
    supports:
      vision: false
      tools: true
      stream: true
      system: true

  glm-vision:
    provider: "zai"
    model_name: "glm-4.6v"
    api_key: "${ZAI_API_KEY}"
    max_tokens: 2000
    temperature: 0.5
    supports:
      vision: true
      tools: true
      stream: true
      system: true

tools:
  article_importer:
    enabled: true
    timeout: 30s
    retry:
      max_attempts: 3
      backoff: "exponential"
      base_delay: 1s

  wb_categories:
    enabled: true
    timeout: 15s

flows:
  article_processor:
    enabled: true
    timeout: 120s
    dependencies:
      - "article_importer"
      - "wb_categories"
      - "glm-vision"

logging:
  level: "info"
  format: "text"

s3:
  url: "https://storage.yandexcloud.net"
  region: "ru-central1"
  bucket: "plm-ai"
  endpoint: "storage.yandexcloud.net"
  use_ssl: true

wildberries:
  base_url: "https://content-api.wildberries.ru"
  timeout: 30
```

## Integration Points

### External Services

1. **S3 Storage (Yandex Cloud)**
   - Purpose: Store and retrieve article data and images
   - Access: Via MinIO-compatible S3 client
   - Authentication: S3_ACCESS_KEY, S3_SECRET_KEY

2. **Wildberries API**
   - Purpose: Fetch categories, subjects, characteristics
   - Access: REST API
   - Authentication: WB_API_CONTENT_KEY

3. **DeepSeek API**
   - Purpose: Text generation and reasoning
   - Access: OpenAI-compatible API
   - Authentication: DEEPSEEK_API_KEY

4. **Z.AI GLM API**
   - Purpose: Vision analysis and multimodal tasks
   - Access: Custom API
   - Authentication: ZAI_API_KEY

### Reference Implementation

Sample code from existing Poncho Tools (GenKit-based) is available in `sample/poncho-tools/` for reference:

- **S3 Integration:** `sample/poncho-tools/s3/client.go`
- **Wildberries Tools:** `sample/poncho-tools/wildberries/`
- **Mini-Agent Flow:** `sample/poncho-tools/mini-agent/`
- **Vision Processing:** `sample/poncho-tools/cmd/simple_glm5v_processor/`

## Migration Strategy

### Phase 1: Core Framework
1. Implement core interfaces and base classes
2. Create registry implementations
3. Build configuration system
4. Establish testing framework

### Phase 2: Model Integration
1. Port DeepSeek model from existing code
2. Port Z.AI GLM model with vision support
3. Implement model adapters
4. Add streaming support

### Phase 3: Tool Migration
1. Migrate S3 tools
2. Migrate Wildberries tools
3. Add vision analysis tools
4. Implement tool validation

### Phase 4: Flow Integration
1. Migrate article importer flow
2. Migrate mini-agent flow
3. Add flow orchestration
4. Implement error handling

## Testing Architecture

### Test Structure
```
tests/
├── unit/                 # Unit tests for components
│   ├── core/
│   ├── models/
│   ├── tools/
│   └── flows/
├── integration/          # Integration tests
│   ├── model_integration/
│   ├── tool_integration/
│   └── flow_integration/
├── benchmark/            # Performance benchmarks
└── e2e/                  # End-to-end tests
```

### Coverage Requirements
- Minimum 90% code coverage
- All public interfaces tested
- Error paths covered
- Performance benchmarks for critical paths

## Critical Paths

### 1. Model Generation Path
```
Request → Framework.Generate()
       → ModelRegistry.Get()
       → Model.ValidateRequest()
       → Model.Generate()
       → Metrics.RecordGeneration()
       → Response
```

### 2. Tool Execution Path
```
Request → Framework.ExecuteTool()
       → ToolRegistry.Get()
       → Tool.Validate()
       → Tool.Execute()
       → Tool.ValidateOutput()
       → Response
```

### 3. Flow Execution Path
```
Request → Framework.ExecuteFlow()
       → FlowRegistry.Get()
       → Flow.Validate()
       → Flow.Execute()
         → ExecuteTool(step1)
         → Generate(step2)
         → ExecuteTool(step3)
       → Flow.ValidateOutput()
       → Response
```

## Performance Considerations

- **Connection Pooling**: HTTP клиенты с пулами соединений
- **Object Pooling**: Снижение GC压力 через pooling
- **Caching**: Многоуровневое (L1: memory, L2: Redis)
- **Concurrency**: Thread-safe регистры с RWMutex

## Security Considerations

- **API Key Management**: Environment variables, шифрование при передаче
- **Input Validation**: JSON schema валидация, защита от injection
- **Rate Limiting**: Per-model/tool лимиты с exponential backoff
- **Audit Logging**: Все запросы с маскировкой чувствительных данных
