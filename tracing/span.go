package tracing

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// FinishFunc is a function type that ends a span and optionally records an error.
// Pass a pointer to the function's named error return to automatically record
// errors on the span.
//
// If the error pointer is nil or points to a nil error, the span is marked as OK.
// If the error pointer points to a non-nil error, the error is recorded and
// the span status is set to Error.
//
// Example:
//
//	func (s *svc) Do(ctx context.Context) (err error) {
//	    ctx, finish := tracing.Start(ctx)
//	    defer finish(&err)
//
//	    // ... do work that might return an error
//	    return nil
//	}
type FinishFunc func(errPtr *error)

// Start creates a new span named after the calling function.
// The span name is derived from runtime.Caller and cleaned to be readable
// (e.g., "service.PaymentOrchestrator.Run").
//
// Source location (code.function, code.filepath, code.lineno) is automatically
// recorded as span attributes for debugging.
//
// Parameters:
//   - ctx: The parent context. If it contains an active span, the new span will be a child.
//   - attrs: Optional attributes to attach to the span.
//
// Returns:
//   - context.Context: A new context containing the span.
//   - FinishFunc: A function that must be called to end the span. Use defer.
//
// Example:
//
//	func (s *service) Process(ctx context.Context, id string) (err error) {
//	    ctx, finish := tracing.Start(ctx, tracing.Attr("id", id))
//	    defer finish(&err)
//
//	    return s.doWork(ctx)
//	}
func Start(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, FinishFunc) {
	return StartNamed(ctx, getCallerFunctionName(), attrs...)
}

// StartNamed creates a new span with an explicit name.
// Use this when the span name should differ from the function name,
// such as when creating spans for external service calls.
//
// Parameters:
//   - ctx: The parent context. If it contains an active span, the new span will be a child.
//   - name: The span name. Use dot-notation for hierarchy (e.g., "http.client", "db.query").
//   - attrs: Optional attributes to attach to the span.
//
// Returns:
//   - context.Context: A new context containing the span.
//   - FinishFunc: A function that must be called to end the span. Use defer.
//
// Example:
//
//	func (s *service) CallAPI(ctx context.Context) (err error) {
//	    ctx, finish := tracing.StartNamed(ctx, "external-api.call")
//	    defer finish(&err)
//
//	    return s.client.Call(ctx)
//	}
func StartNamed(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, FinishFunc) {
	return StartWithKind(ctx, trace.SpanKindInternal, name, attrs...)
}

// StartWithKind creates a span with a specific SpanKind.
// Use this to indicate the role of the span in the trace:
//   - trace.SpanKindClient: For outbound calls to external services (HTTP, gRPC, DB)
//   - trace.SpanKindServer: For incoming server requests (automatically set by otelgin.Middleware)
//   - trace.SpanKindProducer: For sending messages to a message broker
//   - trace.SpanKindConsumer: For consuming messages from a message broker
//   - trace.SpanKindInternal: For internal operations within the service (default)
//
// Parameters:
//   - ctx: The parent context.
//   - kind: The span kind (e.g., trace.SpanKindClient for outbound calls).
//   - name: The span name.
//   - attrs: Optional attributes to attach to the span.
//
// Returns:
//   - context.Context: A new context containing the span.
//   - FinishFunc: A function that must be called to end the span. Use defer.
//
// Example:
//
//	func (s *service) CallExternalAPI(ctx context.Context) (err error) {
//	    ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindClient, "http.client")
//	    defer finish(&err)
//
//	    return s.httpClient.Get(ctx, "/api/data")
//	}
func StartWithKind(ctx context.Context, kind trace.SpanKind, name string, attrs ...attribute.KeyValue) (context.Context, FinishFunc) {
	t := getTracer()
	if t == nil {
		return ctx, func(errPtr *error) {}
	}

	_, file, line, _ := runtime.Caller(1)
	file = filepath.Base(file)

	allAttrs := append([]attribute.KeyValue{
		attribute.String("code.function", name),
		attribute.String("code.filepath", file),
		attribute.Int("code.lineno", line),
	}, attrs...)

	ctx, span := t.Start(ctx, name,
		trace.WithSpanKind(kind),
		trace.WithAttributes(allAttrs...),
	)

	return ctx, func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
			span.SetStatus(codes.Error, (*errPtr).Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}
}

// AddAttrs adds attributes to the current span in the context.
// Safe to call even if there is no active span (no-op in that case).
//
// Attributes are key-value pairs that provide additional context about the span.
// Use the Attr() helper function to create type-safe attributes.
//
// Parameters:
//   - ctx: Context containing the current span.
//   - attrs: Attributes to add to the span.
//
// Example:
//
//	func (s *service) Process(ctx context.Context, user *User) {
//	    tracing.AddAttrs(ctx,
//	        tracing.Attr("user.id", user.ID),
//	        tracing.Attr("user.email", user.Email),
//	    )
//	}
func AddAttrs(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(attrs...)
}

// AddEvent records a timestamped event on the current span.
// Events are useful for marking specific points in time during a span's lifetime,
// such as logging when a particular operation started or completed.
//
// Safe to call even if there is no active span (no-op in that case).
//
// Parameters:
//   - ctx: Context containing the current span.
//   - name: The event name (e.g., "cache-hit", "retry-attempt").
//   - attrs: Optional attributes to attach to the event.
//
// Example:
//
//	func (s *service) Process(ctx context.Context) {
//	    tracing.AddEvent(ctx, "cache-check-started")
//
//	    if s.cache.Has(key) {
//	        tracing.AddEvent(ctx, "cache-hit", tracing.Attr("key", key))
//	        return s.cache.Get(key)
//	    }
//
//	    tracing.AddEvent(ctx, "cache-miss", tracing.Attr("key", key))
//	}
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

func getCallerFunctionName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	name := fn.Name()
	if idx := strings.LastIndex(name, "/"); idx != -1 {
		name = name[idx+1:]
	}

	return name
}
