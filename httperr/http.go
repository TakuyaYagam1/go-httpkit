package httperr

import (
	"errors"
	"net/http"
)

// CodeFromStatus returns the application error code for a given HTTP status.
func CodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusGone:
		return "GONE"
	case http.StatusPaymentRequired:
		return "PAYMENT_REQUIRED"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	default:
		return "INTERNAL_ERROR"
	}
}

// New returns an HTTPError with the given error, status code, and code. IsExpected is true for 4xx.
func New(err error, status int, code string) *HTTPError {
	if err == nil {
		err = errors.New("")
	}
	return &HTTPError{
		Err:        err,
		StatusCode: status,
		Code:       code,
		IsExpected: status >= http.StatusBadRequest && status < 500,
	}
}

// HTTPError represents an error with HTTP status and application code.
type HTTPError struct {
	Err        error
	StatusCode int
	Code       string
	// IsExpected is true for client errors (4xx); callers may use it to avoid logging as server errors.
	IsExpected bool
}

func (e *HTTPError) Error() string   { return e.Err.Error() }
func (e *HTTPError) Unwrap() error   { return e.Err }
func (e *HTTPError) HTTPStatus() int { return e.StatusCode }
func (e *HTTPError) GetCode() string { return e.Code }
