package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	metricsOnce     sync.Once
)

func initMetrics() {
	metricsOnce.Do(func() {
		requestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)
		requestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		)
		prometheus.MustRegister(requestsTotal, requestDuration)
	})
}

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

// PathFromRequest returns the route pattern for the request (e.g. from chi.RouteContext).
// If nil, path will be "/unknown" unless status is 404 then "/not-found".
type PathFromRequest func(*http.Request) string

// Metrics returns middleware that records request count and duration. pathFromRequest can be nil.
func Metrics(pathFromRequest PathFromRequest) func(http.Handler) http.Handler {
	initMetrics()
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
