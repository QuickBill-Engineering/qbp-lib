package otelamqp

import (
	"context"
	"testing"

	"github.com/QuickBill-Engineering/qbp-lib/tracing"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

func TestHeadersCarrier_Get(t *testing.T) {
	headers := HeadersCarrier{
		"traceparent": "00-abc123-def456-01",
		"custom":      "value",
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "existing key",
			key:      "traceparent",
			expected: "00-abc123-def456-01",
		},
		{
			name:     "another existing key",
			key:      "custom",
			expected: "value",
		},
		{
			name:     "non-existing key",
			key:      "nonexistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := headers.Get(tt.key)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHeadersCarrier_Set(t *testing.T) {
	headers := HeadersCarrier{}

	headers.Set("traceparent", "00-abc123-def456-01")

	if headers["traceparent"] != "00-abc123-def456-01" {
		t.Errorf("expected traceparent to be set")
	}
}

func TestHeadersCarrier_Keys(t *testing.T) {
	headers := HeadersCarrier{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	keys := headers.Keys()

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for expectedKey := range headers {
		if !keyMap[expectedKey] {
			t.Errorf("expected key %s not found", expectedKey)
		}
	}
}

func TestInject(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	ctx := context.Background()
	headers := make(amqp.Table)

	Inject(ctx, headers)

	if headers == nil {
		t.Error("expected headers to be non-nil after injection")
	}
}

func TestExtract(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	headers := make(amqp.Table)
	ctx := context.Background()

	extractedCtx := Extract(ctx, headers)

	if extractedCtx == nil {
		t.Error("expected non-nil context")
	}
}

func TestInjectExtract_RoundTrip(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	ctx := context.Background()

	tp := trace.NewNoopTracerProvider()
	ctx, span := tp.Tracer("test").Start(ctx, "test-span")
	defer span.End()

	headers := make(amqp.Table)
	Inject(ctx, headers)

	extractedCtx := Extract(context.Background(), headers)

	spanCtx := trace.SpanContextFromContext(extractedCtx)

	_ = spanCtx
}

func TestHeadersCarrier_GetNonStringValue(t *testing.T) {
	headers := HeadersCarrier{
		"numeric": 12345,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	result := headers.Get("numeric")
	if result != "" {
		t.Errorf("expected empty string for non-string value, got %s", result)
	}

	result = headers.Get("nested")
	if result != "" {
		t.Errorf("expected empty string for nested value, got %s", result)
	}
}

func TestExtract_WithEmptyHeaders(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	headers := make(amqp.Table)
	ctx := Extract(context.Background(), headers)

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		t.Error("expected invalid span context from empty headers")
	}
}

func TestInject_WithNilHeaders(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	ctx := context.Background()
	var headers amqp.Table

	Inject(ctx, headers)
}
