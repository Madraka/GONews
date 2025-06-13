package tracing

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware returns a gin middleware that adds OpenTelemetry tracing to requests
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tracing context from incoming headers
		propagator := propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span for this request
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		ctx, span := StartSpanWithAttributes(ctx,
			spanName,
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPTargetKey.String(c.Request.URL.Path),
			semconv.HTTPRouteKey.String(c.FullPath()),
			semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)
		defer span.End()

		// Set trace ID in response header for debugging
		traceID := span.SpanContext().TraceID().String()
		c.Header("X-Trace-ID", traceID)

		// Store the span and context in the gin.Context
		c.Set("tracing.span", span)
		c.Set("tracing.context", ctx)

		// Process request
		c.Next()

		// Add response attributes after the request is processed
		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(status),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// If error occurred, mark the span accordingly
		if status >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
			// Add error details if available
			if len(c.Errors) > 0 {
				span.SetAttributes(attribute.String("error.details", c.Errors.String()))
			}
		}
	}
}

// GetSpanFromContext retrieves the current active span from the gin.Context
func GetSpanFromContext(c *gin.Context) trace.Span {
	if span, exists := c.Get("tracing.span"); exists {
		if s, ok := span.(trace.Span); ok {
			return s
		}
	}

	// Return a no-op span if not found
	return trace.SpanFromContext(c.Request.Context())
}

// GetTracingContext retrieves the tracing context from the gin.Context
func GetTracingContext(c *gin.Context) interface{} {
	if ctx, exists := c.Get("tracing.context"); exists {
		return ctx
	}

	return c.Request.Context()
}

// StartSpanFromGin starts a new span from a gin context with the given name
func StartSpanFromGin(c *gin.Context, spanName string) (context.Context, trace.Span) {
	// Get parent context from gin
	var parentCtx context.Context
	if ctx, exists := c.Get("tracing.context"); exists {
		if typedCtx, ok := ctx.(context.Context); ok {
			parentCtx = typedCtx
		} else {
			parentCtx = c.Request.Context()
		}
	} else {
		parentCtx = c.Request.Context()
	}

	// Start a new span
	return StartSpan(parentCtx, spanName)
}
