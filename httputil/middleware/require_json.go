package middleware

import (
	"mime"
	"net/http"
	"strings"

	"github.com/wahrwelt-kit/go-httpkit/httperr"
)

const msgRequireJSONContentType = "Content-Type must be application/json"

// RequireJSON requires JSON Content-Type for POST/PUT/PATCH and requests that carry a body.
// It accepts application/json and vendor JSON media types ending in +json.
// When maxBodyBytes > 0, it wraps the request body with http.MaxBytesReader before the handler runs.
func RequireJSON(maxBodyBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requestRequiresJSON(r) && !isJSONContentType(r.Header.Get("Content-Type")) {
				writeJSONError(w, http.StatusUnsupportedMediaType, httperr.CodeUnsupportedMediaType, msgRequireJSONContentType)
				return
			}
			if maxBodyBytes > 0 && r.Body != nil && r.Body != http.NoBody {
				r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func requestRequiresJSON(r *http.Request) bool {
	if r == nil {
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return r.Body != nil && r.Body != http.NoBody && r.ContentLength != 0
	}
}

func isJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	mediaType = strings.ToLower(mediaType)
	return mediaType == contentTypeJSON || strings.HasSuffix(mediaType, "+json")
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"code":"` + code + `","message":"` + message + `"}`))
}
