package tracing

import (
	"context"
	"errors"
	"testing"
)

func TestWrap(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer shutdown(context.Background())

	result, err := Wrap(context.Background(), func(ctx context.Context) (string, error) {
		return "success", nil
	}, Attr("test", "value"))

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestWrapWithError(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer shutdown(context.Background())

	testErr := errors.New("test error")
	result, err := Wrap(context.Background(), func(ctx context.Context) (string, error) {
		return "", testErr
	})

	if err != testErr {
		t.Errorf("expected test error, got %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestWrapVoid(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer shutdown(context.Background())

	err = WrapVoid(context.Background(), func(ctx context.Context) error {
		return nil
	}, Attr("test", "value"))

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestWrapVoidWithError(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer shutdown(context.Background())

	testErr := errors.New("test error")
	err = WrapVoid(context.Background(), func(ctx context.Context) error {
		return testErr
	})

	if err != testErr {
		t.Errorf("expected test error, got %v", err)
	}
}

func TestWrapNamed(t *testing.T) {
	shutdown, err := Init(WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}
	defer shutdown(context.Background())

	result, err := WrapNamed(context.Background(), "custom-span", func(ctx context.Context) (int, error) {
		return 42, nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != 42 {
		t.Errorf("expected result 42, got %v", result)
	}
}
