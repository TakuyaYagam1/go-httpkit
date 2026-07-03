package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer returns middleware that recovers panics, logs the panic and stack trace (if log is non-nil), and responds with 500 JSON. Place at the top of the chain
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &statusWriter{ResponseWriter: w}
			defer func(ctx context.Context) {
				if err := recover(); err != nil {
					if log != nil {
						log.ErrorContext(ctx, "panic recovered",
							slog.Any("panic", err),
							slog.String("stack", string(debug.Stack())),
						)
					}
					if rw.claimHeaderSent() {
						rw.ResponseWriter.Header().Set("Content-Type", contentTypeJSON)
						rw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
						_, _ = rw.ResponseWriter.Write([]byte(internalErrorResponseJSON))
					}
				}
			}(r.Context())
			next.ServeHTTP(rw, r)
		})
	}
}
