package otelamqp

import ()

/*
Package otelamqp provides OpenTelemetry trace context propagation for AMQP.

This package enables W3C trace context propagation through RabbitMQ message headers.

Example (Publisher):

	import (
		"github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp"
		amqp "github.com/rabbitmq/amqp091-go"
	)

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

Example (Consumer):

	import (
		"github.com/QuickBill-Engineering/qbp-lib/tracing/otelamqp"
	)

	func handleDelivery(d amqp.Delivery) {
		ctx := otelamqp.Extract(context.Background(), d.Headers)
		// Process message with trace context...
	}
*/
