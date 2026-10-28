package json5

import (
	"reflect"
	"testing"
)

// TestStringEscapes checks JSON5 escapes the testdata ES5 oracle cannot match.
func TestStringEscapes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// U+2028 after a backslash is a line continuation.
		// The ES5 fixture oracle keeps the separator.
		{name: "line separator", in: "\"a\\\u2028b\"", want: "ab"},
		// U+2029 after a backslash is a line continuation.
		// The ES5 fixture oracle keeps the separator.
		{name: "paragraph separator", in: "\"a\\\u2029b\"", want: "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			if err := Unmarshal([]byte(tt.in), &got); err != nil {
				t.Fatalf("Unmarshal(%q) unexpected error: %v", tt.in, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Unmarshal(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
