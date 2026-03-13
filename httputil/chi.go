package httputil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ChiPathFromRequest returns the route pattern from chi's RouteContext, or empty string if none.
func ChiPathFromRequest(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return ""
	}
	return rctx.RoutePattern()
}
