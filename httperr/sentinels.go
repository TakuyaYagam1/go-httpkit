package httperr

import (
	"errors"
	"net/http"
)

var (
	// ErrInvalidID is a 400 error with code INVALID_ID (e.g. invalid UUID in path).
	ErrInvalidID = &HTTPError{
		Err:        errors.New("invalid ID"),
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_ID",
		IsExpected: true,
	}
	// ErrNotAuthenticated is a 401 error with code NOT_AUTHENTICATED.
	ErrNotAuthenticated = &HTTPError{
		Err:        errors.New("not authenticated"),
		StatusCode: http.StatusUnauthorized,
		Code:       "NOT_AUTHENTICATED",
		IsExpected: true,
	}
	// ErrForbidden is a 403 error with code FORBIDDEN.
	ErrForbidden = &HTTPError{
		Err:        errors.New("forbidden"),
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		IsExpected: true,
	}
	// ErrNotFound is a 404 error with code NOT_FOUND.
	ErrNotFound = &HTTPError{
		Err:        errors.New("not found"),
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		IsExpected: true,
	}
	// ErrConflict is a 409 error with code CONFLICT.
	ErrConflict = &HTTPError{
		Err:        errors.New("conflict"),
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		IsExpected: true,
	}
	// ErrGone is a 410 error with code GONE.
	ErrGone = &HTTPError{
		Err:        errors.New("gone"),
		StatusCode: http.StatusGone,
		Code:       "GONE",
		IsExpected: true,
	}
	// ErrUnprocessableEntity is a 422 error with code VALIDATION_ERROR.
	ErrUnprocessableEntity = &HTTPError{
		Err:        errors.New("unprocessable entity"),
		StatusCode: http.StatusUnprocessableEntity,
		Code:       "VALIDATION_ERROR",
		IsExpected: true,
	}
	// ErrTooManyRequests is a 429 error with code RATE_LIMIT_EXCEEDED.
	ErrTooManyRequests = &HTTPError{
		Err:        errors.New("too many requests"),
		StatusCode: http.StatusTooManyRequests,
		Code:       "RATE_LIMIT_EXCEEDED",
		IsExpected: true,
	}
	// ErrServiceUnavailable is a 503 error with code SERVICE_UNAVAILABLE.
	ErrServiceUnavailable = &HTTPError{
		Err:        errors.New("service unavailable"),
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		IsExpected: false,
	}
)
