package json5

import (
	"reflect"
	"strings"
	"testing"
)

// TestStringEscapes checks JSON5 string escapes.
func TestStringEscapes(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   interface{}
		fail   bool
		offset int64
	}{
		// \v is a vertical tab.
		{name: "vertical tab", in: `"\v"`, want: "\v"},
		// \v works in a single-quoted string.
		{name: "vertical tab single", in: `'\v'`, want: "\v"},
		// \0 is NUL when no digit follows.
		{name: "null", in: `"\0"`, want: "\x00"},
		// A non-digit after \0 is not part of the escape.
		{name: "null then letter", in: `"\0a"`, want: "\x00a"},
		// \xHH is the character with that code point.
		{name: "hex", in: `"\x41"`, want: "A"},
		// Hex digits may be uppercase.
		{name: "hex upper", in: `"\x6A"`, want: "j"},
		// \xFF is U+00FF, not a raw byte.
		{name: "hex latin1", in: `"\xFF"`, want: "\u00ff"},
		// \x0A is a newline character, not a line continuation.
		{name: "hex newline", in: `"\x0A"`, want: "\n"},
		// Backslash and a line feed produce nothing.
		{name: "continue lf", in: "\"a\\\nb\"", want: "ab"},
		// Backslash and a carriage return produce nothing.
		{name: "continue cr", in: "\"a\\\rb\"", want: "ab"},
		// CRLF is one line terminator sequence.
		{name: "continue crlf", in: "\"a\\\r\nb\"", want: "ab"},
		// U+2028 after a backslash is a line continuation.
		// The ES5 fixture oracle keeps the separator, so this stays a unit test.
		{name: "continue ls", in: "\"a\\\u2028b\"", want: "ab"},
		// U+2029 after a backslash is a line continuation.
		// The ES5 fixture oracle keeps the separator, so this stays a unit test.
		{name: "continue ps", in: "\"a\\\u2029b\"", want: "ab"},
		// A non-escape character is itself.
		{name: "identity", in: `"\A"`, want: "A"},
		// An escaped apostrophe is an apostrophe.
		{name: "apostrophe", in: `"\'"`, want: "'"},
		// The spec's identity example, including \/.
		{name: "identity word", in: `'\A\C\/\D\C'`, want: "AC/DC"},
		// Uppercase \X is not a hex escape.
		{name: "identity X", in: `"\X41"`, want: "X41"},
		// A non-ASCII character after a backslash is itself.
		{name: "identity unicode", in: "\"\\\u00e9\"", want: "é"},
		// Letter n is a newline. A line continuation is a real line break.
		{name: "newline escape", in: `"a\nb"`, want: "a\nb"},
		// A surrogate pair is one character, and a following escape still works.
		{name: "surrogate pair", in: `"\uD83C\uDFBC\v"`, want: "\U0001F3BC\v"},
		// Lowercase hex digits work in a surrogate pair.
		{name: "surrogate pair lower", in: `"\ud83c\udfbc"`, want: "\U0001F3BC"},
		// A lone surrogate is still the replacement character.
		{name: "lone surrogate", in: `"\uD800"`, want: "\uFFFD"},
		// Escapes in an object key use the same decoder.
		{name: "hex key", in: `{"\x41":1}`, want: map[string]interface{}{"A": float64(1)}},
		// \1 is not an octal escape.
		{name: "digit", in: `"\1"`, fail: true, offset: 3},
		// \0 followed by a digit is not an octal escape.
		{name: "null digit", in: `"\01"`, fail: true, offset: 4},
		// \x needs two hex digits.
		{name: "hex short", in: `"\x"`, fail: true, offset: 4},
		// A non-hex digit does not complete \xHH.
		{name: "hex bad", in: `"\x4G"`, fail: true, offset: 5},
		// \u needs four hex digits.
		{name: "unicode short", in: `"\u12"`, fail: true, offset: 6},
		// A non-hex digit does not complete \uXXXX.
		{name: "unicode bad", in: `"\u12G4"`, fail: true, offset: 6},
		// \u{...} is not a JSON5 escape.
		{name: "unicode braces", in: `"\u{41}"`, fail: true, offset: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkEscape(t, tt.in, tt.want, tt.fail, tt.offset)
		})
	}

	// \1 through \9 are not octal escapes.
	for _, d := range "123456789" {
		in := "\"\\" + string(d) + "\""
		t.Run("digit "+string(d), func(t *testing.T) {
			checkEscape(t, in, nil, true, 3)
		})
	}

	// \0 followed by a decimal digit is not an octal escape.
	for _, d := range "0123456789" {
		in := "\"\\0" + string(d) + "\""
		t.Run("null digit "+string(d), func(t *testing.T) {
			checkEscape(t, in, nil, true, 4)
		})
	}
}

// checkEscape unmarshals in and compares the result with want.
func checkEscape(t *testing.T, in string, want interface{}, fail bool, offset int64) {
	t.Helper()
	var got interface{}
	err := Unmarshal([]byte(in), &got)
	if fail {
		se, ok := err.(*SyntaxError)
		if !ok {
			t.Fatalf("Unmarshal(%q) error = %v, want syntax error", in, err)
		}
		if !strings.Contains(se.Error(), "escape") {
			t.Fatalf("Unmarshal(%q) error = %v, want an escape error", in, err)
		}
		if se.Offset != offset {
			t.Fatalf("Unmarshal(%q) offset = %d, want %d (%s)", in, se.Offset, offset, se)
		}
		return
	}
	if err != nil {
		t.Fatalf("Unmarshal(%q) unexpected error: %v", in, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Unmarshal(%q) = %#v, want %#v", in, got, want)
	}
}
