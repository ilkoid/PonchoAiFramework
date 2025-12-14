# Technology Stack

## Programming Language

**Go 1.25.1**
- Modern Go with latest features
- Module path: `github.com/ilkoid/PonchoAiFramework`
- Standard library used extensively
- Minimal external dependencies preferred

## AI Models & APIs

### Primary Models

**1. DeepSeek**
- **Purpose**: Text generation and reasoning
- **API**: OpenAI-compatible REST API
- **Authentication**: API key via `DEEPSEEK_API_KEY` environment variable
- **Capabilities**: 
  - Tool calling support
  - Streaming responses
  - System role support
- **Max Tokens**: 4000 (configurable)
- **Temperature**: 0.7 (default)

**2. Z.AI (GLM-4.6V / GLM-4.6V-FLASH)**
- **Purpose**: Vision analysis and multimodal tasks
- **API**: Custom REST API
- **Authentication**: API key via `ZAI_API_KEY` environment variable
- **Capabilities**:
  - Vision analysis (fashion-specific)
  - Tool calling support
  - Streaming responses
  - System role support
- **Max Tokens**: 2000 (configurable)
- **Temperature**: 0.5 (default)
- **Specialization**: Fashion industry image analysis

## External Services

### 1. S3 Storage (Yandex Cloud)
- **Client**: Custom HTTP-based S3-compatible client
- **Authentication**: S3_ACCESS_KEY, S3_SECRET_KEY
- **Usage**: Article data (JSON), fashion sketch images, product photos
- **Features**: Image processing, resizing, optimization

### 2. Wildberries API
- **Authentication**: WB_API_CONTENT_KEY
- **Endpoints**: Parent categories, subjects, characteristics
- **Usage**: Категории, предметы, характеристики товаров

### 3. Redis (Future Phase)
- **Purpose**: Distributed caching layer (L2 cache)
- **Authentication**: REDIS_URL
- **Usage**: Cache model responses, reduce API costs

## Configuration Management

### YAML Configuration

**Primary Config:** `config.yaml`

```yaml
# Image optimization
image_optimization:
  enabled: true
  max_width: 640
  max_height: 480
  quality: 90
  max_size_bytes: 90000

# Logging
logger:
  level: "info"    # debug, info, warn, error
  format: "text"   # text, json
  file: ""         # empty = stdout
```

### Environment Variables

**Required:**
- `DEEPSEEK_API_KEY` - DeepSeek API authentication
- `ZAI_API_KEY` - Z.AI GLM API authentication
- `S3_ACCESS_KEY` - S3 storage access key
- `S3_SECRET_KEY` - S3 storage secret key
- `WB_API_CONTENT_KEY` - Wildberries API key

**Optional:**
- `REDIS_URL` - Redis connection string (future)
- `LOG_LEVEL` - Override logging level
- `PORT` - HTTP server port (if applicable)

## Development Tools

### Testing

**Framework:** Go standard `testing` package
**Additional Libraries (planned):**
- `github.com/stretchr/testify` - Assertions and mocks
- `github.com/stretchr/testify/mock` - Mock objects
- `github.com/stretchr/testify/assert` - Assertions

**Test Structure:**
```
tests/
├── unit/         # Unit tests (90%+ coverage target)
├── integration/  # Integration tests with real APIs
├── benchmark/    # Performance benchmarks
└── e2e/          # End-to-end tests
```

### Code Quality

**Linting:** `golangci-lint` (planned)
**Formatting:** `gofmt` / `goimports`
**Static Analysis:** `go vet`

**Quality Targets:**
- Test Coverage: > 90%
- Cyclomatic Complexity: < 10
- Documentation: All public APIs documented with godoc

### Build & Deployment

**Build Tool:** Native Go toolchain
```bash
go build -o poncho-framework ./cmd/poncho-framework
```

**Containerization:** Docker (planned)
```dockerfile
FROM golang:1.21-alpine AS builder
FROM alpine:latest
```

**Orchestration:** Kubernetes (production deployment)

## HTTP Client

**Configuration:**
```go
MaxIdleConns:        100
MaxIdleConnsPerHost: 10
IdleConnTimeout:     90s
TLSHandshakeTimeout: 10s
```

**Features:**
- Connection pooling
- Automatic retries with exponential backoff
- Timeout management
- HTTP/2 support

## Logging

**Levels:** debug, info, warn, error
**Formats:** text (development), json (production)
**Output:** stdout (default), file (configurable)

**Features:**
- Structured logging
- Context-aware logging
- Sensitive data masking
- Audit trail for all AI operations

## Monitoring & Metrics (Future Phase)

### Prometheus

**Metrics Collected:**
- Request rate (by model, tool, flow)
- Response time (p50, p95, p99)
- Error rate (by type)
- Cache hit rate
- Resource utilization

### Grafana

**Dashboards:**
- PonchoFramework Overview
- Model Performance
- Tool Execution Metrics
- Flow Orchestration
- System Health

