package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_AllOk(t *testing.T) {
	t.Parallel()
	handler := HealthHandler(map[string]Checker{
		"db":    okChecker{},
		"redis": okChecker{},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "ok", body.Checks["db"])
	assert.Equal(t, "ok", body.Checks["redis"])
}

func TestHealthHandler_Degraded(t *testing.T) {
	t.Parallel()
	handler := HealthHandler(map[string]Checker{
		"db":    okChecker{},
		"redis": errChecker{},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler(w, r)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body.Status)
	assert.Equal(t, "ok", body.Checks["db"])
	assert.Equal(t, "error", body.Checks["redis"])
}

func TestHealthHandler_NilChecker(t *testing.T) {
	t.Parallel()
	handler := HealthHandler(map[string]Checker{
		"db": nil,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Checks map[string]string `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Checks["db"])
}

type okChecker struct{}

func (okChecker) Check(context.Context) error { return nil }

type errChecker struct{}

func (errChecker) Check(context.Context) error { return errors.New("unavailable") }
