package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/pprof"
	"testing"
)

func TestHealthHandler_NoGoroutineLeaks(t *testing.T) {
	profile := goroutineLeakProfileOrSkip(t)
	handler := HealthHandler(map[string]Checker{
		"db": okChecker{},
	})
	for range 3 {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		handler(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode health body: %v", err)
		}
	}
	assertNoGoroutineLeaks(t, profile)
}

func goroutineLeakProfileOrSkip(t *testing.T) *pprof.Profile {
	t.Helper()
	profile := pprof.Lookup("goroutineleak")
	if profile == nil {
		t.Skip("goroutineleak profile requires GOEXPERIMENT=goroutineleakprofile")
	}
	return profile
}

func assertNoGoroutineLeaks(t *testing.T, profile *pprof.Profile) {
	t.Helper()
	runtime.GC() //nolint:revive // goroutineleak detection is intentionally GC-driven.
	runtime.GC() //nolint:revive // run twice so finalizers and profile reachability settle.
	if count := profile.Count(); count > 0 {
		var buf bytes.Buffer
		_ = profile.WriteTo(&buf, 1)
		t.Fatalf("goroutine leaks detected: count=%d\n%s", count, buf.String())
	}
}
