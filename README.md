# qbp-lib

Shared internal Go libraries for QuickBill Engineering services.

## Packages

| Package | Description |
|---------|-------------|
| [tracing](./tracing/) | OpenTelemetry distributed tracing with Gin, GORM, and AMQP integrations |

## Installation

Configure Go to access private repositories:

```bash
go env -w GOPRIVATE=github.com/QuickBill-Engineering/*
```

Install the library:

```bash
go get github.com/QuickBill-Engineering/qbp-lib
```

## Usage

### Tracing

Import the tracing package and initialize it at application startup:

```go
package main

import (
    "context"
    "log"

    "github.com/QuickBill-Engineering/qbp-lib/tracing"
)

func main() {
    shutdown, err := tracing.InitFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(context.Background())

    // Your application code...
}
```

See the [tracing package documentation](./tracing/) for detailed usage.

## Development

### Prerequisites

- Go 1.25 or later

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build ./...
```

## Contributing

This library follows Go best practices from:

- [Effective Go](https://golang.org/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Package Names](https://blog.golang.org/package-names)

Key principles:

- No `init()` functions - explicit initialization required
- No mutable globals - use struct-based configuration
- Functional options pattern for extensibility
- Clean package naming to avoid collisions
- Graceful shutdown handling
- Zero-cost when tracing is disabled

## License

Proprietary - QuickBill Engineering