package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
)

func TestHandleError_HTTPError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	err := httperr.New(errors.New("not found"), http.StatusNotFound, "NOT_FOUND")
	HandleError(w, r, err)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	var body ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", body.Code)
	}
	if body.Message != "not found" {
		t.Errorf("Message = %q", body.Message)
	}
}

func TestHandleError_HTTPError_EmptyCodeUsesCodeFromStatus(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	err := httperr.New(errors.New("bad"), http.StatusBadRequest, "")
	HandleError(w, r, err)
	var body ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "BAD_REQUEST" {
		t.Errorf("Code = %q, want BAD_REQUEST", body.Code)
	}
}

func TestHandleError_HTTPError_5xxHidesMessage(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	err := httperr.New(errors.New("internal detail"), http.StatusInternalServerError, "INTERNAL_ERROR")
	HandleError(w, r, err)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", w.Code)
	}
	var body ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Message != "Internal server error" {
		t.Errorf("Message = %q, want generic message", body.Message)
	}
}

func TestHandleError_GenericError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	HandleError(w, r, errors.New("generic"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var body ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %q", body.Code)
	}
	if body.Message != "Internal server error" {
		t.Errorf("Message = %q", body.Message)
	}
}
