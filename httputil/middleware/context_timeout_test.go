package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextTimeout_AttachesDeadline(t *testing.T) {
	t.Parallel()
	var deadline time.Time
	var hasDeadline bool
	chain := ContextTimeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, hasDeadline = r.Context().Deadline()
		w.WriteHeader(http.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	require.True(t, hasDeadline)
	assert.WithinDuration(t, time.Now().Add(time.Second), deadline, 100*time.Millisecond)
}

func TestContextTimeout_CancelsContext(t *testing.T) {
	t.Parallel()
	chain := ContextTimeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		assert.ErrorIs(t, r.Context().Err(), context.DeadlineExceeded)
		w.WriteHeader(http.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestContextTimeout_NonPositiveNoop(t *testing.T) {
	t.Parallel()
	var hasDeadline bool
	chain := ContextTimeout(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasDeadline = r.Context().Deadline()
		w.WriteHeader(http.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	chain.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.False(t, hasDeadline)
}
