package middleware

import (
	"net/http"
	"runtime/debug"

	logger "github.com/TakuyaYagam1/go-logkit"
)

// Recoverer recovers panics, logs the panic and stack trace with log (if non-nil), and responds with 500 JSON.
func Recoverer(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if log != nil {
						log.Error("panic recovered", logger.Fields{
							"panic": err,
							"stack": string(debug.Stack()),
						})
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"Internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
