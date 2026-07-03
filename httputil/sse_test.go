package httputil

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSEWriter_Flushable(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, ok := NewSSEWriter(w)
	require.True(t, ok)
	require.NotNil(t, sw)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewSSEWriter_NotFlushable(t *testing.T) {
	t.Parallel()
	w := &nonFlushWriter{ResponseWriter: httptest.NewRecorder()}
	sw, ok := NewSSEWriter(w)
	assert.False(t, ok)
	assert.Nil(t, sw)
}

type nonFlushWriter struct {
	http.ResponseWriter
}

func TestSSEWriter_Send(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, ok := NewSSEWriter(w)
	require.True(t, ok)
	err := sw.Send("ev", "line1")
	require.NoError(t, err)
	body := w.Body.String()
	assert.Contains(t, body, "event: ev\n")
	assert.Contains(t, body, "data: line1\n")
	assert.True(t, strings.HasSuffix(body, "\n\n"))
}

func TestSSEWriter_Send_NoEvent(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, _ := NewSSEWriter(w)
	err := sw.Send("", "data")
	require.NoError(t, err)
	assert.NotContains(t, w.Body.String(), "event:")
	assert.Contains(t, w.Body.String(), "data: data\n")
}

func TestSSEWriter_Send_MultilineData(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, _ := NewSSEWriter(w)
	err := sw.Send("e", "a\nb")
	require.NoError(t, err)
	body := w.Body.String()
	assert.Contains(t, body, "data: a\n")
	assert.Contains(t, body, "data: b\n")
}

func TestSSEWriter_SendJSON(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, _ := NewSSEWriter(w)
	err := sw.SendJSON("msg", map[string]int{"x": 1})
	require.NoError(t, err)
	body := w.Body.String()
	assert.Contains(t, body, "event: msg\n")
	assert.Contains(t, body, `data: {"x":1}`)
}

func TestSSEWriter_Send_PayloadTooLargeDoesNotCommitHeader(t *testing.T) {
	t.Parallel()
	w := newTrackingFlushWriter()
	sw, ok := NewSSEWriterWithLimit(w, MaxEventBytes(8))
	require.True(t, ok)

	err := sw.Send("event", "payload")

	require.ErrorIs(t, err, ErrSSEPayloadTooLarge)
	assert.Zero(t, w.statusCode())
	assert.Zero(t, w.writeHeaderCount())
	assert.Empty(t, w.bodyString())
}

func TestSSEWriter_Close_NoOpAfter(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, _ := NewSSEWriter(w)
	sw.Close()
	err := sw.Send("e", "d")
	require.ErrorIs(t, err, ErrSSEClosed)
	assert.Empty(t, w.Body.String())
}

func TestSSEWriter_Heartbeat_WritesComment(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, ok := NewSSEWriter(w)
	require.True(t, ok)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		sw.Heartbeat(ctx, 20*time.Millisecond)
	}()

	time.Sleep(60 * time.Millisecond)
	cancel()
	<-done

	body := w.Body.String()
	assert.Contains(t, body, ": ping\n\n")
}

func TestSSEWriter_HeartbeatMarksHeaderSent(t *testing.T) {
	t.Parallel()
	w := newTrackingFlushWriter()
	sw, ok := NewSSEWriter(w)
	require.True(t, ok)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		sw.Heartbeat(ctx, 5*time.Millisecond)
	}()

	require.Eventually(t, func() bool {
		return strings.Contains(w.bodyString(), ": ping\n\n")
	}, 200*time.Millisecond, 5*time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Heartbeat did not stop after context cancellation")
	}

	require.Equal(t, 1, w.writeHeaderCount())
	require.NoError(t, sw.Send("ev", "data"))
	assert.Equal(t, 1, w.writeHeaderCount())
}

func TestSSEWriter_Heartbeat_StopsWhenClosed(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	sw, _ := NewSSEWriter(w)

	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		defer close(done)
		sw.Heartbeat(ctx, 10*time.Millisecond)
	}()

	time.Sleep(25 * time.Millisecond)
	sw.Close()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Heartbeat did not stop after Close()")
	}
}

type trackingFlushWriter struct {
	mu               sync.Mutex
	header           http.Header
	body             bytes.Buffer
	status           int
	writeHeaderCalls int
}

func newTrackingFlushWriter() *trackingFlushWriter {
	return &trackingFlushWriter{header: make(http.Header)}
}

func (w *trackingFlushWriter) Header() http.Header {
	return w.header
}

func (w *trackingFlushWriter) WriteHeader(status int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writeHeaderCalls++
	if w.status == 0 {
		w.status = status
	}
}

func (w *trackingFlushWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status == 0 {
		w.status = http.StatusOK
		w.writeHeaderCalls++
	}
	return w.body.Write(p)
}

func (w *trackingFlushWriter) Flush() {}

func (w *trackingFlushWriter) bodyString() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.body.String()
}

func (w *trackingFlushWriter) statusCode() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *trackingFlushWriter) writeHeaderCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writeHeaderCalls
}
