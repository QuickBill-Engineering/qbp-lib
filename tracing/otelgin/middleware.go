package otelgin

import (
	"unicode"

	"github.com/QuickBill-Engineering/qbp-lib/tracing"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	filter      func(*gin.Context) bool
	propagators propagation.TextMapCarrier
}

// Option configures the tracing middleware.
type Option func(*config)

// WithFilter sets a filter function that determines whether to skip tracing for a request.
// Return true to skip tracing for the request. This is useful for health check endpoints
// or other high-frequency endpoints that don't need tracing.
//
// Parameters:
//   - f: A function that receives the gin.Context and returns true to skip tracing.
//
// Example:
//
//	r.Use(otelgin.Middleware(
//	    otelgin.WithFilter(func(c *gin.Context) bool {
//	        // Skip tracing for health and metrics endpoints
//	        return c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics"
//	    }),
//	))
func WithFilter(f func(*gin.Context) bool) Option {
	return func(c *config) {
		c.filter = f
	}
}

// Middleware returns a Gin middleware that automatically traces HTTP requests.
//
// The middleware:
//  1. Extracts W3C traceparent/tracestate from incoming request headers
//  2. Creates a server span named after the matched route pattern (c.FullPath())
//  3. Records http.method, http.route, http.status_code as span attributes
//  4. Sets span status to Error for 4xx and 5xx responses
//  5. Injects trace context into response headers for distributed tracing
//
// This middleware should be added after RequestID() if you want request IDs in traces.
//
// Parameters:
//   - opts: Optional configuration options (e.g., WithFilter).
//
// Returns:
//   - gin.HandlerFunc: A Gin middleware function.
//
// Example:
//
//	r := gin.Default()
//	r.Use(otelgin.RequestID())
//	r.Use(otelgin.Middleware())
//
//	// Or with filter
//	r.Use(otelgin.Middleware(
//	    otelgin.WithFilter(func(c *gin.Context) bool {
//	        return c.Request.URL.Path == "/health"
//	    }),
//	))
func Middleware(opts ...Option) gin.HandlerFunc {
	cfg := &config{
		filter: nil,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return func(c *gin.Context) {
		if cfg.filter != nil && cfg.filter(c) {
			c.Next()
			return
		}

		req := c.Request
		ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

		spanName := c.FullPath()
		if spanName == "" {
			spanName = req.URL.Path
		}

		ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("qbp-lib").Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(req.Method),
				semconv.URLPath(req.URL.Path),
				attribute.String("http.route", c.FullPath()),
			),
		)
		defer span.End()

		c.Request = req.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))
		if status >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
	}
}

// RequestID returns a Gin middleware that manages request IDs.
//
// The middleware:
//  1. Reads X-Request-ID from the request header
//  2. If not present, generates a new UUID
//  3. Stores it in the context via tracing.WithRequestID()
//  4. Sets X-Request-ID on the response header
//
// This enables request correlation across services and in logs.
// Use this before the tracing Middleware to include request IDs in spans.
//
// Returns:
//   - gin.HandlerFunc: A Gin middleware function.
//
// Example:
//
//	r := gin.Default()
//	r.Use(otelgin.RequestID())
//	r.Use(otelgin.Middleware())
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := sanitizeRequestID(c.GetHeader("X-Request-ID"))

		ctx := tracing.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

const maxRequestIDLen = 128

// sanitizeRequestID validates and sanitizes an incoming X-Request-ID header value.
// It enforces a maximum length, rejects non-printable characters and newlines,
// and regenerates a UUID if the input is invalid or empty.
func sanitizeRequestID(id string) string {
	if id == "" || len(id) > maxRequestIDLen {
		return uuid.New().String()
	}
	for _, r := range id {
		if r == '\n' || r == '\r' || !unicode.IsPrint(r) {
			return uuid.New().String()
		}
	}
	return id
}
