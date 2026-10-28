package json5

import (
	"reflect"
	"testing"
)

// TestBlockCommentTerminator checks that a block comment ends only at "*/".
func TestBlockCommentTerminator(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want interface{}
		fail bool
	}{
		// "* /" does not end the comment; only the later "*/" does.
		{name: "star space slash", in: "/* * / */ 1", want: float64(1)},
		// A "/" after "*" is comment text when other bytes sit between them.
		{name: "star then slash", in: "/* a * b / c */ 1", want: float64(1)},
		// Asterisks may surround the body, including immediately before "*/".
		{name: "asterisks around body", in: "/** x **/ true", want: true},
		// "*" immediately before "*/", with no other body.
		{name: "star before close", in: "/***/ 1", want: float64(1)},
		// "/**/" is an empty comment: "*" then "/".
		{name: "empty", in: "/**/ null", want: nil},
		// A slash before the closing asterisks is still inside the comment.
		{name: "slash then stars", in: "/* / **/ 1", want: float64(1)},
		// The same sequence inside an array still yields both elements.
		{name: "inside array", in: "[1, /* * / */ 2]", want: []interface{}{float64(1), float64(2)}},
		// No "*/" at all: the value is inside the comment and input ends.
		{name: "unterminated", in: "/* * / 1", fail: true},
		// "*" at the end of input is not a terminator.
		{name: "unterminated star", in: "/***", fail: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			err := Unmarshal([]byte(tt.in), &got)
			if tt.fail {
				if _, ok := err.(*SyntaxError); !ok {
					t.Fatalf("Unmarshal(%q) error = %v, want syntax error", tt.in, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unmarshal(%q) unexpected error: %v", tt.in, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Unmarshal(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
