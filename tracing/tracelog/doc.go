package tracelog

import ()

/*
Package tracelog provides trace-aware structured logging with slog.

This package automatically extracts trace_id, span_id, and request_id from
context and includes them in log entries.

Example:

	import (
		"github.com/QuickBill-Engineering/qbp-lib/tracing/tracelog"
	)

	func HandleRequest(ctx context.Context) {
		tracelog.Info(ctx, "processing request",
			"user_id", userID,
			"action", "update",
		)

		if err != nil {
			tracelog.Error(ctx, "failed to process", "error", err)
		}
	}
*/
