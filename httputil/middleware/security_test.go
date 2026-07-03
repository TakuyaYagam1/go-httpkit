package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	chain := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)
	h := w.Header()
	assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", h.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", h.Get("Referrer-Policy"))
	assert.NotEmpty(t, h.Get("Permissions-Policy"))
	assert.Equal(t, "same-origin", h.Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "same-origin", h.Get("Cross-Origin-Resource-Policy"))
	assert.NotEmpty(t, h.Get("Content-Security-Policy"))
	assert.Empty(t, h.Get("Strict-Transport-Security"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecurityHeaders_WithHSTS(t *testing.T) {
	t.Parallel()
	chain := SecurityHeaders(WithHSTS(2*365*24*time.Hour, true, true))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=63072000")
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "includeSubDomains")
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "preload")
}

func TestSecurityHeaders_WithPolicyOverrides(t *testing.T) {
	t.Parallel()
	chain := SecurityHeaders(
		WithCSP(""),
		WithPermissionsPolicy("fullscreen=(self)"),
		WithCrossOriginOpenerPolicy("same-origin-allow-popups"),
		WithCrossOriginResourcePolicy("cross-origin"),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)
	h := w.Header()
	assert.Empty(t, h.Get("Content-Security-Policy"))
	assert.Equal(t, "fullscreen=(self)", h.Get("Permissions-Policy"))
	assert.Equal(t, "same-origin-allow-popups", h.Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "cross-origin", h.Get("Cross-Origin-Resource-Policy"))
}
