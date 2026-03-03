package tracelog

import (
	"context"
	"log/slog"
	"os"

	"github.com/QuickBill-Engineering/qbp-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

// Info logs a message at INFO level with trace context automatically extracted from ctx.
// The log entry includes trace_id, span_id, and request_id if available in the context.
//
// Parameters:
//   - ctx: The context containing trace information.
//   - msg: The log message.
//   - args: Optional key-value pairs to include in the log entry.
//
// Example:
//
//	tracelog.Info(ctx, "user logged in",
//	    "user_id", userID,
//	    "ip", clientIP,
//	)
func Info(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, args...)
}

// Error logs a message at ERROR level with trace context automatically extracted from ctx.
// The log entry includes trace_id, span_id, and request_id if available in the context.
//
// Parameters:
//   - ctx: The context containing trace information.
//   - msg: The log message.
//   - args: Optional key-value pairs to include in the log entry.
//
// Example:
//
//	if err != nil {
//	    tracelog.Error(ctx, "failed to process request",
//	        "error", err,
//	        "retry_count", retryCount,
//	    )
//	}
func Error(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
}

// Warn logs a message at WARN level with trace context automatically extracted from ctx.
// The log entry includes trace_id, span_id, and request_id if available in the context.
//
// Parameters:
//   - ctx: The context containing trace information.
//   - msg: The log message.
//   - args: Optional key-value pairs to include in the log entry.
//
// Example:
//
//	if deprecated {
//	    tracelog.Warn(ctx, "using deprecated API",
//	        "endpoint", endpoint,
//	    )
//	}
func Warn(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, msg, args...)
}

// Debug logs a message at DEBUG level with trace context automatically extracted from ctx.
// The log entry includes trace_id, span_id, and request_id if available in the context.
//
// Parameters:
//   - ctx: The context containing trace information.
//   - msg: The log message.
//   - args: Optional key-value pairs to include in the log entry.
//
// Example:
//
//	tracelog.Debug(ctx, "processing item",
//	    "item_id", itemID,
//	    "status", status,
//	)
func Debug(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, args...)
}

// Fatal logs a message at ERROR level with trace context, then calls os.Exit(1).
// Use this ONLY for unrecoverable errors in main(). Other code should return errors.
//
// WARNING: This exits the program immediately. Use sparingly and only in main().
//
// Parameters:
//   - ctx: The context containing trace information.
//   - msg: The log message.
//   - args: Optional key-value pairs to include in the log entry.
//
// Example:
//
//	if err := tracing.InitFromEnv(); err != nil {
//	    tracelog.Fatal(ctx, "failed to initialize tracing", "error", err)
//	}
func Fatal(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
	os.Exit(1)
}

func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger := WithTrace(ctx, slog.Default())
	logger.Log(ctx, level, msg, args...)
}

// WithTrace returns a new slog.Logger that automatically prepends trace_id, span_id,
// and request_id from ctx to every log entry.
//
// Use this when you need to pass a logger to third-party code or when you want
// to create a logger with additional context.
//
// Parameters:
//   - ctx: The context containing trace information.
//   - logger: The base logger to wrap.
//
// Returns:
//   - *slog.Logger: A new logger with trace attributes prepended.
//
// Example:
//
//	logger := tracelog.WithTrace(ctx, slog.Default())
//	logger.Info("processing request", "user_id", userID)
//
//	// Output: {"level":"INFO","trace_id":"abc123","span_id":"def456","request_id":"req-789",...}
func WithTrace(ctx context.Context, logger *slog.Logger) *slog.Logger {
	attrs := make([]any, 0, 4)

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	if requestID := tracing.RequestID(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	return logger.With(attrs...)
}
