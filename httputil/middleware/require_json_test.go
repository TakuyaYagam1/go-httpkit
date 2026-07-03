package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wahrwelt-kit/go-httpkit/httperr"
)

func TestRequireJSON_AcceptsApplicationJSON(t *testing.T) {
	t.Parallel()
	handler := RequireJSON(0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRequireJSON_AcceptsVendorJSON(t *testing.T) {
	t.Parallel()
	handler := RequireJSON(0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRequireJSON_RejectsMissingContentTypeForBodyMethod(t *testing.T) {
	t.Parallel()
	handler := RequireJSON(0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"ok":true}`))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.JSONEq(t, `{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}`, w.Body.String())
}

func TestRequireJSON_RejectsNonJSONContentType(t *testing.T) {
	t.Parallel()
	handler := RequireJSON(0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Body.String(), httperr.CodeUnsupportedMediaType)
}

func TestRequireJSON_AllowsGetWithoutBody(t *testing.T) {
	t.Parallel()
	handler := RequireJSON(0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRequireJSON_LimitsBody(t *testing.T) {
	t.Parallel()
	var readErr error
	handler := RequireJSON(4)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
		if readErr != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"too":"large"}`))
	req.Header.Set("Content-Type", contentTypeJSON)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var maxBytesErr *http.MaxBytesError
	require.ErrorAs(t, readErr, &maxBytesErr)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}
