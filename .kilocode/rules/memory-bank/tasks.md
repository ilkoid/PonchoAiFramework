# Implementation Tasks

This file documents repetitive tasks and their workflows for future reference.

## Phase 1 Foundation Implementation (COMPLETED)

**Last performed:** December 11, 2025

**Files created/modified:**
- `go.mod` - Go module definition
- `interfaces/types.go` - Core type definitions
- `interfaces/interfaces.go` - Interface aliases  
- `interfaces/logger.go` - Logger interface
- `core/framework.go` - Main framework implementation
- `core/interfaces.go` - Core interface aliases
- `core/logger.go` - Logging system
- `core/config/config.go` - Configuration manager
- `core/config/loader.go` - Configuration loader
- `core/config/validator.go` - Configuration validator
- `core/base/model.go` - Base model implementation
- `core/base/tool.go` - Base tool implementation
- `core/base/flow.go` - Base flow implementation
- `core/registry/model_registry.go` - Model registry
- `core/registry/tool_registry.go` - Tool registry
- `core/registry/flow_registry.go` - Flow registry
- `config.yaml` - Comprehensive configuration file
- Multiple test files for all components

**Steps followed:**
1. **Set up Go module**: Created `go.mod` with Go 1.25.1 and yaml.v3 dependency
2. **Define core types**: Created comprehensive type system in `interfaces/types.go`
3. **Implement interfaces**: Defined all core interfaces with proper method signatures
4. **Build base classes**: Created extensible base implementations for models, tools, flows
5. **Implement registries**: Built thread-safe registries with full CRUD operations
6. **Create configuration system**: Implemented YAML/JSON config with env var support
7. **Build main framework**: Created `PonchoFrameworkImpl` with lifecycle management
8. **Add logging system**: Implemented structured logging with multiple formats
9. **Write comprehensive tests**: Created unit tests for all components with >90% coverage
10. **Create production config**: Built comprehensive `config.yaml` with all components

**Important considerations:**
- Use `sync.RWMutex` for thread-safe registries
- Implement proper error handling with context
- Use dependency injection pattern for configuration
- Follow Go idioms and naming conventions
- Implement comprehensive validation for all inputs
- Use structured logging with contextual fields
- Design for extensibility with interface-based architecture

**Testing patterns established:**
- Table-driven tests for multiple scenarios
- Mock implementations for external dependencies
- Benchmark tests for performance critical paths
- Integration tests for end-to-end workflows
- Coverage target: >90% for all components

## Phase 2 Model Integration (NEXT)

**Target implementation timeframe:** 2-3 weeks

**Files to modify/create:**
- `models/deepseek/client.go` - DeepSeek API client
- `models/deepseek/model.go` - DeepSeek model implementation
- `models/deepseek/streaming.go` - Streaming support
- `models/zai/client.go` - Z.AI API client
- `models/zai/model.go` - Z.AI model implementation
- `models/zai/vision.go` - Vision capabilities
- `models/zai/streaming.go` - Streaming support
- `models/common/request.go` - Shared request utilities
- `models/common/response.go` - Shared response utilities
- `models/common/converter.go` - Format converters

**Steps to follow:**
1. **Create HTTP client base**: Reusable client with connection pooling, retries, timeouts
2. **Implement DeepSeek adapter**: OpenAI-compatible API integration
3. **Implement Z.AI adapter**: Custom API with vision support
4. **Add streaming support**: Real-time response streaming for both providers
5. **Handle errors**: Proper error mapping and retry mechanisms
6. **Add validation**: Request/response validation for both providers
7. **Write integration tests**: Real API calls with mock servers
8. **Add benchmarks**: Performance testing for model adapters

**Important notes:**
- DeepSeek uses OpenAI-compatible API format
- Z.AI requires custom authentication and request format
- Vision support needs base64 image encoding
- Streaming requires proper chunk handling and error recovery
- Rate limiting per provider is essential
- Token counting varies by provider

## Phase 3 Tool Implementation (FUTURE)

