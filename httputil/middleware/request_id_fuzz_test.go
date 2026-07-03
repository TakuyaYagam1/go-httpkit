package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func FuzzRequestID(f *testing.F) {
	f.Add("req-123")
	f.Add("bad\r\nX-Evil: 1")
	f.Add("")
	f.Add(strings.Repeat("a", 129))
	f.Fuzz(func(t *testing.T, id string) {
		id = limitFuzzString(id, 512)
		handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := GetRequestID(r.Context()); got == "" || !validRequestID(got) {
				t.Fatalf("invalid request id in context: %q", got)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("X-Request-ID", id)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		got := w.Header().Get("X-Request-ID")
		if got == "" || !validRequestID(got) {
			t.Fatalf("invalid response request id: %q", got)
		}
		if validRequestID(id) && got != id {
			t.Fatalf("valid request id was not preserved: got %q want %q", got, id)
		}
	})
}

func limitFuzzString(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}
