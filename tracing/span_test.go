package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestStart(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	ctx, finish := Start(context.Background(), Attr("test", "value"))
	if ctx == nil {
		t.Error("expected non-nil context")
	}
	if finish == nil {
		t.Error("expected non-nil finish function")
	}

	finish(nil)
}

func TestStartNamed(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	ctx, finish := StartNamed(context.Background(), "test-span")
	if ctx == nil {
		t.Error("expected non-nil context")
	}
	if finish == nil {
		t.Error("expected non-nil finish function")
	}

	finish(nil)
}

func TestStartWithKind(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	ctx, finish := StartWithKind(
		context.Background(),
		1,
		"client-span",
		Attr("client", "test"),
	)
	if ctx == nil {
		t.Error("expected non-nil context")
	}
	if finish == nil {
		t.Error("expected non-nil finish function")
	}

	finish(nil)
}

func TestFinishWithError(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	ctx, finish := Start(context.Background())
	var testErr error = &testError{msg: "test error"}
	finish(&testErr)

	if ctx == nil {
		t.Error("expected non-nil context")
	}
}

func TestAddAttrs(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	ctx, finish := Start(context.Background())
	defer finish(nil)

	AddAttrs(ctx, Attr("key1", "value1"), Attr("key2", "value2"))
}

func TestAddEvent(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	ctx, finish := Start(context.Background())
	defer finish(nil)

	AddEvent(ctx, "test-event", attribute.String("event.attr", "value"))
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
