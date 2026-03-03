# tracing

OpenTelemetry distributed tracing with Gin, GORM, and AMQP integrations.

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/QuickBill-Engineering/qbp-lib/tracing"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"
    "github.com/gin-gonic/gin"
)

func main() {
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
        return &User{ID: id, Name: "John Doe"}, nil
    }, tracing.Attr("user.id", id))
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENABLED` | `false` | Enable/disable tracing |
| `OTEL_ENDPOINT` | `localhost:4317` | OpenTelemetry collector address |
| `OTEL_SERVICE_NAME` | binary name | Service name for traces |
| `OTEL_SERVICE_VERSION` | `0.0.0` | Service version |
| `OTEL_SAMPLING_RATE` | `1.0` | Sampling rate (0.0-1.0) |
| `OTEL_EXPORTER_TYPE` | `grpc` | Exporter type: `grpc`, `http`, or `stdout` |
| `OTEL_INSECURE` | `true` | Use insecure connection |
| `ENVIRONMENT` | `local` | Deployment environment |

### Programmatic Configuration

```go
shutdown, err := tracing.Init(
    tracing.WithEnabled(true),
    tracing.WithServiceName("payment-service"),
    tracing.WithEndpoint("otel-collector:4317"),
    tracing.WithSamplingRate(0.5),
)
if err != nil {
    log.Fatal(err)
}
defer shutdown(context.Background())
```

## Core API

### Start - Manual Span Creation

```go
func (s *service) Process(ctx context.Context, id string) (err error) {
    ctx, finish := tracing.Start(ctx, tracing.Attr("id", id))
    defer finish(&err)

    return s.doWork(ctx)
}
```

### Wrap - Functional Style

```go
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
    return tracing.Wrap(ctx, func(ctx context.Context) (*User, error) {
        return s.repo.FindByID(ctx, id)
    }, tracing.Attr("user.id", id))
}
```

### WrapVoid - Error-Only Functions

```go
func (s *service) DeleteUser(ctx context.Context, id string) error {
    return tracing.WrapVoid(ctx, func(ctx context.Context) error {
        return s.repo.Delete(ctx, id)
    }, tracing.Attr("user.id", id))
}
```

### Attr - Type-Safe Attributes

```go
tracing.Attr("user.id", "123")        // string
tracing.Attr("age", 30)                // int
tracing.Attr("score", 3.14)            // float64
tracing.Attr("active", true)           // bool
tracing.Attr("tags", []string{"a","b"}) // string slice
```

## Integrations

### Gin Middleware

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"

r := gin.Default()

r.Use(otelgin.RequestID())
r.Use(otelgin.Middleware())

r.Use(otelgin.Middleware(
    otelgin.WithFilter(func(c *gin.Context) bool {
        return c.Request.URL.Path == "/health"
    }),
))
```

### GORM Plugin

```go
import (
    "gorm.io/gorm"
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm"
)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
    log.Fatal(err)
}

err = db.Use(otelgorm.NewPlugin(
    otelgorm.WithDBName("mydb"),
    otelgorm.WithLogQueries(true),
    otelgorm.WithRecordRowsAffected(true),
))
```

### AMQP Propagation

Publisher:

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp"

func publish(ch *amqp.Channel, ctx context.Context, body []byte) error {
    headers := make(amqp.Table)
    otelamqp.Inject(ctx, headers)

    return ch.PublishWithContext(ctx,
        "exchange", "routing-key", false, false,
        amqp.Publishing{
            Headers: headers,
            Body:    body,
        },
    )
}
```

Consumer:

```go
func handleDelivery(d amqp.Delivery) {
    ctx := otelamqp.Extract(context.Background(), d.Headers)

    ctx, finish := tracing.Start(ctx, "process-message")
    defer finish(nil)

    // Process message...
}
```

### Trace-Aware Logging

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog"

func HandleRequest(ctx context.Context) {
    tracelog.Info(ctx, "processing request",
        "user_id", userID,
        "action", "update",
    )

    if err != nil {
        tracelog.Error(ctx, "failed to process", "error", err)
    }
}
```

## Context Utilities

### Request ID

```go
ctx := tracing.WithRequestID(ctx, "req-123")
requestID := tracing.RequestID(ctx)
```

### Detached Context

For goroutines that must outlive the original request:

```go
func (s *service) ProcessAsync(ctx context.Context, job *Job) {
    detached := tracing.DetachedContext(ctx)

    go func() {
        ctx, finish := tracing.Start(detached, "async-job")
        defer finish(nil)

        s.processJob(ctx, job)
    }()
}
```

## Best Practices

### Error Handling

Pass error pointers to `FinishFunc`:

```go
func (s *service) DoWork(ctx context.Context) (err error) {
    ctx, finish := tracing.Start(ctx)
    defer finish(&err)

    return s.work(ctx)
}
```

### Span Names

Use `StartNamed` for custom span names:

```go
ctx, finish := tracing.StartNamed(ctx, "external-api.call")
defer finish(nil)
```

### Client Spans

Use `StartWithKind` for outbound calls:

```go
ctx, finish := tracing.StartWithKind(ctx, trace.SpanKindClient, "http.client")
defer finish(nil)
```

## Migration Guide

From `quickbill-payment-service`:

| Old | New |
|-----|-----|
| `pkg/tracing.Start()` | `tracing.Start()` |
| `pkg/tracing.Attr()` | `tracing.Attr()` |
| `pkg/tracing.NewGormTracingPlugin()` | `otelgorm.NewPlugin()` |
| `internal/infrastructure/otel.InitOtel()` | `tracing.Init()` |
| `internal/middleware.TraceMiddleware()` | `otelgin.Middleware()` |
| `pkg/logger.InfoWithTrace()` | `tracelog.Info()` |
| `pkg/ctxpkg.NewBackgroundWithCopyTracer()` | `tracing.DetachedContext()` |
| `pkg/ctxpkg.GetRequestID()` | `tracing.RequestID()` |
| `rabbitmq.ExtractAMQPHeaders()` | `otelamqp.Extract()` |
| `internal/config.OtelEnv` | `tracing.ConfigFromEnv()` |