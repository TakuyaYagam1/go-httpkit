package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedLogRecord struct {
	level   slog.Level
	message string
	attrs   map[string]any
}

type captureLogSink struct {
	mu      sync.Mutex
	records []capturedLogRecord
}

func (s *captureLogSink) Records() []capturedLogRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]capturedLogRecord(nil), s.records...)
}

func (s *captureLogSink) Last() capturedLogRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) == 0 {
		return capturedLogRecord{attrs: map[string]any{}}
	}
	return s.records[len(s.records)-1]
}

type captureLogHandler struct {
	sink  *captureLogSink
	attrs []slog.Attr
}

func newCaptureLogger() (*slog.Logger, *captureLogSink) {
	sink := &captureLogSink{}
	return slog.New(&captureLogHandler{sink: sink}), sink
}

func (h *captureLogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *captureLogHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, attr := range h.attrs {
		attrs[attr.Key] = attr.Value.Any()
	}
	r.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})
	h.sink.mu.Lock()
	h.sink.records = append(h.sink.records, capturedLogRecord{
		level:   r.Level,
		message: r.Message,
		attrs:   attrs,
	})
	h.sink.mu.Unlock()
	return nil
}

func (h *captureLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := &captureLogHandler{
		sink:  h.sink,
		attrs: append([]slog.Attr(nil), h.attrs...),
	}
	next.attrs = append(next.attrs, attrs...)
	return next
}

func (h *captureLogHandler) WithGroup(string) slog.Handler {
	return h
}

func TestLogger_CallsNext(t *testing.T) {
	t.Parallel()
	log, _ := newCaptureLogger()

	called := false
	handler := Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLogger_LogsInfo_OnSuccess(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("User-Agent", "test-agent")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	rec := sink.Last()
	assert.Equal(t, slog.LevelInfo, rec.level)
	assert.Equal(t, "http request", rec.message)
	assert.Equal(t, "GET", rec.attrs["method"])
	assert.Equal(t, "/health", rec.attrs["path"])
	assert.Equal(t, "192.168.1.1", rec.attrs["ip"])
	assert.Equal(t, "test-agent", rec.attrs["user_agent"])
	assert.EqualValues(t, http.StatusOK, rec.attrs["status"])
	assert.Contains(t, rec.attrs, "latency_ms")
	assert.Contains(t, rec.attrs, "bytes")
}

func TestLogger_LogsWarn_On4xx(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	req := httptest.NewRequest(http.MethodGet, "/forbidden", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	rec := sink.Last()
	assert.Equal(t, slog.LevelWarn, rec.level)
	assert.Equal(t, "http request error", rec.message)
	assert.EqualValues(t, http.StatusForbidden, rec.attrs["status"])
}

func TestLogger_LogsError_On5xx(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/broken", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	rec := sink.Last()
	assert.Equal(t, slog.LevelError, rec.level)
	assert.Equal(t, "http request failed", rec.message)
	assert.EqualValues(t, http.StatusInternalServerError, rec.attrs["status"])
}

func TestLogger_IncludesQueryAndRequestID_WhenSet(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := RequestID()(Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&page=1", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	rec := sink.Last()
	assert.Equal(t, "q=test&page=1", rec.attrs["query"])
	assert.Contains(t, rec.attrs, "request_id")
	assert.NotEmpty(t, rec.attrs["request_id"])
}

func TestLogger_RedactsSensitiveQueryParams(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/?token=secret&page=1", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	query, ok := sink.Last().attrs["query"].(string)
	require.True(t, ok)
	assert.Contains(t, query, "REDACTED")
	assert.NotContains(t, query, "secret")
}

func TestLogger_WithRedactedParams_RedactsCustomParam(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil, WithRedactedParams("apiToken", "x_custom"))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/?apiToken=abc&x_custom=val&safe=ok", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	query, ok := sink.Last().attrs["query"].(string)
	require.True(t, ok)
	assert.Contains(t, query, "REDACTED")
	assert.NotContains(t, query, "abc")
	assert.NotContains(t, query, "val")
	assert.Contains(t, query, "safe=ok")
}

func TestLogger_WithSkipPaths_DoesNotLog(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	called := false
	handler := Logger(log, nil, WithSkipPaths("/health", "/ready"))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called, "handler must still be called for skipped paths")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, sink.Records())
}

func TestLogger_WithSkipPaths_LogsNonSkipped(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil, WithSkipPaths("/health"))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "http request", sink.Last().message)
}

func TestLogger_DoesNotRedactSubstringParamName(t *testing.T) {
	t.Parallel()
	log, sink := newCaptureLogger()

	handler := Logger(log, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/?mytokenvalue=foo", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	query, ok := sink.Last().attrs["query"].(string)
	require.True(t, ok)
	assert.Equal(t, "mytokenvalue=foo", query)
}
