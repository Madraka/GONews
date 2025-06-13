package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestStartSpan(t *testing.T) {
	// Create a span recorder to capture spans
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	// Set the global tracer provider for testing
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { otel.SetTracerProvider(originalTP) }()

	ctx := context.Background()
	spanName := "test-span"

	// Test StartSpan
	_, span := StartSpan(ctx, spanName)
	span.End()

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].Name() != spanName {
		t.Errorf("Expected span name %s, got %s", spanName, spans[0].Name())
	}
}

func TestStartSpanWithAttributes(t *testing.T) {
	// Create a span recorder to capture spans
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	// Set the global tracer provider for testing
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { otel.SetTracerProvider(originalTP) }()

	ctx := context.Background()
	spanName := "test-span-attrs"

	// Test StartSpanWithAttributes
	_, span := StartSpanWithAttributes(ctx, spanName,
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
	span.End()

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes()
	found1, found2 := false, false
	for _, attr := range attrs {
		if attr.Key == "key1" && attr.Value.AsString() == "value1" {
			found1 = true
		}
		if attr.Key == "key2" && attr.Value.AsInt64() == 42 {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Attribute key1=value1 not found")
	}
	if !found2 {
		t.Error("Attribute key2=42 not found")
	}
}

func TestEndSpan(t *testing.T) {
	// Create a span recorder to capture spans
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	// Set the global tracer provider for testing
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { otel.SetTracerProvider(originalTP) }()

	ctx := context.Background()

	// Test EndSpan with error
	_, span := StartSpan(ctx, "error-span")
	testErr := Error("test error")
	EndSpan(span, testErr)

	// Test EndSpan without error
	_, span = StartSpan(ctx, "success-span")
	EndSpan(span, nil)

	spans := sr.Ended()
	if len(spans) != 2 {
		t.Fatalf("Expected 2 spans, got %d", len(spans))
	}

	// Error span should have error status
	errorSpan := spans[0]
	if errorSpan.Status().Code != codes.Error {
		t.Errorf("Expected Error status, got %v", errorSpan.Status().Code)
	}

	// Success span should have OK status
	successSpan := spans[1]
	if successSpan.Status().Code != codes.Ok {
		t.Errorf("Expected Ok status, got %v", successSpan.Status().Code)
	}
}

// Helper for creating error
type Error string

func (e Error) Error() string { return string(e) }

func TestWrapFunction(t *testing.T) {
	// Create a span recorder to capture spans
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	// Set the tracer provider for testing
	originalTP := tracerProvider
	tracerProvider = tp
	defer func() { tracerProvider = originalTP }()

	ctx := context.Background()

	// Test WrapFunction with success
	err := WrapFunction(ctx, "success-function", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Test WrapFunction with error
	testErr := Error("function error")
	err = WrapFunction(ctx, "error-function", func(ctx context.Context) error {
		return testErr
	})
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	spans := sr.Ended()
	if len(spans) != 2 {
		t.Fatalf("Expected 2 spans, got %d", len(spans))
	}

	// Check function names
	if spans[0].Name() != "success-function" {
		t.Errorf("Expected span name success-function, got %s", spans[0].Name())
	}

	if spans[1].Name() != "error-function" {
		t.Errorf("Expected span name error-function, got %s", spans[1].Name())
	}

	// Check error status
	if spans[1].Status().Code != codes.Error {
		t.Errorf("Expected Error status for error-function, got %v", spans[1].Status().Code)
	}
}
