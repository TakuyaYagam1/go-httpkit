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
)
