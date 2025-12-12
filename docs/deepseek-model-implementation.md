# DeepSeek Model Implementation

## Overview

The DeepSeek model adapter provides integration with DeepSeek's OpenAI-compatible API for text generation and reasoning capabilities. This implementation supports streaming responses, tool calling, and comprehensive error handling.

## Features

### Core Capabilities
- ✅ **Text Generation**: High-quality text generation with configurable parameters
- ✅ **Streaming**: Real-time response streaming with proper chunk handling
- ✅ **Tool Calling**: Support for function calling and tool execution
- ✅ **System Role**: Support for system messages and role-based conversations
- ❌ **Vision**: Not supported (DeepSeek is text-only)
- ✅ **Retry Logic**: Automatic retry with exponential backoff
- ✅ **Error Handling**: Comprehensive error mapping and recovery

### Configuration Options
- `api_key`: DeepSeek API key (required)
- `model_name`: Model name (default: "deepseek-chat")
- `max_tokens`: Maximum tokens in response (default: 4000)
- `temperature`: Sampling temperature (0.0-2.0, default: 0.7)
- `base_url`: API base URL (default: "https://api.deepseek.com")
- `timeout`: Request timeout (default: 30s)

## Architecture

### Components

#### 1. DeepSeekModel (`models/deepseek/model.go`)
Main model implementation that implements the `PonchoModel` interface.

**Key Methods:**
- `Generate()`: Non-streaming text generation
- `GenerateStreaming()`: Streaming text generation
- `Initialize()`: Model initialization with configuration
- `Shutdown()`: Graceful shutdown and cleanup

**Features:**
- Request/response conversion between Poncho and DeepSeek formats
- Comprehensive validation and error handling
- Metrics collection and logging
- Tool calling support

#### 2. DeepSeekClient (`models/deepseek/client.go`)
HTTP client wrapper for DeepSeek API communication.

**Key Methods:**
- `PrepareHeaders()`: Sets up authentication and content-type headers
- `ValidateRequest()`: Validates requests against DeepSeek requirements
- `GetModelCapabilities()`: Returns model capability information
- `GetModelMetadata()`: Returns model metadata

**Features:**
- OpenAI-compatible API integration
- Request validation and preprocessing
- Error mapping and retry logic
- Health checks and rate limiting

#### 3. Streaming Handler (`models/deepseek/streaming.go`)
Server-Sent Events (SSE) stream processing.

**Key Functions:**
- `ProcessStreamChunk()`: Parses individual stream chunks
- `ConvertStreamChunkToPoncho()`: Converts to Poncho format
- `ProcessSSEStream()`: Handles SSE stream processing
- `IsStreamComplete()`: Checks stream completion status

**Features:**
- Robust SSE parsing with error recovery
- Tool call delta handling
- Context cancellation support
- Stream completion detection

#### 4. Type Definitions (`models/deepseek/types.go`)
DeepSeek-specific API types and constants.

**Key Types:**
- `DeepSeekRequest`: API request structure
- `DeepSeekResponse`: API response structure
- `DeepSeekStreamResponse`: Streaming response structure
- `DeepSeekTool`: Tool definition structure

**Features:**
- Complete API coverage with all fields
- JSON serialization/deserialization support
- Type-safe constants for finish reasons, tool choices

## Usage Examples

### Basic Text Generation

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/ilkoid/PonchoAiFramework/interfaces"
    "github.com/ilkoid/PonchoAiFramework/models/deepseek"
)

