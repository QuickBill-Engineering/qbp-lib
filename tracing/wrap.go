package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// Wrap executes fn within a new span named after the calling function.
// The span is automatically ended with error status if fn returns an error.
// This is a functional-style alternative to using Start/Finish manually.
//
// The span name is derived from the calling function, not the closure passed to fn.
// For example, if Wrap is called inside GetUser, the span will be named "service.UserService.GetUser".
//
// Type Parameters:
//   - T: The return type of fn.
//
// Parameters:
//   - ctx: The parent context.
//   - fn: The function to execute within the span. Receives a context with the new span.
//   - attrs: Optional attributes to attach to the span.
//
// Returns:
//   - T: The result from fn.
//   - error: The error from fn, if any.
//
// Example:
//
//	func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
//	    return tracing.Wrap(ctx, func(ctx context.Context) (*User, error) {
//	        user, err := s.repo.FindByID(ctx, id)
//	        if err != nil {
//	            return nil, fmt.Errorf("find user: %w", err)
//	        }
//	        return user, nil
//	    }, tracing.Attr("user.id", id))
//	}
func Wrap[T any](ctx context.Context, fn func(context.Context) (T, error), attrs ...attribute.KeyValue) (T, error) {
	return WrapNamed(ctx, getCallerFunctionName(), fn, attrs...)
}

// WrapVoid executes fn within a new span for functions that return only an error.
// This is a convenience wrapper around Wrap for void functions.
//
// The span is automatically ended with error status if fn returns an error.
//
// Parameters:
//   - ctx: The parent context.
//   - fn: The function to execute within the span. Receives a context with the new span.
//   - attrs: Optional attributes to attach to the span.
//
// Returns:
//   - error: The error from fn, if any.
//
// Example:
//
//	func (s *service) DeleteUser(ctx context.Context, id string) error {
//	    return tracing.WrapVoid(ctx, func(ctx context.Context) error {
//	        return s.repo.Delete(ctx, id)
//	    }, tracing.Attr("user.id", id))
//	}
func WrapVoid(ctx context.Context, fn func(context.Context) error, attrs ...attribute.KeyValue) error {
	_, err := Wrap(ctx, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	}, attrs...)
	return err
}

// WrapNamed executes fn within a new span with an explicit name.
// Use this when you need control over the span name.
//
// Type Parameters:
//   - T: The return type of fn.
//
// Parameters:
//   - ctx: The parent context.
//   - name: The span name.
//   - fn: The function to execute within the span. Receives a context with the new span.
//   - attrs: Optional attributes to attach to the span.
//
// Returns:
//   - T: The result from fn.
//   - error: The error from fn, if any.
//
// Example:
//
//	func (s *service) CallExternalAPI(ctx context.Context) (*Response, error) {
//	    return tracing.WrapNamed(ctx, "external-api.call", func(ctx context.Context) (*Response, error) {
//	        return s.client.Get(ctx, "/api/data")
//	    })
//	}
func WrapNamed[T any](ctx context.Context, name string, fn func(context.Context) (T, error), attrs ...attribute.KeyValue) (T, error) {
	ctx, finish := StartNamed(ctx, name, attrs...)
	var zero T

	result, err := fn(ctx)
	if err != nil {
		finish(&err)
		return zero, err
	}

	finish(nil)
	return result, nil
}
