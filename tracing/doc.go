package tracing

import ()

/*
Package tracing provides OpenTelemetry distributed tracing utilities for QuickBill services.

This package offers a simple, production-ready API for instrumenting Go applications
with distributed tracing. It follows Go best practices from Effective Go, Uber Go Style
Guide, and the Go Package Names blog.

# Quick Start

	func main() {
		shutdown, err := tracing.InitFromEnv()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown(context.Background())

		// Your application code...
	}

# Core API

Use Start for automatic span creation:

	func (s *service) Process(ctx context.Context, id string) (err error) {
		ctx, finish := tracing.Start(ctx, tracing.Attr("id", id))
		defer finish(&err)

		// Your code here...
		return nil
	}

Use Wrap for functional-style tracing:

	func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
		return tracing.Wrap(ctx, func(ctx context.Context) (*User, error) {
			return s.repo.FindByID(ctx, id)
		}, tracing.Attr("user.id", id))
	}

# Integrations

For Gin middleware, see the otelgin subpackage.
For GORM tracing, see the otelgorm subpackage.
For AMQP propagation, see the otelamqp subpackage.
For trace-aware logging, see the tracelog subpackage.

# Configuration

Environment variables:

	OTEL_ENABLED         - "true"/"false" (default: false)
	OTEL_ENDPOINT        - collector address (default: "localhost:4317")
	OTEL_SERVICE_NAME    - service name (default: binary name)
	OTEL_SERVICE_VERSION - version string (default: "0.0.0")
	OTEL_SAMPLING_RATE   - 0.0-1.0 (default: 1.0)
	OTEL_EXPORTER_TYPE   - "grpc"/"http"/"stdout" (default: "grpc")
	OTEL_INSECURE        - "true"/"false" (default: true)
	ENVIRONMENT          - "local"/"staging"/"production"

# Best Practices

This package follows Go best practices:

  - No init() functions (explicit initialization)
  - No mutable globals (struct-based configuration)
  - Functional options pattern for extensibility
  - Clean package naming (otelgin, otelgorm, etc.)
  - Graceful shutdown handling
  - Zero-cost when disabled

See the README for detailed documentation and migration guides.
*/
