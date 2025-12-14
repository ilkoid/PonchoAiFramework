# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PonchoAiFramework is a custom AI framework written in Go designed to replace Firebase GenKit in the Poncho Tools project. It provides unified API access to AI models (DeepSeek, Z.AI GLM), tools, and workflows with a specialization in fashion industry tasks and Wildberries marketplace integration.

## Build & Test Commands

```bash
# Format all code
go fmt ./...

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./core/...
go test -v ./models/deepseek/...
go test -v ./models/zai/...
go test -v ./tools/...

# Build main framework
go build -o poncho-framework ./cmd/poncho-framework

# Build utilities
go build -o prompt-tester ./cmd/prompt-tester
go build -o integration-tester ./cmd/integration-tester
go build -o simple-test ./cmd/simple-test
```

## Architecture Overview

The PonchoAiFramework follows Clean Architecture principles with clear separation between layers and dependency inversion. The core layer contains only business logic and interfaces, while all implementation details are in outer layers.

### Directory Structure

```
/interfaces/              # Core interfaces (no implementation dependencies)
├── types.go             # Data structures and model/tool/flow configs
├── interfaces.go        # Component interfaces (PonchoModel, PonchoTool, etc.)
└── factory.go           # Factory interfaces for dependency injection

/core/                   # Business logic (no imports from models/tools)
├── base/                # Base classes with shared functionality
├── config/              # Configuration management
│   ├── loader.go        # Configuration loading
│   ├── validator.go     # Configuration validation
│   └── models.go        # Configuration types
├── registry/            # Component registries
│   ├── model_registry.go
│   ├── tool_registry.go
│   ├── flow_registry.go
│   └── factory_registry.go  # Factory management
├── service_locator.go   # Service locator for dependency injection
└── framework.go         # Main framework orchestrator

/factories/              # Factory implementations (outer layer)
├── models/              # Model factories
│   └── model_factory.go # DeepSeek, Z.AI, OpenAI factory implementations
└── tools/               # Tool factories
    └── s3_tool_factory.go # S3-related tool factories

/models/                 # Model implementations (outer layer)
├── deepseek/            # DeepSeek API integration
├── zai/                 # Z.AI GLM integration
└── common/              # Shared model utilities

/tools/                  # Tool implementations (outer layer)
├── article_importer/    # S3-based article import tool
├── s3/                  # S3-compatible storage client
└── wildberries/         # Wildberries marketplace integration

/prompts/                # Prompt management system
├── manager.go           # Prompt manager
├── parser.go            # V1 format parser
└── validator.go         # Prompt validation
```

### Core Components

1. **Framework Core** (`/core`):
   - Central orchestrator with thread-safe registries
   - **NO direct imports** from model or tool implementations
   - Uses interfaces and dependency injection pattern

2. **Models** (`/models`): AI model adapters
   - `deepseek/`: Text generation and reasoning
   - `zai/`: GLM-4.6/4.6V for vision and multimodal tasks
   - `common/`: Shared utilities (auth, retry, metrics, tokenizer)

3. **Tools** (`/tools`): Implementations for external integrations
   - `article_importer/`: Product data import with S3 integration
   - `s3/`: S3-compatible storage client
   - `wildberries/`: Wildberries API integration

4. **Factories** (`/factories`):
   - Model and tool creation with configuration
   - Maintains clean architecture by separating creation logic

5. **Prompt System** (`/prompts`): V1 format parser with Handlebars templating

### Key Architectural Patterns

- **Clean Architecture**: Core has no dependency on implementation details
- **Service Locator Pattern**: Centralized factory and service management
- **Factory Pattern**: Dynamic creation of model and tool instances
- **Registry Pattern**: Centralized registration and retrieval
- **Interface Segregation**: Clean separation between interfaces and implementations
- **Dependency Inversion**: High-level modules don't depend on low-level modules

## Configuration

### Primary Configuration
- `config.yaml`: Main configuration file (see `configs/config.yaml.example`)
- Models: Configure API keys, timeouts, and capabilities
- Tools: Set up S3 credentials, Wildberries API keys
- Flows: Define workflow orchestration parameters

