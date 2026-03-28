package httputil

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestChiPathFromRequest_NoContext(t *testing.T) {
	t.Parallel()
	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	got := ChiPathFromRequest(r)
	if got != "" {
		t.Errorf("ChiPathFromRequest(no context) = %q, want empty", got)
	}
}

func TestChiPathFromRequest_WithPattern(t *testing.T) {
	t.Parallel()
	r, _ := http.NewRequest(http.MethodGet, "/api/v1/users", http.NoBody)
	rctx := chi.NewRouteContext()
	rctx.RoutePatterns = []string{"/api/v1/users"}
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	got := ChiPathFromRequest(r)
	if got != "/api/v1/users" {
		t.Errorf("ChiPathFromRequest(with pattern) = %q, want /api/v1/users", got)
	}
}

func TestChiPathFromRequest_EmptyPatterns(t *testing.T) {
	t.Parallel()
	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	rctx := chi.NewRouteContext()
	rctx.RoutePatterns = []string{}
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	got := ChiPathFromRequest(r)
	if got != "" {
		t.Errorf("ChiPathFromRequest(empty patterns) = %q, want empty", got)
	}
}
