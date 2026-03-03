package tracing

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestAttr(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected attribute.KeyValue
	}{
		{
			name:     "string",
			key:      "key",
			value:    "value",
			expected: attribute.String("key", "value"),
		},
		{
			name:     "int",
			key:      "key",
			value:    42,
			expected: attribute.Int("key", 42),
		},
		{
			name:     "int64",
			key:      "key",
			value:    int64(42),
			expected: attribute.Int64("key", 42),
		},
		{
			name:     "float64",
			key:      "key",
			value:    3.14,
			expected: attribute.Float64("key", 3.14),
		},
		{
			name:     "bool",
			key:      "key",
			value:    true,
			expected: attribute.Bool("key", true),
		},
		{
			name:     "string slice",
			key:      "key",
			value:    []string{"a", "b", "c"},
			expected: attribute.StringSlice("key", []string{"a", "b", "c"}),
		},
		{
			name:     "int slice",
			key:      "key",
			value:    []int{1, 2, 3},
			expected: attribute.IntSlice("key", []int{1, 2, 3}),
		},
		{
			name:     "bool slice",
			key:      "key",
			value:    []bool{true, false},
			expected: attribute.BoolSlice("key", []bool{true, false}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Attr(tt.key, tt.value)
			if result.Key != tt.expected.Key {
				t.Errorf("expected key %v, got %v", tt.expected.Key, result.Key)
			}
		})
	}
}

func TestAttrStringer(t *testing.T) {
	s := &stringer{value: "test"}
	result := Attr("key", s)
	if result.Key != "key" {
		t.Errorf("expected key 'key', got %v", result.Key)
	}
}

func TestAttrFallback(t *testing.T) {
	result := Attr("key", struct{ name string }{name: "test"})
	if result.Key != "key" {
		t.Errorf("expected key 'key', got %v", result.Key)
	}
}

type stringer struct {
	value string
}

func (s *stringer) String() string {
	return s.value
}
