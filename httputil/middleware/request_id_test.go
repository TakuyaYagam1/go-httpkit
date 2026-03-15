package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_Generated(t *testing.T) {
	t.Parallel()
	chain := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		assert.NotEmpty(t, id)
		assert.Equal(t, id, w.Header().Get("X-Request-ID"))
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	chain.ServeHTTP(w, r)
	require.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestID_FromHeader(t *testing.T) {
	t.Parallel()
	const want = "existing-id-123"
	chain := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		assert.Equal(t, want, id)
		assert.Equal(t, want, w.Header().Get("X-Request-ID"))
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", want)
	chain.ServeHTTP(w, r)
	assert.Equal(t, want, w.Header().Get("X-Request-ID"))
}

func TestGetRequestID_Empty(t *testing.T) {
	t.Parallel()
	assert.Empty(t, GetRequestID(context.Background()))
}
