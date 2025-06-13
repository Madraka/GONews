# News Service Distributed Tracing Guide

## Overview

This guide provides comprehensive information on how distributed tracing is implemented in the News service. Tracing helps developers understand request flows, diagnose performance issues, and debug complex systems by providing end-to-end visibility.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   News API      │───▶│ OpenTelemetry   │───▶│    Jaeger       │
│  (Instrumented) │    │   Collector     │    │  (Trace Store)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

1. **Start the Environment**:
   ```bash
   ./scripts/start_otel_environment.sh
   ```

2. **Verify Tracing**:
   ```bash
   ./scripts/verify_tracing.sh
   ```

3. **Access Jaeger UI**: Open http://localhost:16686

4. **Generate Example Traces**:
   ```bash
   ./scripts/generate_example_traces.sh
   ```

## Tracing Components

1. **Core Libraries**:
   - `internal/tracing/tracing.go`: Main configuration and initialization
   - `internal/tracing/middleware.go`: Gin middleware for HTTP request tracing
   - `internal/tracing/utils.go`: Utility functions for DB, HTTP, Redis tracing
   - `internal/tracing/handlers.go`: HTTP handler wrappers and helpers

2. **Key Components**:
   - `tracing.InitTracing()`: Main setup function called in app startup
   - `tracing.StartSpan()`: Create basic spans
   - `tracing.StartSpanWithAttributes()`: Create spans with metadata
   - `tracing.EndSpan()`: End spans with proper error handling
   - `tracing.TracingMiddleware()`: Gin middleware for request tracing

## Usage Examples

### HTTP Handler

```go
func GetArticleHandler(c *gin.Context) {
    // Extract tracing context
    ctx := tracing.GetTracingContext(c).(context.Context)
    
    // Get article using the tracing context
    articleID := c.Param("id")
    articleService := services.NewArticleService()
    article, err := articleService.GetArticleByIDWithContext(ctx, articleID)
    
    // Response handling...
}
```

### Service Layer

```go
// Regular method
func (s *ArticleService) GetArticleByID(id string) (models.Article, error) {
    return s.GetArticleByIDWithContext(context.Background(), id)
}

// Context-aware method for tracing
func (s *ArticleService) GetArticleByIDWithContext(ctx context.Context, id string) (models.Article, error) {
    ctx, span := tracing.StartSpan(ctx, "ArticleService.GetArticleByID")
    defer span.End()
    
    span.SetAttributes(attribute.String("article.id", id))
    
    // Implementation...
    
    return article, err
}
```

### Database Operations

```go
func (r *ArticleRepository) FindByID(ctx context.Context, id string) (models.Article, error) {
    ctx, span := tracing.StartSpan(ctx, "ArticleRepository.FindByID")
    defer span.End()
    
    span.SetAttributes(attribute.String("db.operation", "find_by_id"))
    span.SetAttributes(attribute.String("db.id", id))
    
    // Implementation...
    
    return article, err
}
```

## Configuration

Tracing uses these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| OTEL_EXPORTER_OTLP_TRACES_ENDPOINT | OpenTelemetry collector endpoint | http://jaeger:4318 |
| OTEL_SERVICE_NAME | Service name in traces | news-service |

## Troubleshooting

1. **No Traces in Jaeger**:
   - Check OTEL_EXPORTER_OTLP_TRACES_ENDPOINT value
   - Verify collector is running via `docker ps`

2. **Missing Spans**:
   - Ensure context is properly propagated between functions
   - Check if the service method has a `WithContext` version

3. **Validation and Fixing**:
   - Run `./scripts/validate_tracing.sh` to check implementation
   - Run `./scripts/fix_tracing_issues.sh` to fix common issues

## Best Practices

1. **Context Propagation**: Always pass context from upstream calls
2. **Method Naming**: Use `MethodNameWithContext` for traced versions
3. **Error Handling**: Always record errors in spans with `span.RecordError()`
4. **Meaningful Attributes**: Add business context with `span.SetAttributes()`
5. **Sampling Strategy**: Use `AlwaysSample()` in development, but consider sampling in production

## Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- Detailed implementation guide: `docs/tracing_best_practices.md`