### Required Environment Variables
```bash
DEEPSEEK_API_KEY=          # DeepSeek API access
ZAI_API_KEY=               # Z.AI GLM API access
S3_ACCESS_KEY=             # S3-compatible storage access
S3_SECRET_KEY=             # S3-compatible storage secret
WB_API_CONTENT_KEY=        # Wildberries content API key
```

## Clean Architecture Rules

**CRITICAL**: The core package must NEVER import from `models/` or `tools/` packages. This maintains clean architecture.

### Dependency Flow
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Commands   │───▶│    Core     │───▶│ Interfaces  │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
                          │                   ▲
                          │                   │
                          ▼                   │
                   ┌─────────────┐           │
                   │ Service     │───────────┘
                   │ Locator     │
                   └─────────────┘
                          │
                          ▼
                   ┌─────────────┐
                   │  Factories  │
                   │  (Models &  │
                   │   Tools)    │
                   └─────────────┘
```

### Factory Pattern Implementation

The framework uses factories to create model and tool instances:

1. **Factory Interfaces** (`/interfaces/factory.go`):
   - `ModelFactory`: Creates model instances from config
   - `ToolFactory`: Creates tool instances from config
   - `ModelFactoryManager`: Manages model factories
   - `ToolFactoryManager`: Manages tool factories

2. **Factory Implementations** (`/factories/`):
   - `factories/models/model_factory.go`: DeepSeek, Z.AI, OpenAI factories
   - `factories/tools/s3_tool_factory.go`: S3-related tool factories

3. **Service Locator** (`/core/service_locator.go`):
   - Manages factory instances
   - Provides factory registration
   - Enables dependency injection without violating architecture

### Tool Configuration Note

Since `ToolConfig` doesn't have a `Type` field, tool types are determined by:
- Passing the type explicitly to factory methods
- Using `config.CustomParams["type"]` field
- Tool-specific factory conventions

## Development Guidelines

### Code Organization
- All public interfaces are defined in `/interfaces/`
- Core contains only business logic and interfaces
- Implementations are in outer layers (`models/`, `tools/`, `factories/`)
- Model adapters follow consistent client pattern with streaming support
- Tools implement the PonchoTool interface with JSON schema validation
- Use the service locator for factory access, not direct imports

### Testing Strategy
- Unit tests in each package (`*_test.go`)
- Integration tests for real API calls (`*_integration_test.go`, `*_real_test.go`)
- Test utilities use testify for assertions and mocking
- Target coverage: >90% for core components

### Error Handling
- Use defined error types from `/models/common/errors.go`
- Implement retry logic with exponential backoff
- Log errors with context using structured logging
- Return wrapped errors with original context preserved

## Special Features

### Multimodal Support
- Vision analysis through GLM-4.6V model
- Image processing for fashion items
- Structured data extraction from product images

### Fashion Domain Specifics
- Integration with Wildberries API for product categories
- Specialized prompts for fashion item analysis
- Support for product specifications and characteristics

### Production Deployment
- Dockerfile.production with multi-stage builds
- Comprehensive monitoring with Prometheus/Grafana
- Health checks and graceful shutdown
- Rate limiting and caching layers

## Framework Rules (Critical)

### Basic Principle
**Single direction of dependencies:**
Dependencies always flow from outside to inside, from concrete to abstract:

`cmd / ui / cli` → `flows` → `core (framework, interfaces, prompts)` → `models / tools` → `infra (HTTP, S3, WB)`

Never in the opposite direction.

### Layer Rules

#### 1. Core / Interfaces
- Packages `core/*`, `interfaces/*`, `prompts/*`:
  - **NEVER** import:
    - `tools/*`
    - `models/*`
    - `cli/*`
    - `ui/*`
    - `cmd/*`
  - Contain only:
    - Interfaces (`PonchoModel`, `PonchoTool`, `PonchoFlow`, logger, etc.)
    - Base types and errors
    - Framework orchestrator and common configuration

#### 2. Models / Tools
- Packages `models/*`, `tools/*`:
  - May only import:
    - `interfaces/*`
    - `core/*` (for base types)
    - Standard library
  - Cannot import:
    - `cli/*`
    - `ui/*`
    - `flows/*`
    - `cmd/*`

#### 3. Flows / Application Logic
- Packages `flows/*`, `cli/articleflow/*`:
  - May import:
    - `interfaces/*`, `core/*`
    - `models/*`, `tools/*` (as dependencies/adapters)
    - Helper packages (concurrency utils, caches)
  - Cannot import:
    - `ui/*`
    - `cmd/*`

#### 4. CLI / UI / cmd
- Packages `cli/*`, `ui/*`, `cmd/*`:
  - May import any internal framework packages
  - No functionality from them should be needed "inside":
    - No packages should import `ui/*` or `cmd/*`

### Interface and Dependency Rules

1. **Interfaces are declared where they are used.**
   If `ArticleFlow` uses `WildberriesClient`, the interface is declared near `ArticleFlow`, not in `tools/wildberries`.

2. **Dependencies are injected, not created internally.**
   Inside `ArticleFlow`, `SimpleConsoleUI`, any flows and services:
   - Forbidden to call `NewXxxClient()` directly from env/config
   - All dependencies come through constructor or factory (DI at Go code level)

3. **No circular dependencies between packages.**
   Any attempt to bypass this through type copying or `interface{}` is considered an architectural defect.

### State and Context Rules

1. **Global state is forbidden.**
   - No global singletons with live state (clients, caches, flow state)
   - Only allowed:
     - Configuration loaded at startup and passed by dependencies
     - Registry objects within the framework, protected by mutexes

2. **In-memory state only per-request / per-flow.**
   - `FlowContext`, `ArticleFlowState` and similar structures live only within a single scenario execution
   - Forbidden to store large binary data "permanently":
     - Images and large payloads should be represented by links (paths, URLs, S3 keys), not `[]byte`

3. **Go context (`context.Context`) is mandatory in public APIs.**
   - All public methods:
     - `Generate`, `ExecuteTool`, `ExecuteFlow`, `Run`, network calls, etc. - accept `context.Context`
   - Any long process (HTTP, LLM, S3, WB) must respect context cancellation

### Error and Logging Rules

1. **Errors are not "swallowed".**
   - Forbidden: ignoring errors (`_ = err`) outside obvious, locally justified places
   - Required: `error wrapping`
     - `fmt.Errorf("s3 load article %s: %w", articleID, err)`

2. **Logging only through common logger interface.**
   - No direct calls to `log.Printf` in business code
   - Use `interfaces.Logger` / `core.Logger`, passed by dependencies

3. **UI output is separated from logs.**
   - Console UI writes to `stdout`/`stderr`, but business code doesn't know UI format
   - Logs - structured, through logger; UI strings - through separate layer

### Test and Extensibility Rules

1. **New functionality - with extension point, not fork.**
   If you need to add a model/tool/flow:
   - Implement existing interface
   - Register through factory/registry
   - Don't break existing contracts without extreme necessity

2. **Testability:**
   Any new "architectural unit" (flow, UI component, cache, adapter):
   - Must be testable with mock interfaces
   - Should not pull live HTTP clients without substitution option

3. **No "god packages" (`common`, `utils`).**
   Cannot put everything in a shared `utils`/`common` without clear responsibility.
   If you want to do this - first clarify the responsibility and layer where it belongs.

### Practical Rule for Any New Code

Before adding a new file/package, the developer must answer "yes" to all questions:

1. Is it unambiguously understood which layer the new code belongs to (core / model / tool / flow / cli / ui / infra)?
2. Does this package import dependencies only "down" by layers, not "up"?
3. Can a concrete implementation (model, tool, client) be replaced without changes in "higher" layers?
4. Can the component be tested by substituting dependencies through interfaces?

If at least one answer is "no" - the architecture is violated, such PR is not accepted until reworked.