func main() {
    // Create model
    model := deepseek.NewDeepSeekModel()
    
    // Initialize with configuration
    config := map[string]interface{}{
        "api_key":    "your-deepseek-api-key",
        "model_name": "deepseek-chat",
        "max_tokens":  1000,
        "temperature": 0.7,
    }
    
    ctx := context.Background()
    err := model.Initialize(ctx, config)
    if err != nil {
        panic(err)
    }
    defer model.Shutdown(ctx)
    
    // Generate response
    req := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {Type: interfaces.PonchoContentTypeText, Text: "Hello, how are you?"},
                },
            },
        },
    }
    
    resp, err := model.Generate(ctx, req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %s\n", resp.Message.Content[0].Text)
}
```

### Streaming Generation

```go
func streamingExample() {
    model := deepseek.NewDeepSeekModel()
    
    config := map[string]interface{}{
        "api_key":    "your-deepseek-api-key",
        "model_name": "deepseek-chat",
    }
    
    ctx := context.Background()
    err := model.Initialize(ctx, config)
    if err != nil {
        panic(err)
    }
    defer model.Shutdown(ctx)
    
    // Streaming request
    req := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {Type: interfaces.PonchoContentTypeText, Text: "Tell me a story"},
                },
            },
        },
        Stream: true,
    }
    
    // Process streaming response
    callback := func(chunk *interfaces.PonchoStreamChunk) error {
        if chunk.Delta != nil && len(chunk.Delta.Content) > 0 {
            for _, part := range chunk.Delta.Content {
                if part.Type == interfaces.PonchoContentTypeText {
                    fmt.Print(part.Text)
                }
            }
        }
        
        if chunk.Done {
            fmt.Println("\n[Stream completed]")
        }
        return nil
    }
    
    err = model.GenerateStreaming(ctx, req, callback)
    if err != nil {
        panic(err)
    }
}
```

### Tool Calling

```go
func toolCallingExample() {
    model := deepseek.NewDeepSeekModel()
    
    config := map[string]interface{}{
        "api_key":    "your-deepseek-api-key",
        "model_name": "deepseek-chat",
    }
    
    ctx := context.Background()
    err := model.Initialize(ctx, config)
    if err != nil {
        panic(err)
    }
    defer model.Shutdown(ctx)
    
    // Request with tools
    req := &interfaces.PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*interfaces.PonchoMessage{
            {
                Role: interfaces.PonchoRoleUser,
                Content: []*interfaces.PonchoContentPart{
                    {Type: interfaces.PonchoContentTypeText, Text: "What's the weather in Boston?"},
                },
            },
        },
        Tools: []*interfaces.PonchoToolDef{
            {
                Name:        "get_weather",
                Description: "Get current weather for a location",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type":        "string",
                            "description": "The city and state, e.g. San Francisco, CA",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    }
    
    resp, err := model.Generate(ctx, req)
    if err != nil {
        panic(err)
    }
    
    // Check for tool calls in response
    for _, part := range resp.Message.Content {
        if part.Type == interfaces.PonchoContentTypeTool {
            fmt.Printf("Tool call: %s\n", part.Tool.Name)
            fmt.Printf("Arguments: %v\n", part.Tool.Args)
        } else if part.Type == interfaces.PonchoContentTypeText {
            fmt.Printf("Text: %s\n", part.Text)
        }
    }
}
```

## Error Handling

### Error Types
The implementation uses the common error system with specific error codes:

- `INVALID_CONFIG`: Configuration errors
- `INVALID_API_KEY`: Authentication failures
- `RATE_LIMIT_ERROR`: Rate limiting
- `NETWORK_ERROR`: Network connectivity issues
- `TIMEOUT_ERROR`: Request timeouts
- `SERVER_ERROR`: DeepSeek API errors
- `STREAM_ERROR`: Streaming failures

### Retry Logic
Automatic retry is implemented with:
- Exponential backoff (1s, 2s, 4s, 8s, 16s, 30s max)
- Jitter to prevent thundering herd
- Retryable error detection
- Context cancellation support

## Performance Considerations

### Optimization Features
- **Connection Pooling**: HTTP client with connection reuse
- **Request Validation**: Early validation to avoid unnecessary API calls
- **Streaming**: Reduced latency for long responses
- **Metrics Collection**: Performance monitoring and debugging

### Benchmarks
Typical performance characteristics:
- **First Token Latency**: 200-500ms
- **Tokens/Second**: 50-100 tokens
- **Request Overhead**: 10-20ms
- **Memory Usage**: ~50MB base + request/response data

## Testing

### Test Coverage
Comprehensive test suite includes:
- Unit tests for all components
- Integration tests with mock API
- Streaming functionality tests
- Error handling validation
- Performance benchmarks

### Running Tests
```bash
# Run all tests
go test -v ./models/deepseek/...

# Run specific test
go test -v ./models/deepseek/... -run TestBasicFunctionality

# Run benchmarks
go test -v ./models/deepseek/... -bench=.
```

## Configuration

### Environment Variables
```bash
export DEEPSEEK_API_KEY="your-api-key"
export DEEPSEEK_BASE_URL="https://api.deepseek.com"  # Optional
```

### YAML Configuration
```yaml
models:
  deepseek-chat:
    provider: "deepseek"
    model_name: "deepseek-chat"
    api_key: "${DEEPSEEK_API_KEY}"
    max_tokens: 4000
    temperature: 0.7
    timeout: 30s
    supports:
      vision: false
      tools: true
      stream: true
      system: true
```

## Integration with PonchoFramework

### Registration
```go
import (
    "github.com/ilkoid/PonchoAiFramework/core"
    "github.com/ilkoid/PonchoAiFramework/models/deepseek"
)

func main() {
    // Create framework
    framework := core.NewPonchoFramework(config)
    
    // Register DeepSeek model
    deepseekModel := deepseek.NewDeepSeekModel()
    err := framework.RegisterModel("deepseek-chat", deepseekModel)
    if err != nil {
        panic(err)
    }
    
    // Start framework
    ctx := context.Background()
    err = framework.Start(ctx)
    if err != nil {
        panic(err)
    }
}
```

### Usage Through Framework
```go
// Generate using framework
resp, err := framework.Generate(ctx, &interfaces.PonchoModelRequest{
    Model: "deepseek-chat",
    Messages: []*interfaces.PonchoMessage{
        {
            Role: interfaces.PonchoRoleUser,
            Content: []*interfaces.PonchoContentPart{
                {Type: interfaces.PonchoContentTypeText, Text: "Hello!"},
            },
        },
    },
})
```

## Monitoring and Debugging

### Logging
The implementation provides structured logging with:
- Request/response logging
- Error details with context
- Performance metrics
- Debug information for troubleshooting

### Metrics
Automatic collection of:
- Request duration and success rates
- Token usage statistics
- Error rates by type
- Cache hit/miss ratios

## Limitations and Considerations

### Current Limitations
- No vision/image support
- Rate limits apply (configurable)
- Token limits vary by model
- No built-in caching (framework-level only)

### Best Practices
1. **API Key Security**: Use environment variables, never commit keys
2. **Error Handling**: Always check and handle errors appropriately
3. **Context Usage**: Properly handle context cancellation
4. **Resource Cleanup**: Always call Shutdown() when done
5. **Streaming**: Use streaming for long responses to improve UX

## Troubleshooting

### Common Issues
1. **Authentication Errors**: Check API key validity and permissions
2. **Rate Limiting**: Implement backoff and rate limiting
3. **Timeouts**: Increase timeout for long requests
4. **Streaming Issues**: Check network stability and context handling

### Debug Mode
Enable debug logging:
```go
config := map[string]interface{}{
    "api_key":    "your-api-key",
    "debug":       true,  // Enable debug logging
    "log_level":   "debug",
}
```

## Future Enhancements

### Planned Features
- [ ] Advanced caching strategies
- [ ] Request/response compression
- [ ] Additional model variants
- [ ] Enhanced metrics and monitoring
- [ ] Load balancing across multiple endpoints

### Contribution Guidelines
When contributing to the DeepSeek implementation:
1. Maintain test coverage >90%
2. Follow Go best practices and idioms
3. Update documentation for API changes
4. Add comprehensive error handling
5. Include performance benchmarks for new features