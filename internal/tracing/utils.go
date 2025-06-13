package tracing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPClient is an interface for HTTP clients that can be traced
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TraceHTTPRequest wraps an outgoing HTTP request with tracing context
func TraceHTTPRequest(ctx context.Context, req *http.Request) (context.Context, *http.Request, trace.Span) {
	spanName := fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)
	ctx, span := StartSpanWithAttributes(ctx, spanName,
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.host", req.URL.Host),
		attribute.String("span.kind", "client"),
	)

	// Inject tracing context into the outgoing request headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Return the context, request with context, and span
	return ctx, req.WithContext(ctx), span
}

// WrapHTTPClient wraps an HTTP client with tracing
func WrapHTTPClient(client HTTPClient) HTTPClient {
	return &tracedHTTPClient{client: client}
}

type tracedHTTPClient struct {
	client HTTPClient
}

func (c *tracedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	_, req, span := TraceHTTPRequest(req.Context(), req)
	defer span.End()

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	// Add response attributes
	span.SetAttributes(attribute.Int("http.request_duration_ms", int(duration.Milliseconds())))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	// Add response code and size
	if resp != nil {
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		span.SetAttributes(attribute.Int64("http.response_size", resp.ContentLength))

		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
	}

	return resp, nil
}

// TraceSQLQuery wraps a SQL query with tracing
func TraceSQLQuery(ctx context.Context, operation string, query string, args ...interface{}) (context.Context, trace.Span) {
	ctx, span := StartSpanWithAttributes(ctx, "DB."+operation,
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", sanitizeSQL(query)),
		attribute.String("db.operation", operation),
	)

	// Add arguments count but not the actual values for security
	span.SetAttributes(attribute.Int("db.args.count", len(args)))

	return ctx, span
}

// TraceDBExec wraps a database Exec call with tracing
func TraceDBExec(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	ctx, span := TraceSQLQuery(ctx, "Exec", query, args...)
	defer span.End()

	start := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	span.SetAttributes(attribute.Int("db.duration_ms", int(duration.Milliseconds())))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		// Add result information
		if rowsAffected, err := result.RowsAffected(); err == nil {
			span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
		}
	}

	return result, err
}

// TraceDBQuery wraps a database Query call with tracing
func TraceDBQuery(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, span := TraceSQLQuery(ctx, "Query", query, args...)
	defer span.End()

	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	span.SetAttributes(attribute.Int("db.duration_ms", int(duration.Milliseconds())))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return rows, err
}

// TraceGinHandler wraps a Gin handler function with tracing
func TraceGinHandler(operationName string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract any incoming trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header))

		// Start a new span
		ctx, span := StartSpanWithAttributes(ctx, operationName,
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)
		defer span.End()

		// Add trace ID to the response headers for debugging
		traceID := span.SpanContext().TraceID().String()
		c.Header("X-Trace-ID", traceID)

		// Store span in context for child handlers
		c.Set("tracing.span", span)
		c.Set("tracing.context", ctx)

		// Process the request
		c.Next()

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Handle errors
		if len(c.Errors) > 0 {
			span.RecordError(fmt.Errorf("gin errors: %s", c.Errors.String()))
			span.SetStatus(codes.Error, c.Errors.String())
		} else if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, http.StatusText(c.Writer.Status()))
		}
	}
}

// Helper function to sanitize SQL queries
// Removes potentially sensitive data from SQL statements
func sanitizeSQL(query string) string {
	// Basic implementation - in production you might want to use a more sophisticated SQL parser
	// This just ensures we don't include sensitive user data in traces
	return query
}

// TraceRedisOperation wraps Redis operations with tracing
func TraceRedisOperation(ctx context.Context, operation string, key string) (context.Context, trace.Span) {
	return StartSpanWithAttributes(ctx, "Redis."+operation,
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", operation),
		attribute.String("db.redis.key", key),
	)
}

// TraceCachedOperation wraps an operation with cache checking and tracing
func TraceCachedOperation(ctx context.Context, operationName string,
	cacheKey string,
	getCachedValue func(string) (interface{}, error),
	getFromSource func() (interface{}, error)) (interface{}, error) {

	ctx, span := StartSpan(ctx, operationName)
	defer span.End()

	// Try to get from cache
	ctx, cacheSpan := StartSpan(ctx, operationName+".Cache")
	cachedValue, err := getCachedValue(cacheKey)
	if err == nil {
		cacheSpan.SetAttributes(attribute.Bool("cache.hit", true))
		cacheSpan.End()
		span.SetAttributes(attribute.Bool("cache.hit", true))
		return cachedValue, nil
	}
	cacheSpan.SetAttributes(attribute.Bool("cache.hit", false))
	cacheSpan.End()

	// Get from source
	_, sourceSpan := StartSpan(ctx, operationName+".Source")
	value, err := getFromSource()
	if err != nil {
		sourceSpan.RecordError(err)
		sourceSpan.SetStatus(codes.Error, err.Error())
		sourceSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	sourceSpan.End()

	return value, nil
}
