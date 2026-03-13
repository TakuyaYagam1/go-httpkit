package httputil

import (
	"net/http"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
)

// ParseMultipartFormLimit parses multipart form with maxBodySize (MaxBytesReader) and maxMemory (ParseMultipartForm). On error writes response and returns false.
func ParseMultipartFormLimit(w http.ResponseWriter, r *http.Request, maxBodySize, maxMemory int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		HandleError(w, r, httperr.NewValidationErrorf("failed to parse form"))
		return false
	}
	return true
}
