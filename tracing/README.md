# tracing

OpenTelemetry distributed tracing for Go services. Provides tracer initialization, span creation, and integrations for Gin, GORM, AMQP, and structured logging.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Programmatic Configuration](#programmatic-configuration)
  - [Configuration Options Reference](#configuration-options-reference)
- [Core API](#core-api)
  - [Init / InitFromEnv](#init--initfromenv)
  - [Start](#start)
  - [StartNamed](#startnamed)
  - [StartWithKind](#startwithkind)
  - [Wrap / WrapVoid / WrapNamed](#wrap--wrapvoid--wrapnamed)
  - [Attr](#attr)
  - [AddAttrs](#addattrs)
  - [AddEvent](#addevent)
  - [FinishFunc](#finishfunc)
- [Context Utilities](#context-utilities)
  - [WithRequestID / RequestID](#withrequestid--requestid)
  - [DetachedContext](#detachedcontext)
- [Integrations](#integrations)
  - [otelgin -- Gin HTTP Middleware](#otelgin----gin-http-middleware)
  - [otelgorm -- GORM Database Plugin](#otelgorm----gorm-database-plugin)
  - [otelamqp -- RabbitMQ Propagation](#otelamqp----rabbitmq-propagation)
  - [tracelog -- Structured Logging](#tracelog----structured-logging)
- [Patterns and Best Practices](#patterns-and-best-practices)
- [Complete Service Example](#complete-service-example)
- [Migration Guide](#migration-guide)

---

## Installation

```bash
go get github.com/QuickBill-Engineering/qbp-lib
```

Import the packages you need:

```go
import (
    "github.com/QuickBill-Engineering/qbp-lib/tracing"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog"
)
```

---

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/QuickBill-Engineering/qbp-lib/tracing"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog"
    "github.com/gin-gonic/gin"
)

func main() {
    // Initialize tracing from environment variables
    shutdown, err := tracing.InitFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(context.Background())

    r := gin.Default()
    r.Use(otelgin.RequestID())
    r.Use(otelgin.Middleware())

    r.GET("/users/:id", func(c *gin.Context) {
        ctx := c.Request.Context()
        tracelog.Info(ctx, "fetching user", "id", c.Param("id"))

        user, err := GetUser(ctx, c.Param("id"))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, user)
    })

    r.Run(":8080")
}

func GetUser(ctx context.Context, id string) (*User, error) {
    return tracing.Wrap(ctx, func(ctx context.Context) (*User, error) {
        // This function body runs inside a span named "main.GetUser"
        return &User{ID: id, Name: "John"}, nil
    }, tracing.Attr("user.id", id))
}
```

---

## Configuration

### Environment Variables

Set these before calling `tracing.InitFromEnv()`:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `OTEL_ENABLED` | bool | `false` | Enable or disable tracing. When disabled, a no-op provider is installed and all tracing calls are zero-cost. |
| `OTEL_ENDPOINT` | string | `localhost:4317` | OpenTelemetry collector address in `host:port` format. Use port `4317` for gRPC, `4318` for HTTP. |
| `OTEL_SERVICE_NAME` | string | binary name | Service name that appears in your tracing backend (Jaeger, Tempo, etc.). |
| `OTEL_SERVICE_VERSION` | string | `0.0.0` | Service version for trace attribution. |
| `OTEL_SAMPLING_RATE` | float | `1.0` | Fraction of traces to sample. `1.0` = 100%, `0.1` = 10%, `0.0` = none. Values outside `[0.0, 1.0]` are clamped. |
| `OTEL_EXPORTER_TYPE` | string | `grpc` | Transport protocol: `grpc` (recommended), `http`, or `stdout` (for local debugging). |
| `OTEL_INSECURE` | bool | `true` | When `true`, TLS is disabled. Set to `false` in production with TLS-enabled collectors. |
| `ENVIRONMENT` | string | `local` | Deployment environment. Common values: `local`, `development`, `staging`, `production`. |

### Programmatic Configuration

Use `tracing.Init()` with functional options when you need fine-grained control:

```go
shutdown, err := tracing.Init(
    tracing.WithEnabled(true),
    tracing.WithServiceName("payment-service"),
    tracing.WithServiceVersion("1.2.3"),
    tracing.WithEndpoint("otel-collector:4317"),
    tracing.WithEnvironment("production"),
    tracing.WithSamplingRate(0.5),
    tracing.WithExporter(tracing.ExporterGRPC),
    tracing.WithInsecure(false),
)
if err != nil {
    log.Fatal(err)
}
defer shutdown(context.Background())
```

You can also combine env vars with programmatic overrides:

```go
// Load from env, then override specific values
opts := tracing.ConfigFromEnv()
opts = append(opts, tracing.WithServiceName("my-override"))
shutdown, err := tracing.Init(opts...)
```

### Configuration Options Reference

| Function | Type | Default | Description |
|----------|------|---------|-------------|
| `WithEnabled(v bool)` | `bool` | `false` | Enable/disable tracing |
| `WithEndpoint(v string)` | `string` | `localhost:4317` | Collector address |
| `WithServiceName(v string)` | `string` | `os.Args[0]` | Service name |
| `WithServiceVersion(v string)` | `string` | `0.0.0` | Service version |
| `WithEnvironment(v string)` | `string` | `$ENVIRONMENT` or `local` | Deployment environment |
| `WithSamplingRate(v float64)` | `float64` | `1.0` | Sampling rate (clamped to 0.0-1.0) |
| `WithExporter(v ExporterType)` | `ExporterType` | `ExporterGRPC` | Transport protocol |
| `WithInsecure(v bool)` | `bool` | `true` | Disable TLS |

**Exporter types:**

| Constant | Value | Description |
|----------|-------|-------------|
| `tracing.ExporterGRPC` | `1` | gRPC transport (recommended for production) |
| `tracing.ExporterHTTP` | `2` | HTTP/Protobuf transport (use behind HTTP load balancers) |
| `tracing.ExporterStdout` | `3` | Writes JSON to stdout (local development only) |

---

## Core API

### Init / InitFromEnv

Initialize the global tracer provider. Call once at application startup in `main()`.

```go
// From environment variables
shutdown, err := tracing.InitFromEnv()

// Programmatic
shutdown, err := tracing.Init(
    tracing.WithEnabled(true),
    tracing.WithServiceName("my-service"),
)
```

**Returns:**
- `shutdown func(context.Context) error` -- Flushes pending spans and shuts down. Always defer this.
- `err error` -- Non-nil if initialization failed.

**Important:** When tracing is disabled (`WithEnabled(false)`), `Init` still succeeds and returns a valid no-op shutdown function. All span creation becomes zero-cost.

---

### Start

Creates a span named after the calling function. This is the primary way to trace functions.

```go
func Start(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, FinishFunc)
```

The span name is automatically derived from the calling function name using `runtime.Caller`. For example, if called inside `service.PaymentOrchestrator.Run`, the span name is `service.PaymentOrchestrator.Run`.

```go
func (s *service) Process(ctx context.Context, id string) (err error) {
    ctx, finish := tracing.Start(ctx, tracing.Attr("id", id))
    defer finish(&err)

    // All work here is inside the span.
    // If err is non-nil when the function returns, the span is marked as Error.
    return s.repo.Save(ctx, id)
}
```

**Key behavior:**
- Pass `&err` (pointer to named return) to automatically record errors on the span.
- Pass `nil` to `finish` if you don't need automatic error recording.
- Span attributes `code.function`, `code.filepath`, and `code.lineno` are automatically set.

---

### StartNamed

Creates a span with an explicit name instead of deriving it from the caller.

```go
func StartNamed(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, FinishFunc)
```

```go
ctx, finish := tracing.StartNamed(ctx, "external-api.call")
defer finish(&err)
```

---

### StartWithKind

Creates a span with a specific `SpanKind`, indicating the role of the span in a distributed trace.

```go
func StartWithKind(ctx context.Context, kind trace.SpanKind, name string, attrs ...attribute.KeyValue) (context.Context, FinishFunc)
```

```go
// Client span (outbound call to another service)
ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindClient, "http.client.call")
defer finish(&err)

// Producer span (publishing a message)
ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindProducer, "amqp.publish")
defer finish(&err)
```

**SpanKind values:**

| Kind | Use Case |
|------|----------|
| `trace.SpanKindInternal` | Internal operations (default for `Start`/`StartNamed`) |
| `trace.SpanKindServer` | Incoming server requests (set automatically by `otelgin.Middleware`) |
| `trace.SpanKindClient` | Outbound calls to external services (HTTP, gRPC, DB) |
| `trace.SpanKindProducer` | Publishing messages to a queue/topic |
| `trace.SpanKindConsumer` | Consuming messages from a queue/topic |

---

### Wrap / WrapVoid / WrapNamed

Functional-style span creation. Wraps a function call in a span and automatically handles error recording.

```go
// Wrap -- for functions returning (T, error)
func Wrap[T any](ctx context.Context, fn func(context.Context) (T, error), attrs ...attribute.KeyValue) (T, error)

// WrapVoid -- for functions returning only error
func WrapVoid(ctx context.Context, fn func(context.Context) error, attrs ...attribute.KeyValue) error

// WrapNamed -- Wrap with explicit span name
func WrapNamed[T any](ctx context.Context, name string, fn func(context.Context) (T, error), attrs ...attribute.KeyValue) (T, error)
```

```go
// Wrap: span name auto-derived from calling function
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
    return tracing.Wrap(ctx, func(ctx context.Context) (*User, error) {
        return s.repo.FindByID(ctx, id)
    }, tracing.Attr("user.id", id))
}

// WrapVoid: for operations that return only error
func (s *service) DeleteUser(ctx context.Context, id string) error {
    return tracing.WrapVoid(ctx, func(ctx context.Context) error {
        return s.repo.Delete(ctx, id)
    }, tracing.Attr("user.id", id))
}

// WrapNamed: explicit span name
func (s *service) CallAPI(ctx context.Context) (*Response, error) {
    return tracing.WrapNamed(ctx, "external-api.call", func(ctx context.Context) (*Response, error) {
        return s.client.Get(ctx, "/api/data")
    })
}
```

**When to use Wrap vs Start:**
- Use `Wrap` when the entire function body is the span and you want minimal boilerplate.
- Use `Start` when you need the span context for intermediate operations (e.g., adding events, calling `AddAttrs`).

---

### Attr

Creates a type-safe OpenTelemetry attribute with automatic type detection.

```go
func Attr(key string, value interface{}) attribute.KeyValue
```

```go
tracing.Attr("user.id", "abc-123")           // string
tracing.Attr("retry_count", 3)               // int
tracing.Attr("amount", int64(10000))          // int64
tracing.Attr("score", 3.14)                  // float64
tracing.Attr("active", true)                 // bool
tracing.Attr("tags", []string{"a", "b"})     // string slice
tracing.Attr("ids", []int{1, 2, 3})          // int slice
tracing.Attr("values", []int64{10, 20})      // int64 slice
tracing.Attr("rates", []float64{0.1, 0.5})   // float64 slice
tracing.Attr("flags", []bool{true, false})   // bool slice
tracing.Attr("user", userWithStringer)       // calls .String() if implements fmt.Stringer
tracing.Attr("data", anyOtherType)           // falls back to fmt.Sprintf("%v", v)
```

Use dot-notation for attribute keys to namespace them: `user.id`, `http.status_code`, `db.name`, etc.

---

### AddAttrs

Adds attributes to the current active span in the context. Safe to call even if there is no active span (no-op).

```go
func AddAttrs(ctx context.Context, attrs ...attribute.KeyValue)
```

```go
func (s *service) Process(ctx context.Context, user *User) {
    tracing.AddAttrs(ctx,
        tracing.Attr("user.id", user.ID),
        tracing.Attr("user.tier", user.Tier),
    )
}
```

---

### AddEvent

Records a timestamped event on the current span. Useful for marking points of interest within a span's lifetime.

```go
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)
```

```go
func (s *service) Process(ctx context.Context, key string) {
    tracing.AddEvent(ctx, "cache-check")

    if s.cache.Has(key) {
        tracing.AddEvent(ctx, "cache-hit", tracing.Attr("key", key))
        return
    }

    tracing.AddEvent(ctx, "cache-miss", tracing.Attr("key", key))
}
```

---

### FinishFunc

The function type returned by `Start`, `StartNamed`, and `StartWithKind`. Ends the span and optionally records an error.

```go
type FinishFunc func(errPtr *error)
```

**Behavior:**
- `finish(nil)` -- Marks span as OK and ends it.
- `finish(&err)` where `err == nil` -- Marks span as OK and ends it.
- `finish(&err)` where `err != nil` -- Records the error, sets span status to Error, and ends it.

**The `&err` pattern:** Always use named returns with `defer finish(&err)` so the span captures errors from any return path:

```go
func (s *service) DoWork(ctx context.Context) (err error) {
    ctx, finish := tracing.Start(ctx)
    defer finish(&err)

    if err := s.step1(ctx); err != nil {
        return err // span automatically marked as Error
    }
    return s.step2(ctx) // if this fails, span is also marked as Error
}
```

---

## Context Utilities

### WithRequestID / RequestID

Store and retrieve a request ID in the context.

```go
func WithRequestID(ctx context.Context, id string) context.Context
func RequestID(ctx context.Context) string
```

```go
// Store
ctx := tracing.WithRequestID(ctx, "req-abc-123")

// Retrieve
requestID := tracing.RequestID(ctx)
// Returns "" if no request ID was set
```

The `otelgin.RequestID()` middleware sets this automatically from the `X-Request-ID` header (or generates a UUID if absent).

---

### DetachedContext

Creates a new `context.Background()` that preserves only trace context and request ID from the original context. The returned context is **not** canceled when the original request context is canceled.

```go
func DetachedContext(ctx context.Context) context.Context
```

Use this for goroutines that must outlive the original HTTP request:

```go
func (s *service) ProcessAsync(ctx context.Context, job *Job) {
    detached := tracing.DetachedContext(ctx)

    go func() {
        // This goroutine survives after the HTTP response is sent,
        // but spans are still linked to the original trace.
        ctx, finish := tracing.Start(detached, tracing.Attr("job.id", job.ID))
        defer finish(nil)

        s.processJob(ctx, job)
    }()
}
```

**What is preserved:**
- Trace context (trace ID, span ID) -- so child spans link to the original trace
- Request ID

**What is NOT preserved:**
- Cancellation signal -- the detached context has no deadline/cancel
- Any other context values

---

## Integrations

### otelgin -- Gin HTTP Middleware

Import: `github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin`

#### Middleware()

Automatically traces every HTTP request. Creates a server-kind span named after the matched route pattern.

```go
func Middleware(opts ...Option) gin.HandlerFunc
```

**What it does:**
1. Extracts incoming W3C `traceparent`/`tracestate` headers (distributed trace propagation)
2. Creates a `SpanKindServer` span named after the route (e.g., `/users/:id`)
3. Records `http.request.method`, `url.path`, `http.route`, and `http.response.status_code`
4. Sets span status to Error for HTTP 4xx and 5xx responses
5. Injects trace context into response headers

```go
r := gin.Default()

// Basic usage
r.Use(otelgin.Middleware())

// With filter to skip certain endpoints
r.Use(otelgin.Middleware(
    otelgin.WithFilter(func(c *gin.Context) bool {
        return c.Request.URL.Path == "/health" ||
               c.Request.URL.Path == "/metrics"
    }),
))
```

#### RequestID()

Manages `X-Request-ID` headers. Reads from request, generates UUID if absent, stores in context, sets on response.

```go
func RequestID() gin.HandlerFunc
```

```go
r := gin.Default()
r.Use(otelgin.RequestID())  // Always add BEFORE Middleware()
r.Use(otelgin.Middleware())
```

**Security:** The middleware validates incoming `X-Request-ID` values:
- Maximum 128 characters
- No non-printable characters or newlines (prevents log injection)
- Invalid values are replaced with a generated UUID

#### Middleware order

Always register `RequestID()` before `Middleware()` so the request ID is available when the tracing span is created:

```go
r.Use(otelgin.RequestID())   // 1st: sets request ID in context
r.Use(otelgin.Middleware())   // 2nd: creates span (request ID already available)
```

---

### otelgorm -- GORM Database Plugin

Import: `github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm`

Automatically traces all GORM database operations (Create, Query, Update, Delete, Row, Raw). Each operation produces a `SpanKindClient` span.

```go
func NewPlugin(opts ...Option) gorm.Plugin
```

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
    log.Fatal(err)
}

err = db.Use(otelgorm.NewPlugin(
    otelgorm.WithDBName("payments"),
    otelgorm.WithLogQueries(true),
    otelgorm.WithRecordRowsAffected(true),
))
```

#### Options

| Function | Type | Default | Description |
|----------|------|---------|-------------|
| `WithDBName(name string)` | `string` | `""` | Database name (set as `db.name` attribute) |
| `WithLogQueries(enabled bool)` | `bool` | `false` | Record SQL statements in `db.statement`. **WARNING:** May expose sensitive data. |
| `WithRecordRowsAffected(enabled bool)` | `bool` | `false` | Record `db.rows_affected` count |
| `WithRecordSQLParameters(enabled bool)` | `bool` | `false` | Record SQL parameters. **WARNING:** Only enable in development. |
| `WithExcludeTables(tables ...string)` | `[]string` | `[]` | Skip tracing for specific tables (e.g., sessions, audit logs) |

#### Span attributes

Each database span includes:

| Attribute | Example | Description |
|-----------|---------|-------------|
| `db.system` | `sql` | Database system |
| `db.operation` | `Query`, `Create` | Operation type |
| `db.sql.table` | `users` | Target table |
| `db.name` | `payments` | Database name (if configured) |
| `db.statement` | `SELECT * FROM users WHERE id = ?` | SQL query (if `WithLogQueries(true)`) |
| `db.rows_affected` | `1` | Rows affected (if `WithRecordRowsAffected(true)`) |

#### Span naming

Span names follow the pattern `{table}.{operation}`:
- `users.Query`
- `payments.Create`
- `orders.Delete`

If the table name is not available, the span name is just the operation (e.g., `Raw`).

---

### otelamqp -- RabbitMQ Propagation

Import: `github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp`

Propagates W3C trace context through RabbitMQ message headers, enabling end-to-end tracing across services communicating via message queues.

#### Inject (Publisher side)

Injects trace context from the current context into AMQP message headers.

```go
func Inject(ctx context.Context, headers amqp.Table)
```

```go
func (p *publisher) Publish(ctx context.Context, body []byte) error {
    ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindProducer, "order.publish")
    defer finish(nil)

    headers := make(amqp.Table)
    otelamqp.Inject(ctx, headers)

    return p.ch.PublishWithContext(ctx,
        "orders",       // exchange
        "order.created", // routing key
        false, false,
        amqp.Publishing{
            Headers:     headers,
            ContentType: "application/json",
            Body:        body,
        },
    )
}
```

#### Extract (Consumer side)

Extracts trace context from AMQP message headers into a new context.

```go
func Extract(ctx context.Context, headers amqp.Table) context.Context
```

```go
func (c *consumer) HandleDelivery(d amqp.Delivery) {
    // Extract trace context -- the consumer span will be linked to the producer span
    ctx := otelamqp.Extract(context.Background(), d.Headers)

    ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindConsumer, "order.process")
    defer finish(nil)

    // Process message with full trace lineage...
    c.processOrder(ctx, d.Body)
}
```

#### HeadersCarrier

A type adapter implementing `propagation.TextMapCarrier` over `amqp.Table`. Used internally by `Inject`/`Extract`, but exposed if you need custom propagation logic.

```go
type HeadersCarrier amqp.Table

// Methods: Get(key string) string, Set(key string, value string), Keys() []string
```

---

### tracelog -- Structured Logging

Import: `github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog`

Structured logging functions that automatically extract and include `trace_id`, `span_id`, and `request_id` from the context. Built on Go's standard `log/slog`.

#### Log functions

```go
func Info(ctx context.Context, msg string, args ...any)
func Error(ctx context.Context, msg string, args ...any)
func Warn(ctx context.Context, msg string, args ...any)
func Debug(ctx context.Context, msg string, args ...any)
func Fatal(ctx context.Context, msg string, args ...any) // logs then os.Exit(1)
```

```go
tracelog.Info(ctx, "payment processed",
    "amount", 50000,
    "currency", "IDR",
    "method", "bank_transfer",
)
// Output includes: trace_id, span_id, request_id, msg, amount, currency, method

tracelog.Error(ctx, "payment failed",
    "error", err,
    "user_id", userID,
)

tracelog.Warn(ctx, "rate limit approaching",
    "current_rate", 95,
    "limit", 100,
)

tracelog.Debug(ctx, "cache lookup",
    "key", cacheKey,
    "hit", true,
)
```

#### WithTrace

Wraps an existing `*slog.Logger` to automatically prepend trace fields to every log entry. Useful when passing loggers to third-party code.

```go
func WithTrace(ctx context.Context, logger *slog.Logger) *slog.Logger
```

```go
logger := tracelog.WithTrace(ctx, slog.Default())
logger.Info("processing", "step", 1)
// Output: {"level":"INFO","trace_id":"abc123","span_id":"def456","request_id":"req-789","msg":"processing","step":1}
```

#### Automatically injected fields

| Field | Source | Example |
|-------|--------|---------|
| `trace_id` | OpenTelemetry span context | `a1b2c3d4e5f6...` |
| `span_id` | OpenTelemetry span context | `1a2b3c4d...` |
| `request_id` | `tracing.RequestID(ctx)` | `550e8400-e29b-41d4-a716-446655440000` |

Fields are only included if available in the context. If tracing is disabled or no span is active, trace_id and span_id are omitted.

---

## Patterns and Best Practices

### Always use named error returns with defer

This ensures the span captures errors from any return path:

```go
func (s *service) DoWork(ctx context.Context) (err error) {
    ctx, finish := tracing.Start(ctx)
    defer finish(&err) // captures error from any return path

    if err := s.step1(ctx); err != nil {
        return fmt.Errorf("step1: %w", err)
    }
    return s.step2(ctx)
}
```

### Use Attr for readable, type-safe attributes

```go
// Good: readable, type-safe
tracing.Start(ctx,
    tracing.Attr("user.id", userID),
    tracing.Attr("order.total", total),
)

// Avoid: raw OTel API is verbose
tracing.Start(ctx,
    attribute.String("user.id", userID),
    attribute.Float64("order.total", total),
)
```

### Use DetachedContext for async goroutines

Never pass the HTTP request context directly to goroutines that outlive the request:

```go
// WRONG: goroutine dies when request context is canceled
go func() {
    s.process(ctx, job) // ctx canceled after response sent
}()

// CORRECT: detach the context
go func() {
    detached := tracing.DetachedContext(ctx)
    ctx, finish := tracing.Start(detached, tracing.Attr("job.id", job.ID))
    defer finish(nil)
    s.process(ctx, job)
}()
```

### Filter noisy endpoints

Skip tracing for health checks and metrics to reduce trace volume:

```go
r.Use(otelgin.Middleware(
    otelgin.WithFilter(func(c *gin.Context) bool {
        switch c.Request.URL.Path {
        case "/health", "/ready", "/metrics":
            return true
        }
        return false
    }),
))
```

### Use SpanKind for correct trace visualization

Setting the correct `SpanKind` helps tracing backends visualize service dependencies:

```go
// Outbound HTTP call to another service
ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindClient, "billing-api.charge")

// Publishing a message to RabbitMQ
ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindProducer, "order.publish")

// Consuming a message from RabbitMQ
ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindConsumer, "order.process")
```

### Protect sensitive data in GORM queries

```go
// Development: log everything
db.Use(otelgorm.NewPlugin(
    otelgorm.WithLogQueries(true),
    otelgorm.WithRecordSQLParameters(true),
))

// Production: no query logging
db.Use(otelgorm.NewPlugin(
    otelgorm.WithDBName("payments"),
    otelgorm.WithRecordRowsAffected(true),
    // WithLogQueries and WithRecordSQLParameters default to false
))
```

### Exclude high-frequency tables

```go
db.Use(otelgorm.NewPlugin(
    otelgorm.WithExcludeTables("sessions", "health_checks", "audit_log"),
))
```

---

## Complete Service Example

A full example showing all integrations working together:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/QuickBill-Engineering/qbp-lib/tracing"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog"
    "github.com/gin-gonic/gin"
    amqp "github.com/rabbitmq/amqp091-go"
    "go.opentelemetry.io/otel/trace"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // 1. Initialize tracing
    shutdown, err := tracing.Init(
        tracing.WithEnabled(true),
        tracing.WithServiceName("order-service"),
        tracing.WithEndpoint("otel-collector:4317"),
        tracing.WithEnvironment("production"),
        tracing.WithSamplingRate(0.5),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(context.Background())

    // 2. Set up database with tracing
    db, err := gorm.Open(postgres.Open("host=localhost dbname=orders"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    db.Use(otelgorm.NewPlugin(
        otelgorm.WithDBName("orders"),
        otelgorm.WithRecordRowsAffected(true),
    ))

    // 3. Set up RabbitMQ
    conn, _ := amqp.Dial("amqp://localhost:5672")
    ch, _ := conn.Channel()

    svc := &OrderService{db: db, ch: ch}

    // 4. Set up Gin with tracing middleware
    r := gin.Default()
    r.Use(otelgin.RequestID())
    r.Use(otelgin.Middleware(
        otelgin.WithFilter(func(c *gin.Context) bool {
            return c.Request.URL.Path == "/health"
        }),
    ))

    r.POST("/orders", svc.CreateOrder)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    r.Run(":8080")
}

type OrderService struct {
    db *gorm.DB
    ch *amqp.Channel
}

type Order struct {
    ID     string  `json:"id" gorm:"primaryKey"`
    UserID string  `json:"user_id"`
    Amount float64 `json:"amount"`
}

func (s *OrderService) CreateOrder(c *gin.Context) {
    ctx := c.Request.Context()

    var order Order
    if err := c.ShouldBindJSON(&order); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := s.saveOrder(ctx, &order); err != nil {
        tracelog.Error(ctx, "failed to create order", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }

    // Publish event asynchronously
    detached := tracing.DetachedContext(ctx)
    go s.publishOrderCreated(detached, &order)

    tracelog.Info(ctx, "order created",
        "order_id", order.ID,
        "amount", order.Amount,
    )
    c.JSON(http.StatusCreated, order)
}

func (s *OrderService) saveOrder(ctx context.Context, order *Order) (err error) {
    ctx, finish := tracing.Start(ctx,
        tracing.Attr("order.id", order.ID),
        tracing.Attr("order.amount", order.Amount),
    )
    defer finish(&err)

    return s.db.WithContext(ctx).Create(order).Error
}

func (s *OrderService) publishOrderCreated(ctx context.Context, order *Order) {
    ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindProducer, "order.publish",
        tracing.Attr("order.id", order.ID),
    )
    defer finish(nil)

    headers := make(amqp.Table)
    otelamqp.Inject(ctx, headers)

    body := []byte(fmt.Sprintf(`{"order_id":"%s"}`, order.ID))
    err := s.ch.PublishWithContext(ctx, "orders", "order.created", false, false,
        amqp.Publishing{
            Headers:     headers,
            ContentType: "application/json",
            Body:        body,
        },
    )
    if err != nil {
        tracelog.Error(ctx, "failed to publish order event", "error", err)
    }
}
```

---

## Migration Guide

For teams migrating from `quickbill-payment-service` internal packages:

| Old (internal) | New (qbp-lib) |
|-----|-----|
| `pkg/tracing.Start()` | `tracing.Start()` |
| `pkg/tracing.Attr()` | `tracing.Attr()` |
| `pkg/tracing.NewGormTracingPlugin()` | `otelgorm.NewPlugin()` |
| `internal/infrastructure/otel.InitOtel()` | `tracing.Init()` or `tracing.InitFromEnv()` |
| `internal/middleware.TraceMiddleware()` | `otelgin.Middleware()` |
| `internal/middleware.RequestIDMiddleware()` | `otelgin.RequestID()` |
| `pkg/logger.InfoWithTrace()` | `tracelog.Info()` |
| `pkg/logger.ErrorWithTrace()` | `tracelog.Error()` |
| `pkg/ctxpkg.NewBackgroundWithCopyTracer()` | `tracing.DetachedContext()` |
| `pkg/ctxpkg.GetRequestID()` | `tracing.RequestID()` |
| `rabbitmq.ExtractAMQPHeaders()` | `otelamqp.Extract()` |
| `rabbitmq.InjectAMQPHeaders()` | `otelamqp.Inject()` |
| `internal/config.OtelEnv` struct | `tracing.ConfigFromEnv()` |
