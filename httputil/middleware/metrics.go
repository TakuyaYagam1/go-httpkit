package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type wrapWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrapWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *wrapWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *wrapWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *wrapWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}

func (w *wrapWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// PathFromRequest returns the route pattern for the request (e.g. from chi.RouteContext).
// If nil, path will be "/unknown" unless status is 404 then "/not-found".
type PathFromRequest func(*http.Request) string

// Metrics returns middleware that records request count and duration. reg can be nil to use prometheus.DefaultRegisterer. pathFromRequest can be nil.
func Metrics(reg prometheus.Registerer, pathFromRequest PathFromRequest) func(http.Handler) http.Handler {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	_ = reg.Register(requestsTotal)
	_ = reg.Register(requestDuration)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &wrapWriter{ResponseWriter: w, status: 0}
			next.ServeHTTP(ww, r)
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.Status())
			path := "/unknown"
			if pathFromRequest != nil {
				path = pathFromRequest(r)
			}
			if path == "" {
				if ww.Status() == http.StatusNotFound {
					path = "/not-found"
				} else {
					path = "/unknown"
				}
			}
			method := r.Method
			requestsTotal.WithLabelValues(method, path, status).Inc()
			requestDuration.WithLabelValues(method, path, status).Observe(duration)
		})
	}
}
