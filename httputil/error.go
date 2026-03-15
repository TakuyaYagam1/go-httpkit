package httputil

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
	logger "github.com/TakuyaYagam1/go-logkit"
)

type ErrorLogEvent interface {
	Info(msg string)
	Error(msg string)
}

type ErrorLogger interface {
	WithError(err error) ErrorLogEvent
}

type logkitErrorAdapter struct{ l logger.Logger }

func (a *logkitErrorAdapter) WithError(err error) ErrorLogEvent {
	return &logkitErrorEvent{a.l.WithError(err)}
}

type logkitErrorEvent struct{ l logger.Logger }

func (e *logkitErrorEvent) Info(msg string)  { e.l.Info(msg) }
func (e *logkitErrorEvent) Error(msg string) { e.l.Error(msg) }

func NewErrorLogger(l logger.Logger) ErrorLogger {
	if l == nil {
		return nil
	}
	return &logkitErrorAdapter{l: l}
}

type ErrorHandler struct {
	Logger ErrorLogger
}

func (h *ErrorHandler) Handle(w http.ResponseWriter, r *http.Request, err error, msg string) bool {
	if err == nil {
		return false
	}
	if h.Logger != nil {
		ev := h.Logger.WithError(err)
		if httperr.IsExpectedClientError(err) {
			ev.Info(msg)
		} else {
			ev.Error(msg)
		}
	}
	HandleError(w, r, err)
	return true
}

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
