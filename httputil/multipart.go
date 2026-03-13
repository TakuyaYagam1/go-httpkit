package httputil

import (
	"net/http"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
)

// ParseMultipartFormLimit parses multipart form with maxMemory limit. On error writes response and returns false.
func ParseMultipartFormLimit(w http.ResponseWriter, r *http.Request, maxMemory int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		HandleError(w, r, httperr.NewValidationErrorf("failed to parse form"))
		return false
	}
	return true
}
