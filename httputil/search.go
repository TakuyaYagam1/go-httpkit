package httputil

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultSearchMaxLen = 100

// ValidateSearchQ returns true if q is valid for search (length and no control chars).
func ValidateSearchQ(q string) bool {
	if utf8.RuneCountInString(q) > DefaultSearchMaxLen {
		return false
	}
	for _, r := range q {
		if r == 0 || r == '\n' || r == '\r' || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// EscapeILIKE escapes %, _, \ for safe use in PostgreSQL ILIKE.
func EscapeILIKE(s string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = DefaultSearchMaxLen
	}
	if utf8.RuneCountInString(s) > maxLen {
		runes := []rune(s)
		s = string(runes[:maxLen])
	}
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '%':
			b.WriteString(`\%`)
		case '_':
			b.WriteString(`\_`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SanitizeSearchQ is EscapeILIKE with default max length.
func SanitizeSearchQ(q string, maxLen int) string {
	return EscapeILIKE(q, maxLen)
}
