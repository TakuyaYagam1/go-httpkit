package metrics

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_RecordsRequest(t *testing.T) {
	t.Parallel()
	reg := prometheus.NewRegistry()
	pathFn := func(*http.Request) string { return "/test" }
	chain := Middleware(reg, pathFn)
	handler := chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	metrics, err := reg.Gather()
	require.NoError(t, err)
	var found bool
	for _, m := range metrics {
		if m.GetName() == metricHTTPRequestsTotal {
			found = true
			break
		}
	}
	assert.True(t, found, "http_requests_total should be registered")
}

func TestMiddleware_WithOptions(t *testing.T) {
	t.Parallel()
	reg := prometheus.NewRegistry()
	requestsName := prometheus.BuildFQName("wahrwelt", "api", metricHTTPRequestsTotal)
	durationName := prometheus.BuildFQName("wahrwelt", "api", metricHTTPRequestDurationSeconds)
	inFlightName := prometheus.BuildFQName("wahrwelt", "api", metricHTTPRequestsInFlight)
	handler := Middleware(
		reg,
		func(*http.Request) string { return "/items/{id}" },
		WithNamespace("wahrwelt"),
		WithSubsystem("api"),
		WithBuckets([]float64{0.001, 0.01}),
		WithConstLabels(prometheus.Labels{"service": "orders"}),
		WithInFlight(),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/items/1", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	metrics, err := reg.Gather()
	require.NoError(t, err)
	familyByName := make(map[string]int, len(metrics))
	for i, family := range metrics {
		familyByName[family.GetName()] = i
	}
	requestsIdx, ok := familyByName[requestsName]
	require.True(t, ok)
	durationIdx, ok := familyByName[durationName]
	require.True(t, ok)
	inFlightIdx, ok := familyByName[inFlightName]
	require.True(t, ok)

	counterMetric := metrics[requestsIdx].GetMetric()[0]
	metricLabel := func(name string) string {
		for _, label := range counterMetric.GetLabel() {
			if label.GetName() == name {
				return label.GetValue()
			}
		}
		return ""
	}
	assert.Equal(t, "orders", metricLabel("service"))
	assert.Equal(t, http.MethodPost, metricLabel(keyMethod))
	assert.Equal(t, "/items/{id}", metricLabel(keyPath))
	assert.Equal(t, "201", metricLabel(keyStatus))
	require.NotNil(t, metrics[durationIdx].GetMetric()[0].GetHistogram())
	assert.Len(t, metrics[durationIdx].GetMetric()[0].GetHistogram().GetBucket(), 2)
	require.NotNil(t, metrics[inFlightIdx].GetMetric()[0].GetGauge())
	assert.InDelta(t, 0, metrics[inFlightIdx].GetMetric()[0].GetGauge().GetValue(), 0.0001)
}

func TestMiddleware_AlreadyRegisteredCollectorWithIncompatibleLabelsDoesNotPanic(t *testing.T) {
	t.Parallel()
	reg := &alreadyRegisteredRegisterer{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: metricHTTPRequestsTotal, Help: "Total number of HTTP requests"},
			[]string{keyMethod},
		),
		histogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: metricHTTPRequestDurationSeconds, Help: "HTTP request duration in seconds"},
			[]string{keyMethod},
		),
	}
	handler := Middleware(reg, func(*http.Request) string { return "/test" })(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() {
		handler.ServeHTTP(w, req)
	})
}

func TestMiddleware_UsesNotFoundPathFallback(t *testing.T) {
	t.Parallel()
	reg := prometheus.NewRegistry()
	handler := Middleware(reg, func(*http.Request) string { return "" })(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/missing/123", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	families, err := reg.Gather()
	require.NoError(t, err)
	var path string
	for _, family := range families {
		if family.GetName() != metricHTTPRequestsTotal {
			continue
		}
		for _, label := range family.GetMetric()[0].GetLabel() {
			if label.GetName() == keyPath {
				path = label.GetValue()
			}
		}
	}
	assert.Equal(t, "/not-found", path)
}

func TestStatusWriter_WriteImplicitStatus(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	w := &statusWriter{ResponseWriter: rec}

	n, err := w.Write([]byte("hello"))

	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())
}

func TestStatusWriter_FlushMarksOK(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	w := &statusWriter{ResponseWriter: rec}

	w.Flush()

	assert.True(t, rec.Flushed)
	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, http.StatusOK, rec.Code)
}

type fakeHijacker struct {
	http.ResponseWriter

	conn net.Conn
}

func (f *fakeHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw := bufio.NewReadWriter(bufio.NewReader(f.conn), bufio.NewWriter(f.conn))
	return f.conn, rw, nil
}

func TestStatusWriter_HijackSupported(t *testing.T) {
	t.Parallel()
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	w := &statusWriter{ResponseWriter: &fakeHijacker{ResponseWriter: httptest.NewRecorder(), conn: server}}

	conn, rw, err := w.Hijack()

	require.NoError(t, err)
	assert.NotNil(t, conn)
	assert.NotNil(t, rw)
}

func TestStatusWriter_HijackUnsupported(t *testing.T) {
	t.Parallel()
	w := &statusWriter{ResponseWriter: httptest.NewRecorder()}

	conn, rw, err := w.Hijack()

	require.ErrorIs(t, err, http.ErrNotSupported)
	assert.Nil(t, conn)
	assert.Nil(t, rw)
}

func TestStatusWriter_ReadFrom(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	w := &statusWriter{ResponseWriter: rec}

	n, err := w.ReadFrom(strings.NewReader("stream"))

	require.NoError(t, err)
	assert.Equal(t, int64(6), n)
	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "stream", rec.Body.String())
}

type readerFromWriter struct {
	*httptest.ResponseRecorder
}

func (w *readerFromWriter) ReadFrom(r io.Reader) (int64, error) {
	return w.Body.ReadFrom(r)
}

func TestStatusWriter_ReadFromUnderlyingReaderFrom(t *testing.T) {
	t.Parallel()
	rec := &readerFromWriter{ResponseRecorder: httptest.NewRecorder()}
	w := &statusWriter{ResponseWriter: rec}

	n, err := w.ReadFrom(strings.NewReader("reader-from"))

	require.NoError(t, err)
	assert.Equal(t, int64(11), n)
	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "reader-from", rec.Body.String())
}

type alreadyRegisteredRegisterer struct {
	counter   *prometheus.CounterVec
	histogram *prometheus.HistogramVec
}

func (r *alreadyRegisteredRegisterer) Register(c prometheus.Collector) error {
	switch c.(type) {
	case *prometheus.CounterVec:
		return prometheus.AlreadyRegisteredError{ExistingCollector: r.counter, NewCollector: c}
	case *prometheus.HistogramVec:
		return prometheus.AlreadyRegisteredError{ExistingCollector: r.histogram, NewCollector: c}
	default:
		return nil
	}
}

func (r *alreadyRegisteredRegisterer) MustRegister(...prometheus.Collector) {}

func (r *alreadyRegisteredRegisterer) Unregister(prometheus.Collector) bool {
	return false
}
