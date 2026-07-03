package httputil

import (
	"strings"
	"testing"
	"unicode"
)

func FuzzSanitizeSSEField(f *testing.F) {
	f.Add("message")
	f.Add("event\r\nid: injected")
	f.Add("hello\x00world")
	f.Fuzz(func(t *testing.T, field string) {
		got := sanitizeSSEField(field)
		if strings.ContainsAny(got, "\r\n") {
			t.Fatalf("field contains newline after sanitize: %q", got)
		}
		for _, r := range got {
			if unicode.IsControl(r) {
				t.Fatalf("field contains control rune %U after sanitize: %q", r, got)
			}
		}
	})
}
