package httputil

import (
	"context"
	"encoding/json"
	"net/http"
)

// Checker performs a single health check. Check returns nil for success.
type Checker interface {
	Check(ctx context.Context) error
}

// HealthHandler returns a handler that runs all checkers and responds with JSON: {"status":"ok"|"degraded","checks":{name:"ok"|"error"}}.
// Status 200 when all pass, 503 when any check fails. Nil checkers are treated as ok.
func HealthHandler(checkers map[string]Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		results := make(map[string]string, len(checkers))
		allOk := true
		for name, c := range checkers {
			if c == nil {
				results[name] = "ok"
				continue
			}
			if err := c.Check(ctx); err != nil {
				results[name] = "error"
				allOk = false
			} else {
				results[name] = "ok"
			}
		}
		status := "ok"
		code := http.StatusOK
		if !allOk {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": status,
			"checks": results,
		})
	}
}
