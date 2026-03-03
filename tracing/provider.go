package tracing

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	// globalTracer stores a trace.Tracer; accessed atomically.
	globalTracer atomic.Value
)

// Init creates and registers a global OpenTelemetry TracerProvider.
//
// This function should be called once at application startup, typically in main().
// It returns a shutdown function that must be called before the application exits
// to flush pending spans and release resources. Use defer to ensure cleanup.
//
// When tracing is disabled (WithEnabled(false) or OTEL_ENABLED=false), a no-op
// provider is installed globally and all tracing calls become zero-cost.
// The shutdown function is safe to call regardless of whether tracing is enabled.
//
// Parameters:
//   - opts: Functional options to configure the tracer (e.g., WithServiceName, WithEndpoint)
//
// Returns:
//   - shutdown: A function that flushes and closes the tracer provider. Always call this.
//   - err: Non-nil if the tracer could not be initialized.
//
// Example:
//
//	func main() {
//	    shutdown, err := tracing.Init(
//	        tracing.WithEnabled(true),
//	        tracing.WithServiceName("payment-service"),
//	        tracing.WithEndpoint("otel-collector:4317"),
//	    )
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer shutdown(context.Background())
//
//	    // Your application code...
//	}
func Init(opts ...Option) (shutdown func(context.Context) error, err error) {
	cfg := newConfig(opts...)

	if !cfg.enabled {
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)
		globalTracer.Store(tp.Tracer("qbp-lib"))
		return func(ctx context.Context) error { return nil }, nil
	}

	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("create exporter: %w", err)
	}

	res, err := createResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.samplingRate)),
	)

	otel.SetTracerProvider(tp)
	globalTracer.Store(tp.Tracer(cfg.serviceName))

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

// InitFromEnv is a convenience function that reads OTEL_* environment variables
// and calls Init with the resulting configuration.
//
// This is equivalent to calling:
//
//	shutdown, err := tracing.Init(tracing.ConfigFromEnv()...)
//
// See ConfigFromEnv for the list of supported environment variables.
//
// Example:
//
//	func main() {
//	    shutdown, err := tracing.InitFromEnv()
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer shutdown(context.Background())
//
//	    // Your application code...
//	}
func InitFromEnv() (func(context.Context) error, error) {
	return Init(ConfigFromEnv()...)
}

func createExporter(cfg *Config) (sdktrace.SpanExporter, error) {
	switch cfg.exporter {
	case ExporterGRPC:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.endpoint),
		}
		if cfg.insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.New(context.Background(), opts...)
	case ExporterHTTP:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.endpoint),
		}
		if cfg.insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(context.Background(), opts...)
	case ExporterStdout:
		return stdouttrace.New()
	default:
		return nil, fmt.Errorf("unknown exporter type: %v", cfg.exporter)
	}
}

func createResource(cfg *Config) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.serviceName),
			semconv.ServiceVersion(cfg.serviceVersion),
			attribute.String("deployment.environment", cfg.environment),
			attribute.String("library", "qbp-lib"),
		),
	)
}

// getTracer returns the globally configured tracer, or nil if Init has not been called.
func getTracer() trace.Tracer {
	t, _ := globalTracer.Load().(trace.Tracer)
	return t
}
