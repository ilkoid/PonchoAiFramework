# Wildberries API Integration

This package provides comprehensive integration with the Wildberries marketplace API, specifically designed for fashion industry applications in the PonchoAiFramework.

## Features

- **Full API Coverage**: Complete implementation of Wildberries Content API v2
- **Rate Limiting**: Built-in token bucket rate limiting (100 requests/minute)
- **Fashion AI Analysis**: Integration with GLM-4.6v for fashion item analysis
- **Category Matching**: Intelligent category and subject matching algorithms
- **Caching**: Intelligent caching for categories, subjects, and characteristics
- **Error Handling**: Comprehensive error handling with retry logic
- **Production Ready**: Thread-safe with proper logging and monitoring

## API Endpoints Supported

### Core Categories API
- `GET /content/v2/object/parent/all` - Get all parent categories
- `GET /content/v2/object/all` - Get subjects with filtering
- `GET /content/v2/object/charcs/{subjectId}` - Get subject characteristics

### Directory APIs
- `GET /api/content/v1/brands` - Get brands by subject
- `GET /content/v2/directory/colors` - Get color values
- `GET /content/v2/directory/kinds` - Get gender values
- `GET /content/v2/directory/seasons` - Get season values
- `GET /content/v2/directory/vat` - Get VAT rates
- `GET /content/v2/directory/tnved` - Get HS-codes
- `GET /content/v2/directory/countries` - Get countries

### System
- `GET /ping` - API health check

## Quick Start

### 1. Configuration

```yaml
tools:
  wildberries:
    api_key: "${WB_CONTENT_API_KEY}"
    base_url: "https://content-api.wildberries.ru"
```

### 2. Initialize Tool

```go
import (
    "github.com/ilkoid/PonchoAiFramework/tools/wildberries"
    "github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Create tool
tool := wildberries.NewWBTool(modelRegistry, logger)

// Initialize with config
err := tool.Initialize(ctx, map[string]interface{}{
    "api_key": os.Getenv("WB_CONTENT_API_KEY"),
})
```

### 3. API Usage Examples

#### Ping API
```go
result, err := tool.Execute(ctx, map[string]interface{}{
    "action": "ping",
})
```

#### Get Categories
```go
result, err := tool.Execute(ctx, map[string]interface{}{
    "action": "get_categories",
})
```

#### Analyze Fashion Item
```go
result, err := tool.Execute(ctx, map[string]interface{}{
    "action": "analyze_fashion",
    "image_url": "https://example.com/dress.jpg",
    "description": "Summer dress",
})
```

#### Find Suitable Category
```go
result, err := tool.Execute(ctx, map[string]interface{}{
    "action": "find_category",
    "keywords": []string{"платье", "женское", "лето"},
})
```

## Fashion AI Analysis

The package includes AI-powered fashion analysis that combines:

1. **Vision Analysis**: Uses GLM-4.6v model to analyze product images
2. **Category Matching**: Matches items to Wildberries categories
3. **Characteristic Extraction**: Identifies required product characteristics
4. **Schema Generation**: Creates product schemas compliant with Wildberries requirements

### Example Analysis Result

```json
{
  "type": "платье",
  "style": "летний",
  "season": "лето",
  "gender": "женский",
  "material": "хлопок",
  "color": "синий",
  "name": "Летнее хлопковое платье",
  "description": "Удобное летнее платье из натурального хлопка",
  "tags": ["платье", "лето", "хлопок"],
  "confidence": 0.95,
  "subjectId": 105,
  "subject": "Платья",
  "processedAt": "2024-08-16T11:30:00Z"
}
```

## Rate Limiting

The implementation uses a token bucket algorithm to respect Wildberries API limits:

- **Rate Limit**: 100 requests per minute
- **Burst Capacity**: 5 concurrent requests
- **Refill Interval**: 600ms between requests
- **Automatic Retry**: Built-in retry with exponential backoff for 429 errors

## Error Handling

The package provides structured error handling:

```go
type WBError struct {
    Code    int
    Message string
    Details string
}
```

Common HTTP status codes:
- `200` - Success
- `401` - Unauthorized (invalid/expired token)
- `429` - Rate limit exceeded
- `5XX` - Server errors

## Authentication

Wildberries API uses Bearer token authentication:

1. Generate token in seller account settings
2. Token validity: 180 days
3. Set environment variable: `WB_CONTENT_API_KEY`

## Testing

Run tests with:

```bash
go test ./tools/wildberries/... -v
```

Tests include:
- Unit tests for all components
- Rate limiting behavior
- Fashion analysis algorithms
- Category matching logic
- Error handling scenarios

## Development

### Adding New Features

1. Extend `WBClient` for new API endpoints
2. Add corresponding methods in `WBTool`
3. Update schema definitions
4. Add comprehensive tests
5. Update documentation

### Best Practices

1. Always check rate limit headers
2. Implement proper caching for static data
3. Use context for timeout handling
4. Log all API interactions for debugging
5. Handle all error scenarios gracefully

## Production Deployment

For production use:

1. Use production API URL
2. Set proper monitoring and alerting
3. Configure appropriate timeouts
4. Implement circuit breakers if needed
5. Monitor API quota usage

## Support

For issues and questions:
- Check API documentation: https://dev.wildberries.ru/
- Review example configurations
- Enable debug logging for troubleshooting