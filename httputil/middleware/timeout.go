package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type timeoutWriter struct {
	http.ResponseWriter
	mu       sync.Mutex
	timedOut bool
	wrote    bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.wrote || tw.timedOut {
		return
	}
	tw.wrote = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	if tw.timedOut {
		tw.mu.Unlock()
		return 0, context.DeadlineExceeded
	}
	tw.wrote = true
	tw.mu.Unlock()
	return tw.ResponseWriter.Write(b)
}

// Timeout runs the next handler with a context deadline of d. If the context is cancelled before the handler completes, responds with 503 JSON.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			r = r.WithContext(ctx)
			tw := &timeoutWriter{ResponseWriter: w}
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(tw, r)
				close(done)
			}()
			select {
			case <-done:
				return
			case <-ctx.Done():
				tw.mu.Lock()
				if tw.wrote {
					tw.mu.Unlock()
					return
				}
				tw.timedOut = true
				tw.ResponseWriter.Header().Set("Content-Type", "application/json")
				tw.ResponseWriter.WriteHeader(http.StatusServiceUnavailable)
				_, _ = tw.ResponseWriter.Write([]byte(`{"error":"request timeout"}`))
				tw.mu.Unlock()
			}
		})
	}
}
