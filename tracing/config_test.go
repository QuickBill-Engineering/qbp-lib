package tracing

import (
	"os"
	"testing"
)

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:     "default values",
			envVars:  map[string]string{},
			expected: Config{},
		},
		{
			name: "enabled",
			envVars: map[string]string{
				"OTEL_ENABLED": "true",
			},
			expected: Config{enabled: true},
		},
		{
			name: "custom endpoint",
			envVars: map[string]string{
				"OTEL_ENDPOINT": "collector:4317",
			},
			expected: Config{endpoint: "collector:4317"},
		},
		{
			name: "service name and version",
			envVars: map[string]string{
				"OTEL_SERVICE_NAME":    "test-service",
				"OTEL_SERVICE_VERSION": "1.0.0",
			},
			expected: Config{
				serviceName:    "test-service",
				serviceVersion: "1.0.0",
			},
		},
		{
			name: "sampling rate",
			envVars: map[string]string{
				"OTEL_SAMPLING_RATE": "0.5",
			},
			expected: Config{samplingRate: 0.5},
		},
		{
			name: "exporter type http",
			envVars: map[string]string{
				"OTEL_EXPORTER_TYPE": "http",
			},
			expected: Config{exporter: ExporterHTTP},
		},
		{
			name: "exporter type stdout",
			envVars: map[string]string{
				"OTEL_EXPORTER_TYPE": "stdout",
			},
			expected: Config{exporter: ExporterStdout},
		},
		{
			name: "insecure false",
			envVars: map[string]string{
				"OTEL_INSECURE": "false",
			},
			expected: Config{insecure: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			opts := ConfigFromEnv()
			cfg := newConfig(opts...)

			switch tt.name {
			case "enabled":
				if cfg.enabled != tt.expected.enabled {
					t.Errorf("expected enabled %v, got %v", tt.expected.enabled, cfg.enabled)
				}
			case "custom endpoint":
				if cfg.endpoint != tt.expected.endpoint {
					t.Errorf("expected endpoint %v, got %v", tt.expected.endpoint, cfg.endpoint)
				}
			case "service name and version":
				if cfg.serviceName != tt.expected.serviceName {
					t.Errorf("expected serviceName %v, got %v", tt.expected.serviceName, cfg.serviceName)
				}
				if cfg.serviceVersion != tt.expected.serviceVersion {
					t.Errorf("expected serviceVersion %v, got %v", tt.expected.serviceVersion, cfg.serviceVersion)
				}
			case "sampling rate":
				if cfg.samplingRate != tt.expected.samplingRate {
					t.Errorf("expected samplingRate %v, got %v", tt.expected.samplingRate, cfg.samplingRate)
				}
			case "exporter type http", "exporter type stdout":
				if cfg.exporter != tt.expected.exporter {
					t.Errorf("expected exporter %v, got %v", tt.expected.exporter, cfg.exporter)
				}
			case "insecure false":
				if cfg.insecure != tt.expected.insecure {
					t.Errorf("expected insecure %v, got %v", tt.expected.insecure, cfg.insecure)
				}
			}
		})
	}
}

func TestWithEnabled(t *testing.T) {
	cfg := newConfig(WithEnabled(true))
	if !cfg.enabled {
		t.Error("expected enabled to be true")
	}
}

func TestWithEndpoint(t *testing.T) {
	cfg := newConfig(WithEndpoint("localhost:4318"))
	if cfg.endpoint != "localhost:4318" {
		t.Errorf("expected endpoint localhost:4318, got %v", cfg.endpoint)
	}
}

func TestWithServiceName(t *testing.T) {
	cfg := newConfig(WithServiceName("my-service"))
	if cfg.serviceName != "my-service" {
		t.Errorf("expected serviceName my-service, got %v", cfg.serviceName)
	}
}

func TestWithSamplingRate(t *testing.T) {
	cfg := newConfig(WithSamplingRate(0.25))
	if cfg.samplingRate != 0.25 {
		t.Errorf("expected samplingRate 0.25, got %v", cfg.samplingRate)
	}
}

func TestWithExporter(t *testing.T) {
	cfg := newConfig(WithExporter(ExporterHTTP))
	if cfg.exporter != ExporterHTTP {
		t.Errorf("expected exporter ExporterHTTP, got %v", cfg.exporter)
	}
}
