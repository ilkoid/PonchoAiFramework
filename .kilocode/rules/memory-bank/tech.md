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

**Configuration:**
```yaml
s3:
  url: "https://storage.yandexcloud.net"
  region: "ru-central1"
  bucket: "plm-ai"
  endpoint: "storage.yandexcloud.net"
  use_ssl: true
```

**Client:** MinIO Go SDK (S3-compatible)
**Authentication:** 
- Access Key: `S3_ACCESS_KEY` environment variable
- Secret Key: `S3_SECRET_KEY` environment variable

**Usage:**
- Article data storage (JSON)
- Fashion sketch images
- Product photos
- Technical drawings

### 2. Wildberries API

**Configuration:**
```yaml
wildberries:
  base_url: "https://content-api.wildberries.ru"
  timeout: 30
```

**Authentication:** API key via `WB_API_CONTENT_KEY` environment variable

**Endpoints Used:**
- `/content/v2/object/parent/all` - Parent categories
- `/content/v2/object/all` - Subjects/categories
- `/content/v2/object/charc/all` - Characteristics

### 3. Redis (Future Phase)

**Purpose:** Distributed caching layer (L2 cache)
**Authentication:** Via `REDIS_URL` environment variable
**Usage:**
- Cache model responses
- Reduce API call costs
- Improve response times

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

### Planned Go Modules

**Core:**
- Standard library (primary)
- `gopkg.in/yaml.v3` - YAML parsing
- JSON from stdlib

**HTTP & Networking:**
- `net/http` from stdlib (primary)
- Connection pooling built-in

**S3 Client:**
- MinIO Go SDK or AWS SDK v2

**Testing:**
- `github.com/stretchr/testify` - Testing utilities

**Monitoring:**
- `github.com/prometheus/client_golang` - Prometheus metrics

**Caching:**
- `github.com/go-redis/redis/v8` - Redis client (L2 cache)

### Dependency Philosophy

**Principles:**
1. **Minimize external dependencies**
2. **Prefer standard library**
3. **Thoroughly vet third-party libraries**
4. **Keep dependency tree shallow**
5. **Use stable, well-maintained packages**

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
- Local development environment
- Docker Compose (optional)
- Direct Go execution

### Staging
- Kubernetes cluster
- Separate namespace
- Test data and APIs

### Production
- Kubernetes cluster (multi-region)
- High availability (3+ replicas)
- Auto-scaling enabled
- Full monitoring stack

## Performance Targets

**Response Time:**
- p50: < 1 second
- p95: < 2 seconds
- p99: < 5 seconds

**Throughput:**
- > 100 requests/second per instance
- Horizontal scaling for higher loads

**Resource Usage:**
- Memory: < 512MB per instance
- CPU: < 70% average utilization

**Reliability:**
- Uptime: > 99.9%
- Error rate: < 1%
- MTTR: < 5 minutes
