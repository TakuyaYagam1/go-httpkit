package httperr

import (
	"fmt"
	"net/http"
)

// NewValidationErrorf creates an HTTPError with status 400 and code VALIDATION_ERROR for dynamic validation messages.
// For semantic "request body valid JSON but business validation failed" use ErrUnprocessableEntity (422) from sentinels.
func NewValidationErrorf(format string, args ...any) *HTTPError {
	return &HTTPError{
		Err:        fmt.Errorf(format, args...),
		StatusCode: http.StatusBadRequest,
		Code:       "VALIDATION_ERROR",
		IsExpected: true,
	}
}
