# CLI Article Flow Package

This package provides a simple, type-safe flow for processing fashion articles through the PonchoAiFramework. It implements the Version 1 linear processing pipeline without polluting the core framework.

## Quick Usage

```go
import "github.com/ilkoid/PonchoAiFramework/cli/articleflow"

// Create flow with dependencies
flow := articleflow.NewArticleFlow(
    s3Tool,
    visionModel,
    textModel,
    wbClient,
    mediaPipeline,
    logger,
    config,
    wbCache,
)

// Run the flow
state, err := flow.Run(ctx, "article-123")
if err != nil {
    log.Fatal(err)
}

// Access results
fmt.Println(string(state.FinalWBPayload))
```

## Architecture

- `ArticleFlowState`: Type-safe state container for the flow
- `ArticleFlow`: Orchestrator that runs the 11-step pipeline
- `WBMemoryCache`: In-memory cache for Wildberries data
- `ProcessImagesConcurrently`: Utility for parallel image processing

## Components

### ArticleFlowState
Holds all data through the processing pipeline:
- Raw S3 data (JSON + images)
- Vision analysis results
- Creative descriptions
- Wildberries data (parents, subjects, characteristics)
- Final payload

### ArticleFlow
Implements the complete processing pipeline:
1. Load article from S3
2. Process/resize images
3. Run vision analysis (Prompt 1)
4. Generate creative descriptions (Prompt 2)
5. Fetch Wildberries data
6. Select WB subject (Prompt 3)
7. Fetch characteristics
8. Generate final payload (Prompt 4)

### WBMemoryCache
Caches Wildberries API responses with TTL:
- Parent categories (24h TTL)
- Subjects (12h TTL)
- Characteristics (6h TTL)
- Configurable cache size limit

### Configuration
All settings are configurable via YAML:
- Image resize parameters
- Concurrency limits
- Cache TTLs
- Model selection and parameters

## CLI Command

See `cmd/poncho-cli` for a complete command-line implementation.