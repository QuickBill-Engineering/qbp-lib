package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// contextKey is an unexported type for context keys.
// This prevents collisions with keys from other packages.
type contextKey int

const requestIDKey contextKey = iota

// WithRequestID returns a copy of ctx with the given request ID stored.
// The request ID can be retrieved later using RequestID().
// This is typically set by the otelgin.RequestID() middleware.
//
// Parameters:
//   - ctx: The context to add the request ID to.
//   - id: The request ID to store (typically a UUID).
//
// Returns:
//   - context.Context: A new context with the request ID.
//
// Example:
//
//	ctx := tracing.WithRequestID(ctx, "req-123")
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID retrieves the request ID from ctx, or an empty string if not set.
// The request ID is typically set by the otelgin.RequestID() middleware.
//
// Parameters:
//   - ctx: The context to retrieve the request ID from.
//
// Returns:
//   - string: The request ID, or "" if not set.
//
// Example:
//
//	requestID := tracing.RequestID(ctx)
//	if requestID != "" {
//	    log.Printf("request_id=%s", requestID)
//	}
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// DetachedContext returns a new background context that preserves only the
// trace context (span context) and request ID from the original context.
//
// Use this for goroutines that must outlive the original request but still
// need trace correlation. The returned context is not canceled when the
// original request context is canceled.
//
// This is useful for:
//   - Async operations that continue after the HTTP response is sent
//   - Background jobs triggered by a request
//   - Fire-and-forget operations that still need tracing
//
// Parameters:
//   - ctx: The original context to detach from.
//
// Returns:
//   - context.Context: A new background context with trace context and request ID preserved.
//
// Example:
//
//	func (s *service) ProcessAsync(ctx context.Context, job *Job) {
//	    // Detach from request context so goroutine survives after response
//	    detached := tracing.DetachedContext(ctx)
//
//	    go func() {
//	        ctx, finish := tracing.Start(detached, "async-job")
//	        defer finish(nil)
//
//	        s.processJob(ctx, job)
//	    }()
//	}
func DetachedContext(ctx context.Context) context.Context {
	bg := context.Background()

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		bg = trace.ContextWithSpanContext(bg, spanCtx)
	}

	if requestID := RequestID(ctx); requestID != "" {
		bg = WithRequestID(bg, requestID)
	}

	return bg
}
