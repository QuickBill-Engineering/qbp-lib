package tracing

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

// Attr creates an OpenTelemetry attribute with automatic type detection.
// This helper simplifies attribute creation by inferring the type from the value.
//
// Supported types:
//   - string → attribute.String
//   - int → attribute.Int
//   - int64 → attribute.Int64
//   - float64 → attribute.Float64
//   - bool → attribute.Bool
//   - []string → attribute.StringSlice
//   - []int → attribute.IntSlice
//   - []int64 → attribute.Int64Slice
//   - []float64 → attribute.Float64Slice
//   - []bool → attribute.BoolSlice
//   - fmt.Stringer → attribute.String (calls String() method)
//   - any other type → attribute.String (uses fmt.Sprintf("%v", value))
//
// Parameters:
//   - key: The attribute key. Use dot-notation for namespacing (e.g., "user.id", "http.status_code").
//   - value: The attribute value. Type is automatically detected.
//
// Returns:
//   - attribute.KeyValue: A typed key-value pair for use with OpenTelemetry.
//
// Example:
//
//	// Simple types
//	tracing.Attr("user.id", "123")
//	tracing.Attr("age", 30)
//	tracing.Attr("score", 3.14)
//	tracing.Attr("active", true)
//
//	// Slices
//	tracing.Attr("tags", []string{"premium", "verified"})
//
//	// Stringer interface
//	tracing.Attr("user", user) // calls user.String()
//
//	// Any other type
//	tracing.Attr("metadata", complexStruct) // uses fmt.Sprintf
func Attr(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int:
		return attribute.IntSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	case fmt.Stringer:
		return attribute.String(key, v.String())
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}
