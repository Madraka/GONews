# Tracing Best Practices for News Service

## Introduction

This document provides best practices for implementing and using distributed tracing in the News service. Effective tracing helps in understanding request flows, identifying bottlenecks, and troubleshooting issues in distributed systems.

## Core Principles

1. **Consistency**: Follow consistent naming conventions and span hierarchies
2. **Completeness**: Trace all important operations, especially those crossing service boundaries
3. **Context**: Include relevant metadata with each span to aid debugging
4. **Performance**: Balance visibility with overhead to avoid impacting service performance

## Implementing Tracing in Your Code

### Service Layer

The service layer is where most of your business logic resides and is a prime location for tracing:

```go
// Example of tracing a service method
func (s *NewsService) GetNewsWithContext(ctx context.Context, id string) (models.News, error) {
    // Create a span for this operation
    ctx, span := tracing.StartSpan(ctx, "NewsService.GetNews")
    defer span.End()
    
    // Add relevant metadata
    span.SetAttributes(attribute.String("news.id", id))
    
    // Business logic...
    news, err := repositories.GetNewsByID(id)
    
    // Record errors if they occur
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return models.News{}, err
    }
    
    // Add result metadata
    span.SetAttributes(
        attribute.String("news.title", news.Title),
        attribute.String("news.author", news.Author)
    )
    
    return news, nil
}
```

### Handler Layer

Handlers should utilize tracing middleware rather than implementing tracing directly:

```go
func SetupNewsRoutes(router *gin.Engine) {
    // Apply tracing middleware
    newsGroup := router.Group("/news", tracing.TracingMiddleware("news-service"))
    
    // Define routes
    newsGroup.GET("/:id", GetNewsHandler)
}

// Handler can still add additional span attributes
func GetNewsHandler(c *gin.Context) {
    // Extract the tracing context and span
    ctx := tracing.GetTracingContext(c).(context.Context)
    span := tracing.GetSpanFromContext(c)
    
    // Add handler-specific attributes
    span.SetAttributes(attribute.String("handler", "GetNewsHandler"))
    
    // Extract ID parameter
    id := c.Param("id")
    
    // Use the service with tracing context
    newsService := services.NewsService{}
    news, err := newsService.GetNewsByIdWithContext(ctx, id)
    
    // Handle response...
}
```

### Data Access Layer

Trace database operations to identify slow queries:

```go
func GetNewsByID(ctx context.Context, id string) (models.News, error) {
    ctx, span := tracing.StartSpan(ctx, "Repository.GetNewsByID")
    defer span.End()
    
    span.SetAttributes(attribute.String("db.query", "SELECT * FROM news WHERE id = ?"))
    span.SetAttributes(attribute.String("db.query.params", id))
    
    // Database operations...
    
    return news, err
}
```

### External Services

Always trace calls to external services:

```go
func CallExternalAPI(ctx context.Context, url string) ([]byte, error) {
    ctx, span := tracing.StartSpan(ctx, "ExternalAPI.Call")
    defer span.End()
    
    span.SetAttributes(attribute.String("http.url", url))
    
    // Create request with tracing headers
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    // Inject tracing context into headers
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
    
    // Make the request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    defer resp.Body.Close()
    
    span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
    
    // Read and return body...
}
```

## Naming Conventions

- **Span Names**: Use `ServiceName.OperationName` format (e.g., `NewsService.GetById`)
- **HTTP Spans**: Use `HTTP Method Path` format (e.g., `GET /news/:id`)
- **DB Spans**: Use `DB.Operation` format (e.g., `DB.Query` or `DB.Update`)

## Useful Attributes to Include

- **Service Context**: `service.name`, `service.version`
- **User Context**: `user.id` (if applicable, but anonymize sensitive data)
- **Request Details**: `http.method`, `http.url`, `http.status_code`
- **Database**: `db.system`, `db.statement` (sanitize sensitive data)
- **Business Context**: Entity IDs and relevant business information
- **Error Details**: Error types, messages, stack traces (sanitize sensitive data)

## Common Pitfalls

- **Missing Context Propagation**: Always pass the context with the span through the function calls
- **Not Ending Spans**: Ensure spans are properly closed, typically with defer
- **Overinstrumentation**: Too many spans can cause performance overhead and trace clutter
- **Underinstrumentation**: Missing critical spans makes the trace incomplete
- **Missing Error Handling**: Always record errors in spans
- **Missing Operation Status**: Set span status according to operation outcome

## Testing Tracing

- Manually verify trace data appears in your tracing system (Jaeger)
- Write integration tests that assert on trace presence
- Monitor trace sampling rates to ensure adequate coverage
- Test distributed context propagation between services

## Troubleshooting Tracing

- Check collector endpoints are correct in configuration
- Verify trace sampling is enabled
- Check for dropped spans in the collector logs
- Ensure context propagation headers are correctly configured