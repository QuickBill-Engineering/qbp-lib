package tracelog

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/QuickBill-Engineering/qbp-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

func TestInfo(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	Info(ctx, "test message", "key", "value")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

func TestError(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	Error(ctx, "test error", "error", "test error value")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

func TestWarn(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	Warn(ctx, "test warning", "key", "value")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

func TestDebug(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx := context.Background()
	Debug(ctx, "test debug", "key", "value")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

func TestWithTrace_NoSpan(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	logger := WithTrace(ctx, baseLogger)

	logger.Info("test message")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

func TestWithTrace_WithRequestID(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	ctx = tracing.WithRequestID(ctx, "test-request-id")

	logger := WithTrace(ctx, baseLogger)
	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

func TestWithTrace_WithSpan(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("test").Start(ctx, "test-span")
	defer span.End()

	logger := WithTrace(ctx, baseLogger)
	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

func TestWithTrace_WithSpanAndRequestID(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	ctx = tracing.WithRequestID(ctx, "test-request-id")
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("test").Start(ctx, "test-span")
	defer span.End()

	logger := WithTrace(ctx, baseLogger)
	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

func TestInfo_WithMultipleArgs(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	Info(ctx, "test message",
		"key1", "value1",
		"key2", "value2",
		"key3", 123,
	)

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}
