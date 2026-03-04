package otelgin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuickBill-Engineering/qbp-lib/tracing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMiddleware_BasicRequest(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer func() { _ = shutdown(nil) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_WithFilter(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer func() { _ = shutdown(nil) }()

	r := gin.New()
	r.Use(Middleware(
		WithFilter(func(c *gin.Context) bool {
			return c.Request.URL.Path == "/health"
		}),
	))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/api", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_WithRoutePattern(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer func() { _ = shutdown(nil) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_ErrorStatus(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer func() { _ = shutdown(nil) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRequestID_ExistingHeader(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		requestID := tracing.RequestID(c.Request.Context())
		if requestID != "existing-request-id" {
			t.Errorf("expected request ID 'existing-request-id', got %s", requestID)
		}
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-request-id")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "existing-request-id" {
		t.Errorf("expected X-Request-ID header 'existing-request-id', got %s", w.Header().Get("X-Request-ID"))
	}
}

func TestRequestID_GenerateNew(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		requestID := tracing.RequestID(c.Request.Context())
		if requestID == "" {
			t.Error("expected non-empty request ID")
		}
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestID_MultipleRequests(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": tracing.RequestID(c.Request.Context())})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	id1 := w1.Header().Get("X-Request-ID")
	id2 := w2.Header().Get("X-Request-ID")

	if id1 == id2 {
		t.Error("expected different request IDs for different requests")
	}
}

func TestMiddleware_SpanNameWithQuery(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/users/123?date=2022-01-03", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "expected one span")

	span := spans[0]
	assert.Equal(t, "/users/123?date=2022-01-03", span.Name, "span name should be request URI with query string")
}

func TestMiddleware_SpanAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/company/:id/balance", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/company/456/balance?date=2022-01-03", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "expected one span")

	span := spans[0]

	attrs := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrs[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, "GET", attrs["http.request.method"])
	assert.Equal(t, "/company/456/balance", attrs["url.path"])
	assert.Equal(t, "date=2022-01-03", attrs["url.query"])
	assert.Equal(t, "/company/:id/balance", attrs["http.route"])
	assert.Equal(t, int64(200), attrs["http.response.status_code"])
}

func TestMiddleware_SpanNameWithoutQuery(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "expected one span")

	span := spans[0]
	assert.Equal(t, "/health", span.Name, "span name should be path without query when no query present")
}

func TestMiddleware_UsesGlobalTracerProvider(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1, "middleware should use global TracerProvider and create spans")
}
