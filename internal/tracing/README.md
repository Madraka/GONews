# Tracing Package Documentation

## Overview

This package provides OpenTelemetry-based distributed tracing functionality for the News service. It allows for tracking requests as they flow through different services and components, making it easier to debug, monitor performance, and understand system behavior.

## Getting Started

### Initialize Tracing

Before using any tracing functionality, you must initialize the tracing system. This is typically done in your application's main function:

```go
package main

import (
    "news/internal/tracing"
    "log"
)

func main() {
    // Initialize tracing with service name
    cleanup, err := tracing.InitTracing("news-service")
    if err != nil {
        log.Fatalf("Failed to initialize tracing: %v", err)
    }
    defer cleanup() // Ensure proper shutdown
    
    // ... rest of your application code
}
```

### Configuration

Tracing uses the following environment variables:
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: Primary endpoint for the OpenTelemetry collector
- `OTEL_COLLECTOR_URL`: Fallback endpoint if the primary is not set
- If neither is set, defaults to "http://jaeger:4318"

## Core Features

### Creating Spans

There are several ways to create spans depending on your needs:

1. Basic spans:
```go
ctx, span := tracing.StartSpan(ctx, "operation-name")
defer span.End()
```

2. Spans with attributes:
```go
ctx, span := tracing.StartSpanWithAttributes(ctx, "operation-name",
    attribute.String("key", "value"),
    attribute.Int("count", 42))
defer span.End()
```

3. Error handling with spans:
```go
ctx, span := tracing.StartSpan(ctx, "operation-name")
defer tracing.EndSpan(span, err) // Automatically records error if err != nil
```

### Integration with HTTP/Gin

For HTTP handlers:

```go
// Standard http handler
http.Handle("/path", tracing.TracingHandler(myHandler))

// For Gin
router.GET("/path", tracing.WithTracing("operation-name", myHandler))
```

Or with middleware:

```go
router.Use(tracing.TracingMiddleware("my-service"))
```

## Best Practices

1. **Naming Conventions**:
   - Use descriptive span names that indicate the operation being performed
   - Follow a consistent format like `"ServiceName.OperationName"` or `"HTTP.Method.Path"`

2. **Adding Context**:
   - Always propagate the context with the span through function calls
   - Use attributes to add meaningful data about the operation

3. **Error Handling**:
   - Always mark spans with errors when they occur
   - Include relevant information to help with debugging

4. **Span Lifecycle**:
   - Create spans as close as possible to the operation they measure
   - End spans as soon as the operation completes

5. **Performance**:
   - Avoid creating too many spans (one per significant operation)
   - Be mindful of adding too many attributes that might cause overhead

## Examples

### Service Layer Example

```go
func (s *NewsService) GetArticleWithTracing(ctx context.Context, id string) (*models.Article, error) {
    ctx, span := tracing.StartSpanWithAttributes(ctx, "NewsService.GetArticle", 
        attribute.String("article.id", id))
    defer tracing.EndSpan(span, nil) // We'll update with error if needed
    
    // Database operations
    article, err := s.repository.GetByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    // Add additional context to the span
    span.SetAttributes(
        attribute.String("article.title", article.Title),
        attribute.String("article.author", article.Author),
    )
    
    return article, nil
}
```

### Wrapped Function Example

```go
func GetCachedData(ctx context.Context, key string) ([]byte, error) {
    var data []byte
    err := tracing.WrapFunction(ctx, "Cache.Get", func(ctx context.Context) error {
        var err error
        data, err = cache.Get(key)
        return err
    })
    return data, err
}
```

## Troubleshooting

- **Missing Traces**: Check that the OpenTelemetry collector is properly configured and accessible
- **Broken Trace Context**: Ensure that context is properly propagated between services
- **Too Many/Few Spans**: Adjust your instrumentation to create spans at the appropriate level of granularity