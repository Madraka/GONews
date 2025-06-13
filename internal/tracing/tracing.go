// filepath: internal/tracing/tracing.go
package tracing

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Global tracer provider
var tracerProvider *sdktrace.TracerProvider

// InitTracing initializes OpenTelemetry tracing
func InitTracing(serviceName string) (func(), error) {
	// Read environment variables for configuration
	// Try standard OTEL env var first, fallback to custom one
	collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if collectorURL == "" {
		collectorURL = os.Getenv("OTEL_COLLECTOR_URL")
	}
	if collectorURL == "" {
		// Default to Jaeger in Docker Compose setup
		collectorURL = "http://jaeger:4318"
	}

	// Debug logs
	fmt.Printf("Original collector URL: %s\n", collectorURL)

	// Extract just the host and port from the URL
	var endpoint string
	var host string
	var path string

	// Parse the URL to separate parts
	parsedURL, err := url.Parse(collectorURL)
	if err != nil {
		fmt.Printf("Error parsing URL %s: %v\n", collectorURL, err)
		// If parsing fails, use the original URL as fallback
		endpoint = collectorURL
	} else {
		// Successfully parsed URL
		host = parsedURL.Host
		path = parsedURL.Path

		// If host is empty, use the whole path (might be just a hostname)
		if host == "" {
			host = parsedURL.Path
			path = ""
		}

		fmt.Printf("Parsed URL - Host: %s, Path: %s\n", host, path)

		// No need to add http:// as the WithEndpoint function does not expect scheme
		endpoint = host
	}

	fmt.Printf("Using endpoint for OTLP exporter: %s\n", endpoint)

	// Create OTLP HTTP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // For development, use TLS in production
	)

	// Create exporter
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", os.Getenv("GIN_MODE")),
			attribute.String("service.version", "1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	// Configure trace provider with batch span processor
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return a cleanup function
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}, nil
}

// Tracer returns a named tracer from the global provider
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name and returns the new context and span
func StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return Tracer("news-service").Start(ctx, spanName)
}

// StartSpanWithAttributes starts a new span with the given name and attributes
// and returns the new context and span
func StartSpanWithAttributes(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := Tracer("news-service").Start(ctx, spanName)
	span.SetAttributes(attrs...)
	return ctx, span
}

// AddSpanTags adds multiple tags to the span in the given context
func AddSpanTags(ctx context.Context, tags map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	for k, v := range tags {
		switch val := v.(type) {
		case string:
			span.SetAttributes(attribute.String(k, val))
		case int:
			span.SetAttributes(attribute.Int(k, val))
		case int64:
			span.SetAttributes(attribute.Int64(k, val))
		case float64:
			span.SetAttributes(attribute.Float64(k, val))
		case bool:
			span.SetAttributes(attribute.Bool(k, val))
		default:
			span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
}

// EndSpan ends a span with an optional error
func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// WrapFunction wraps a function execution with a tracing span
// and returns the error from the function
func WrapFunction(ctx context.Context, spanName string, fn func(ctx context.Context) error) error {
	ctx, span := StartSpan(ctx, spanName)
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", r)
			}
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
			panic(r) // re-panic after recording
		}
	}()

	err := fn(ctx)
	EndSpan(span, err)
	return err
}
