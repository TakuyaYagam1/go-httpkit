package httputil

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
)

type httpErrorWithStatus interface {
	Error() string
	HTTPStatus() int
	GetCode() string
}

// ErrorResponse is the JSON shape for error responses.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorItem represents a single validation error.
type ValidationErrorItem struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse is the JSON shape for validation errors.
type ValidationErrorResponse struct {
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Errors  []ValidationErrorItem `json:"errors,omitempty"`
}

// HandleError writes a JSON error response. If err implements httpErrorWithStatus, uses its status and code; otherwise 500.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var httpErr httpErrorWithStatus
	if errors.As(err, &httpErr) {
		code := httpErr.GetCode()
		if code == "" {
			code = httperr.CodeFromStatus(httpErr.HTTPStatus())
		}
		message := httpErr.Error()
		if httpErr.HTTPStatus() >= http.StatusInternalServerError {
			message = "Internal server error"
		}
		render.Status(r, httpErr.HTTPStatus())
		render.JSON(w, r, ErrorResponse{
			Code:    code,
			Message: message,
		})
		return
	}

	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, ErrorResponse{Code: httperr.CodeFromStatus(http.StatusInternalServerError), Message: "Internal server error"})
}