**Files to modify/create:**
- `tools/article_importer/tool.go` - S3 article import tool
- `tools/wildberries/categories.go` - Wildberries categories tool
- `tools/wildberries/characteristics.go` - Wildberries characteristics tool
- `tools/s3/storage.go` - S3 storage operations tool
- `tools/vision/analyzer.go` - Vision analysis tool
- `tools/common/validation.go` - Shared validation utilities
- `tools/common/retry.go` - Retry mechanisms

**Implementation patterns:**
- Implement `PonchoTool` interface
- Use `PonchoBaseTool` for common functionality
- Define input/output JSON schemas
- Implement proper error handling
- Add configuration validation
- Use dependency injection for external clients

## Phase 4 Flow Implementation (FUTURE)

**Files to modify/create:**
- `flows/article_processor/flow.go` - Article processing workflow
- `flows/mini_agent/flow.go` - Mini-agent workflow
- `flows/common/orchestrator.go` - Flow orchestration utilities
- `flows/common/dependency.go` - Dependency resolution
- `flows/common/validation.go` - Flow validation utilities

**Implementation patterns:**
- Implement `PonchoFlow` interface
- Use `PonchoBaseFlow` for common functionality
- Support both sequential and parallel execution
- Implement proper dependency resolution
- Add flow-level error handling and recovery
- Support streaming flows where applicable

## Configuration Management Patterns

**Environment variable handling:**
- Use `${VAR_NAME}` syntax in YAML for substitution
- Support type conversion (string to int, bool, duration)
- Use prefixing to avoid conflicts (`PONCHO_`)
- Provide sensible defaults for all configuration

**Validation patterns:**
- Validate required fields exist
- Check type compatibility
- Validate ranges and constraints
- Provide clear error messages
- Use struct tags for metadata

## Testing Patterns

**Unit tests:**
- Table-driven tests for multiple scenarios
- Mock external dependencies
- Test both success and error paths
- Use testify for assertions and mocks
- Target >90% code coverage

**Integration tests:**
- Use testcontainers for external services
- Mock HTTP servers for API testing
- Test real configuration loading
- Test error scenarios and recovery

**Benchmark tests:**
- Test performance critical paths
- Measure memory allocation
- Test concurrent access
- Compare implementation alternatives

## Error Handling Patterns

**Framework errors:**
- Use custom error types with context
- Include error codes and descriptions
- Wrap errors with proper context
- Log errors with structured data
- Provide recovery mechanisms where possible

**API errors:**
- Map HTTP status codes to domain errors
- Handle rate limiting with backoff
- Retry transient failures
- Validate API responses
- Handle authentication failures

## Logging Patterns

**Structured logging:**
- Use key-value pairs for context
- Include request IDs for tracing
- Log at appropriate levels (debug, info, warn, error)
- Mask sensitive data (API keys, tokens)
- Use consistent field names

**Performance logging:**
- Log request durations
- Track token usage
- Monitor error rates
- Log resource utilization
- Include correlation IDs

## Memory Management Patterns

**Object pooling:**
- Pool frequently allocated objects
- Use sync.Pool for temporary objects
- Reset objects before returning to pool
- Monitor pool effectiveness

**Resource cleanup:**
- Use context for cancellation
- Close HTTP clients properly
- Release resources in defer statements
- Implement graceful shutdown
- Monitor memory usage patterns

## Concurrency Patterns

**Thread safety:**
- Use RWMutex for read-heavy operations
- Use channels for communication
- Avoid shared mutable state
- Use context for cancellation
- Implement proper shutdown

**Registry patterns:**
- Thread-safe CRUD operations
- Atomic operations where possible
- Copy data for read operations
- Validate inputs before operations
- Handle concurrent access gracefully

## Configuration Reloading (FUTURE)

**Hot reload patterns:**
- Use file system watchers
- Validate new configuration before applying
- Gracefully handle configuration errors
- Notify components of changes
- Maintain backward compatibility

**Atomic updates:**
- Update configuration atomically
- Rollback on validation failures
- Maintain service availability during reload
- Test reload scenarios thoroughly