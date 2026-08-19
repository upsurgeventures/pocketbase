package blob

import "testing"

func TestHexEscapeRoundTrip(t *testing.T) {
	tests := []string{"plain", "hello world", "cafeU0001f642", "already__0x20__escaped"}
	for _, input := range tests {
		escaped := HexEscape(input, func(_ []rune, i int) bool { return i%2 == 0 })
		if got := HexUnescape(escaped); got != input {
			t.Errorf("HexUnescape(HexEscape(%q)) = %q", input, got)
		}
	}
}

func TestHexUnescapeLeavesMalformedSequences(t *testing.T) {
	for _, input := range []string{"__0x__", "__0x20_", "__0xZZ__", "text"} {
		if got := HexUnescape(input); got != input {
			t.Errorf("HexUnescape(%q) = %q", input, got)
		}
	}
}
