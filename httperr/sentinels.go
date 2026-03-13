package httperr

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidID = &HTTPError{
		Err:        errors.New("invalid ID"),
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_ID",
	}
	ErrNotAuthenticated = &HTTPError{
		Err:        errors.New("not authenticated"),
		StatusCode: http.StatusUnauthorized,
		Code:       "NOT_AUTHENTICATED",
	}
)
