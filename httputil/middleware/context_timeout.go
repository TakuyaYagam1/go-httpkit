package middleware

import (
	"context"
	"net/http"
	"time"
)

// ContextTimeout attaches a timeout to the request context and runs the handler inline.
// Handlers should pass r.Context() into database, cache, RPC, and other blocking calls.
func ContextTimeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
