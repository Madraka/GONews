package tracing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingHandler wraps an HTTP handler with tracing
func TracingHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		spanName := r.Method + " " + r.URL.Path

		ctx, span := StartSpanWithAttributes(ctx, spanName,
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.String("http.remote_addr", r.RemoteAddr),
		)
		defer span.End()

		// Add trace ID to response headers for debugging
		traceID := span.SpanContext().TraceID().String()
		w.Header().Set("X-Trace-ID", traceID)

		// Create a wrapped response writer to capture status code
		wrappedWriter := &responseWriterInterceptor{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Execute the handler with the traced context
		handler.ServeHTTP(wrappedWriter, r.WithContext(ctx))

		// Record response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", wrappedWriter.statusCode),
		)

		// If error status code, mark span as error
		if wrappedWriter.statusCode >= 400 {
			span.SetStatus(codes.Error, http.StatusText(wrappedWriter.statusCode))
		}
	})
}

// responseWriterInterceptor is a wrapper for http.ResponseWriter that captures the status code
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before calling the wrapped WriteHeader
func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// WithTracing adds tracing to a gin handler function
func WithTracing(operationName string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := StartSpanFromGin(c, operationName)
		defer span.End()

		// Store the traced context
		c.Set("tracing.context", ctx)

		// Call the handler
		handler(c)

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Handle errors
		if len(c.Errors) > 0 {
			span.RecordError(c.Errors[0])
			span.SetStatus(codes.Error, c.Errors.String())
		} else if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, http.StatusText(c.Writer.Status()))
		}
	}
}

// TraceRequest adds tracing to an outgoing HTTP request
func TraceRequest(ctx context.Context, req *http.Request) (context.Context, *http.Request, trace.Span) {
	spanName := req.Method + " " + req.URL.Path
	ctx, span := StartSpanWithAttributes(ctx, spanName,
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("span.kind", "client"),
	)

	// Inject tracing headers into request
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Return the enriched context and request along with the span
	return ctx, req.WithContext(ctx), span
}

// RecordError records an error in the current span
func RecordError(ctx context.Context, err error, message string) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(
		attribute.String("error.message", message),
	))
	span.SetStatus(codes.Error, message)
}

// AddAttribute adds an attribute to the current span
func AddAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}
