package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverer_NoPanic(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.DiscardHandler)
	chain := Recoverer(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecoverer_Panic_500(t *testing.T) {
	t.Parallel()
	chain := Recoverer(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	if ct := w.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Fatalf("Content-Type = %q, want %q", ct, contentTypeJSON)
	}
	assert.JSONEq(t, internalErrorResponseJSON, w.Body.String())
}

func TestRecoverer_NilLogger(t *testing.T) {
	t.Parallel()
	chain := Recoverer(nil)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("x")
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
