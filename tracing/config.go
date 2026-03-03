package tracing

import (
	"os"
	"strconv"
)

// Config holds OpenTelemetry tracer provider configuration.
// Use functional options or ConfigFromEnv to construct.
// Fields are unexported to prevent construction with invalid values.
type Config struct {
	enabled        bool
	endpoint       string
	serviceName    string
	serviceVersion string
	environment    string
	samplingRate   float64
	exporter       ExporterType
	insecure       bool
}

// ExporterType determines the OTLP transport protocol for exporting traces.
type ExporterType int

const (
	// ExporterGRPC uses gRPC to send traces to the collector.
	// This is the recommended option for production use.
	ExporterGRPC ExporterType = iota + 1

	// ExporterHTTP uses HTTP/Protobuf to send traces to the collector.
	// Useful when gRPC is not available or when behind an HTTP load balancer.
	ExporterHTTP

	// ExporterStdout writes traces to stdout in JSON format.
	// Useful for local development and debugging.
	ExporterStdout
)

// Option configures the tracing provider using the functional options pattern.
// Options can be passed to Init() to customize the tracer configuration.
type Option func(*Config)

// WithEnabled enables or disables tracing.
// When disabled (default), a no-op provider is installed and all tracing
// calls become zero-cost. The Init() function still succeeds and returns
// a valid shutdown function.
//
// Example:
//
//	tracing.Init(tracing.WithEnabled(true))
func WithEnabled(v bool) Option {
	return func(c *Config) { c.enabled = v }
}

// WithEndpoint sets the OpenTelemetry collector endpoint address.
// Format: "host:port" (e.g., "localhost:4317" for gRPC, "localhost:4318" for HTTP).
// Default: "localhost:4317"
//
// Example:
//
//	tracing.Init(tracing.WithEndpoint("otel-collector:4317"))
func WithEndpoint(v string) Option {
	return func(c *Config) { c.endpoint = v }
}

// WithServiceName sets the service name for trace attribution.
// This appears in tracing backends (Jaeger, Tempo, etc.) and helps
// identify the source of spans.
// Default: binary name from os.Args[0]
//
// Example:
//
//	tracing.Init(tracing.WithServiceName("payment-service"))
func WithServiceName(v string) Option {
	return func(c *Config) { c.serviceName = v }
}

// WithServiceVersion sets the service version for trace attribution.
// Useful for tracking which version of the service produced a trace.
// Default: "0.0.0"
//
// Example:
//
//	tracing.Init(tracing.WithServiceVersion("1.2.3"))
func WithServiceVersion(v string) Option {
	return func(c *Config) { c.serviceVersion = v }
}

// WithEnvironment sets the deployment environment.
// Common values: "local", "development", "staging", "production".
// Default: value of ENVIRONMENT env var, or "local"
//
// Example:
//
//	tracing.Init(tracing.WithEnvironment("production"))
func WithEnvironment(v string) Option {
	return func(c *Config) { c.environment = v }
}

// WithSamplingRate sets the trace sampling rate.
// Value must be between 0.0 (no traces) and 1.0 (all traces).
// Lower rates reduce storage costs but may miss some traces.
// Default: 1.0 (100% sampling)
//
// Example:
//
//	tracing.Init(tracing.WithSamplingRate(0.1)) // 10% of traces
func WithSamplingRate(v float64) Option {
	return func(c *Config) {
		switch {
		case v < 0.0:
			c.samplingRate = 0.0
		case v > 1.0:
			c.samplingRate = 1.0
		default:
			c.samplingRate = v
		}
	}
}

// WithExporter sets the exporter type for sending traces.
// Options: ExporterGRPC, ExporterHTTP, or ExporterStdout.
// Default: ExporterGRPC
//
// Example:
//
//	tracing.Init(tracing.WithExporter(tracing.ExporterHTTP))
func WithExporter(v ExporterType) Option {
	return func(c *Config) { c.exporter = v }
}

// WithInsecure sets whether to use an insecure connection to the collector.
// When true, TLS is disabled. Should be false in production with TLS enabled.
// Default: true
//
// Example:
//
//	tracing.Init(tracing.WithInsecure(false)) // Enable TLS
func WithInsecure(v bool) Option {
	return func(c *Config) { c.insecure = v }
}

func newConfig(opts ...Option) *Config {
	c := &Config{
		enabled:        false,
		endpoint:       "localhost:4317",
		serviceName:    getBinaryName(),
		serviceVersion: "0.0.0",
		environment:    getEnvWithDefault("ENVIRONMENT", "local"),
		samplingRate:   1.0,
		exporter:       ExporterGRPC,
		insecure:       true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ConfigFromEnv reads configuration from environment variables and returns
// a slice of Option functions that can be passed to Init().
//
// Environment variables:
//
//	OTEL_ENABLED          - "true"/"false" (default: false)
//	OTEL_ENDPOINT         - collector address (default: "localhost:4317")
//	OTEL_SERVICE_NAME     - resource name (default: binary name from os.Args[0])
//	OTEL_SERVICE_VERSION  - version string (default: "0.0.0")
//	OTEL_SAMPLING_RATE    - 0.0-1.0 (default: 1.0)
//	OTEL_EXPORTER_TYPE    - "grpc"/"http"/"stdout" (default: "grpc")
//	OTEL_INSECURE         - "true"/"false" (default: true)
//	ENVIRONMENT           - "local"/"staging"/"production"
//
// Example:
//
//	shutdown, err := tracing.Init(tracing.ConfigFromEnv()...)
func ConfigFromEnv() []Option {
	var opts []Option

	if v := os.Getenv("OTEL_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			opts = append(opts, WithEnabled(enabled))
		}
	}

	if v := os.Getenv("OTEL_ENDPOINT"); v != "" {
		opts = append(opts, WithEndpoint(v))
	}

	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		opts = append(opts, WithServiceName(v))
	}

	if v := os.Getenv("OTEL_SERVICE_VERSION"); v != "" {
		opts = append(opts, WithServiceVersion(v))
	}

	if v := os.Getenv("ENVIRONMENT"); v != "" {
		opts = append(opts, WithEnvironment(v))
	}

	if v := os.Getenv("OTEL_SAMPLING_RATE"); v != "" {
		if rate, err := strconv.ParseFloat(v, 64); err == nil {
			opts = append(opts, WithSamplingRate(rate))
		}
	}

	if v := os.Getenv("OTEL_EXPORTER_TYPE"); v != "" {
		switch v {
		case "grpc":
			opts = append(opts, WithExporter(ExporterGRPC))
		case "http":
			opts = append(opts, WithExporter(ExporterHTTP))
		case "stdout":
			opts = append(opts, WithExporter(ExporterStdout))
		}
	}

	if v := os.Getenv("OTEL_INSECURE"); v != "" {
		if insecure, err := strconv.ParseBool(v); err == nil {
			opts = append(opts, WithInsecure(insecure))
		}
	}

	return opts
}

func getBinaryName() string {
	if len(os.Args) > 0 {
		return os.Args[0]
	}
	return "unknown"
}

func getEnvWithDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
