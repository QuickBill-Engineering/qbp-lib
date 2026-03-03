package otelamqp

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
)

// HeadersCarrier implements propagation.TextMapCarrier over amqp.Table,
// enabling W3C trace context propagation through RabbitMQ message headers.
// This allows traces to flow across service boundaries when using message queues.
//
// HeadersCarrier is used internally by Inject and Extract functions.
type HeadersCarrier amqp.Table

var _ propagation.TextMapCarrier = (HeadersCarrier)(nil)

// Get retrieves a string value from the AMQP headers by key.
// Returns an empty string if the key doesn't exist or the value is not a string.
func (c HeadersCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Set stores a string value in the AMQP headers.
func (c HeadersCarrier) Set(key string, value string) {
	c[key] = value
}

// Keys returns all header keys present in the carrier.
func (c HeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// Inject injects the trace context from ctx into AMQP headers.
// Call this before publishing a message to propagate the trace to the consumer.
//
// The function injects W3C trace context headers:
//   - traceparent: Contains trace ID, span ID, and trace flags
//   - tracestate: Contains vendor-specific trace information
//
// Parameters:
//   - ctx: The context containing the current span.
//   - headers: The AMQP headers table to inject into (can be nil, will be created).
//
// Example (Publisher):
//
//	func publish(ch *amqp.Channel, ctx context.Context, body []byte) error {
//	    headers := make(amqp.Table)
//	    otelamqp.Inject(ctx, headers)
//
//	    return ch.PublishWithContext(ctx,
//	        "exchange", "routing-key", false, false,
//	        amqp.Publishing{
//	            Headers: headers,
//	            Body:    body,
//	        },
//	    )
//	}
func Inject(ctx context.Context, headers amqp.Table) {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	propagator.Inject(ctx, HeadersCarrier(headers))
}

// Extract returns a new context with trace context extracted from AMQP headers.
// Call this when consuming a message to continue the trace from the producer.
//
// The function extracts W3C trace context headers and creates a new context
// that contains the span context, allowing the consumer to create child spans.
//
// Parameters:
//   - ctx: The base context (typically context.Background() or context.TODO()).
//   - headers: The AMQP headers table from the received message.
//
// Returns:
//   - context.Context: A context containing the extracted trace context.
//
// Example (Consumer):
//
//	func handleDelivery(d amqp.Delivery) {
//	    ctx := otelamqp.Extract(context.Background(), d.Headers)
//
//	    ctx, finish := tracing.Start(ctx, "process-message")
//	    defer finish(nil)
//
//	    // Process message with trace context...
//	}
func Extract(ctx context.Context, headers amqp.Table) context.Context {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return propagator.Extract(ctx, HeadersCarrier(headers))
}
