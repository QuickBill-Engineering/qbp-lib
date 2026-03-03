# qbp-lib

Shared Go libraries for QuickBill Engineering services.

## Packages

| Package | Import Path | Description |
|---------|------------|-------------|
| [tracing](./tracing/) | `github.com/QuickBill-Engineering/qbp-lib/tracing` | OpenTelemetry tracer initialization, span creation, and context utilities |
| [otelgin](./tracing/otelgin/) | `github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin` | Gin HTTP middleware for automatic request tracing and request ID management |
| [otelgorm](./tracing/otelgorm/) | `github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm` | GORM plugin for automatic database query tracing |
| [otelamqp](./tracing/otelamqp/) | `github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp` | RabbitMQ/AMQP trace context propagation (inject/extract) |
| [tracelog](./tracing/tracelog/) | `github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog` | Structured logging with automatic trace correlation (trace_id, span_id, request_id) |

## Requirements

- Go 1.25 or later

## Installation

```bash
go get github.com/QuickBill-Engineering/qbp-lib
```

## Quick Start

### 1. Initialize tracing at startup

```go
package main

import (
    "context"
    "log"

    "github.com/QuickBill-Engineering/qbp-lib/tracing"
)

func main() {
    // Option A: Read config from environment variables
    shutdown, err := tracing.InitFromEnv()

    // Option B: Programmatic configuration
    shutdown, err := tracing.Init(
        tracing.WithEnabled(true),
        tracing.WithServiceName("payment-service"),
        tracing.WithEndpoint("otel-collector:4317"),
        tracing.WithSamplingRate(1.0),
        tracing.WithEnvironment("production"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(context.Background())
}
```

### 2. Add Gin middleware

```go
import (
    "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"
    "github.com/gin-gonic/gin"
)

r := gin.Default()
r.Use(otelgin.RequestID())  // Adds X-Request-ID to every request
r.Use(otelgin.Middleware())  // Creates server spans for every request
```

### 3. Trace your functions

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing"

// Option A: Start/Finish pattern (recommended for most cases)
func (s *service) Process(ctx context.Context, id string) (err error) {
    ctx, finish := tracing.Start(ctx, tracing.Attr("id", id))
    defer finish(&err) // automatically records error status on the span

    return s.doWork(ctx)
}

// Option B: Wrap pattern (functional style, less boilerplate)
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
    return tracing.Wrap(ctx, func(ctx context.Context) (*User, error) {
        return s.repo.FindByID(ctx, id)
    }, tracing.Attr("user.id", id))
}
```

### 4. Add GORM tracing

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm"

db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
db.Use(otelgorm.NewPlugin(
    otelgorm.WithDBName("payments"),
    otelgorm.WithRecordRowsAffected(true),
))
```

### 5. Propagate traces through RabbitMQ

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp"

// Publisher: inject trace context into message headers
headers := make(amqp.Table)
otelamqp.Inject(ctx, headers)

// Consumer: extract trace context from message headers
ctx := otelamqp.Extract(context.Background(), delivery.Headers)
```

### 6. Log with trace correlation

```go
import "github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog"

tracelog.Info(ctx, "payment processed", "amount", 100, "currency", "IDR")
// Output: {"level":"INFO","trace_id":"abc...","span_id":"def...","request_id":"req-789","msg":"payment processed","amount":100,"currency":"IDR"}
```

## Full Documentation

See the **[tracing package documentation](./tracing/README.md)** for the complete API reference, all configuration options, integration guides, and best practices.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENABLED` | `false` | Enable/disable tracing |
| `OTEL_ENDPOINT` | `localhost:4317` | OpenTelemetry collector address (host:port) |
| `OTEL_SERVICE_NAME` | binary name | Service name for trace attribution |
| `OTEL_SERVICE_VERSION` | `0.0.0` | Service version |
| `OTEL_SAMPLING_RATE` | `1.0` | Sampling rate (0.0 to 1.0) |
| `OTEL_EXPORTER_TYPE` | `grpc` | Transport: `grpc`, `http`, or `stdout` |
| `OTEL_INSECURE` | `true` | Disable TLS (set `false` in production) |
| `ENVIRONMENT` | `local` | Deployment environment name |

## Development

```bash
# Run all tests
go test ./... -race

# Run vet
go vet ./...

# Check formatting
gofmt -l .
```

## Design Principles

- **No `init()` functions** -- explicit initialization via `tracing.Init()` required
- **No mutable globals** -- thread-safe via `sync/atomic`; struct-based configuration
- **Functional options pattern** -- extensible without breaking changes
- **Zero-cost when disabled** -- no-op provider installed when `WithEnabled(false)`
- **Graceful shutdown** -- `Init()` returns a shutdown function; always `defer` it
- **Context propagation** -- all trace state flows through `context.Context`

## License

MIT License - see [LICENSE](./LICENSE) for details.
