package httputil

import (
	"strings"
	"testing"
)

func FuzzSanitizeContentDispositionFilename(f *testing.F) {
	f.Add("report.json")
	f.Add("../secret.txt")
	f.Add("bad\r\nContent-Type: text/html")
	f.Add(`quote"slash\name`)
	f.Fuzz(func(t *testing.T, name string) {
		got := sanitizeContentDispositionFilename(name)
		if got == "" || got == "." {
			t.Fatalf("empty unsafe filename for %q", name)
		}
		if len(got) > maxContentDispositionFilenameLen {
			t.Fatalf("filename len = %d, want <= %d", len(got), maxContentDispositionFilenameLen)
		}
		if strings.ContainsAny(got, "\r\n\"\\") {
			t.Fatalf("filename contains unsafe chars: %q", got)
		}
	})
}
