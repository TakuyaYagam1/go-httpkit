package httputil

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
)

// RenderJSON writes data as JSON with the given status.
func RenderJSON[T any](w http.ResponseWriter, r *http.Request, status int, data T) {
	render.Status(r, status)
	render.JSON(w, r, data)
}

// RenderNoContent sends 204 No Content.
func RenderNoContent(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// RenderCreated sends 201 Created with JSON body.
func RenderCreated[T any](w http.ResponseWriter, r *http.Request, data T) {
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, data)
}

// RenderAccepted sends 202 Accepted with JSON body.
func RenderAccepted[T any](w http.ResponseWriter, r *http.Request, data T) {
	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, data)
}

// RenderOK sends 200 OK with JSON body.
func RenderOK[T any](w http.ResponseWriter, r *http.Request, data T) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// RenderError sends JSON error with status and message; code is derived from status.
func RenderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	render.Status(r, status)
	render.JSON(w, r, ErrorResponse{Code: httperr.CodeFromStatus(status), Message: message})
}

// RenderErrorWithCode sends JSON error with explicit code.
func RenderErrorWithCode(w http.ResponseWriter, r *http.Request, status int, message, code string) {
	render.Status(r, status)
	render.JSON(w, r, ErrorResponse{Code: code, Message: message})
}

// RenderInvalidID sends 400 with INVALID_ID code.
func RenderInvalidID(w http.ResponseWriter, r *http.Request) {
	RenderErrorWithCode(w, r, http.StatusBadRequest, "invalid ID", "INVALID_ID")
}

// RenderText writes a text response with given content type and body.
func RenderText(w http.ResponseWriter, _ *http.Request, status int, contentType, body string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