## Performance Optimization

### Memory Management

**Techniques:**
- Object pooling for frequently allocated structures
- Careful buffer management
- Graceful memory release

**Targets:**
- < 512MB per instance
- Minimal GC pressure
- No memory leaks

### Concurrency

**Patterns:**
- Goroutines for parallel execution
- Channels for communication
- Context for cancellation
- RWMutex for registries

**Features:**
- Thread-safe registries
- Concurrent request handling
- Rate limiting per provider
- Connection pooling

### Caching

**L1 Cache:** In-memory (fast, small)
**L2 Cache:** Redis (persistent, distributed)

**Strategy:**
- Cache key based on request hash
- TTL per component
- Cache invalidation on configuration change
- Progressive cache warming

## Security

### Encryption

**In Transit:**
- TLS 1.3 for all external connections
- Certificate validation
- Secure cipher suites

**At Rest:**
- API keys in environment variables
- No secrets in configuration files
- Encrypted storage for cached data

### Authentication & Authorization

**API Keys:**
- Environment variable storage
- Never logged or exposed
- Rotation support

**Rate Limiting:**
- Per-model limits
- Per-tool limits
- Exponential backoff
- Circuit breaker pattern

### Data Protection

**Features:**
- Input validation
- Output sanitization
- Sensitive data masking in logs
- GDPR compliance

## Reference Implementation

### Legacy Code (Sample)

**Location:** `/sample/poncho-tools/`

**Purpose:** Reference implementation using Firebase GenKit

**Key Components:**
- `s3/client.go` - S3 integration patterns
- `wildberries/` - Wildberries API tools
- `mini-agent/` - Mini-agent flow architecture
- `cmd/simple_glm5v_processor/` - GLM vision examples

**Usage:** Study patterns, don't copy directly

## Dependencies (Future)

### Current Go Dependencies

**Core:**
- Standard library (primary)
- `gopkg.in/yaml.v3` - YAML parsing ✅
- JSON from stdlib ✅

**HTTP & Networking:**
- `net/http` from stdlib (primary) ✅
- Connection pooling built-in ✅

**S3 Client:**
- Custom S3-compatible implementation using `net/http` ✅
- Image processing using `image/jpeg`, `image/png` ✅
- Base64 encoding using `encoding/base64` ✅

**Testing:**
- `github.com/stretchr/testify` - Testing utilities (planned)

**Monitoring:**
- `github.com/prometheus/client_golang` - Prometheus metrics (planned for Phase 6)

**Caching:**
- `github.com/go-redis/redis/v8` - Redis client (L2 cache) (planned for Phase 6)

**CLI & Flow System:**
- Custom CLI implementations using standard library ✅
- Flow orchestration with context management ✅
- Media processing and image handling ✅
- Wildberries API integration ✅

### Dependency Philosophy

**Principles:**
1. **Minimize external dependencies**
2. **Prefer standard library**
3. **Thoroughly vet third-party libraries**
4. **Keep dependency tree shallow**
5. **Use stable, well-maintained packages**

### Prompt System Dependencies

**Core Components:**
- **Template Parser**: Custom V1 format parser using `regexp` ✅
- **Variable Processor**: Template variable substitution and validation ✅
- **Template Validator**: Comprehensive validation with fashion-specific rules ✅
- **Cache Implementation**: Thread-safe LRU cache using `container/list` ✅
- **Template Executor**: Model request building and execution ✅

**External Dependencies:**
- None for core prompt system (uses only Go standard library)
- Framework integration through interfaces ✅

### Tool System Dependencies

**Core Components:**
- **S3 Client**: Custom HTTP-based S3-compatible client ✅
- **Image Processing**: Go standard library `image` package ✅
- **Tool Factory**: Dynamic tool creation and validation ✅
- **Configuration System**: YAML-based tool configuration ✅

**External Dependencies:**
- None for core tool system (uses only Go standard library)
- S3 compatibility through HTTP API ✅
- Image format support (JPEG, PNG) ✅

## Development Environment

### Setup Requirements

**Go Version:** 1.25.1 or higher
**OS:** Linux (primary), macOS, Windows (supported)
**IDE:** VS Code with Go extension (recommended)

**Environment Files:**
- `.env` - Local development secrets (not committed)
- `.env.example` - Template for environment variables
- `config.yaml` - Application configuration

### Local Development

**Run Commands:**
```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build
go build -o poncho-framework ./cmd/poncho-framework

# Run
./poncho-framework
```

## Deployment Architecture

### Development
- Local environment с Docker Compose
- Direct Go execution

### Production
- Kubernetes cluster (multi-region)
- High availability (3+ replicas)
- Auto-scaling с monitoring stack

## Performance Targets

- **Response Time**: p50 < 1s, p95 < 2s, p99 < 5s
- **Throughput**: > 100 requests/second per instance
- **Resource Usage**: Memory < 512MB, CPU < 70%
- **Reliability**: Uptime > 99.9%, Error rate < 1%