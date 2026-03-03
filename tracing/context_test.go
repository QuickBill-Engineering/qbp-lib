package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-id")

	requestID := RequestID(ctx)
	if requestID != "test-request-id" {
		t.Errorf("expected request ID 'test-request-id', got %v", requestID)
	}
}

func TestRequestID_Empty(t *testing.T) {
	ctx := context.Background()
	requestID := RequestID(ctx)
	if requestID != "" {
		t.Errorf("expected empty request ID, got %v", requestID)
	}
}

func TestDetachedContext(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer shutdown(context.Background())

	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-id")

	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	detached := DetachedContext(ctx)

	requestID := RequestID(detached)
	if requestID != "test-request-id" {
		t.Errorf("expected request ID 'test-request-id', got %v", requestID)
	}
}

func TestDetachedContext_NoSpan(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-id")

	detached := DetachedContext(ctx)

	requestID := RequestID(detached)
	if requestID != "test-request-id" {
		t.Errorf("expected request ID 'test-request-id', got %v", requestID)
	}

	spanCtx := trace.SpanContextFromContext(detached)
	if spanCtx.IsValid() {
		t.Error("expected no valid span context")
	}
}
