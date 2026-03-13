package httperr

import (
	"net/http"
	"testing"
)

func TestCodeFromStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status int
		want   string
	}{
		{http.StatusBadRequest, "BAD_REQUEST"},
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusForbidden, "FORBIDDEN"},
		{http.StatusNotFound, "NOT_FOUND"},
		{http.StatusConflict, "CONFLICT"},
		{http.StatusGone, "GONE"},
		{http.StatusPaymentRequired, "PAYMENT_REQUIRED"},
		{http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED"},
		{http.StatusInternalServerError, "INTERNAL_ERROR"},
		{999, "INTERNAL_ERROR"},
	}
	for _, tt := range tests {
		got := CodeFromStatus(tt.status)
		if got != tt.want {
			t.Errorf("CodeFromStatus(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	err := New(nil, http.StatusBadRequest, "CUSTOM")
	if err == nil {
		t.Fatal("New(nil, ...) should not return nil")
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusBadRequest)
	}
	if err.GetCode() != "CUSTOM" {
		t.Errorf("GetCode() = %q, want CUSTOM", err.GetCode())
	}
	if err.Unwrap() == nil {
		t.Error("Unwrap() should not be nil")
	}
}

func TestNewValidationErrorf(t *testing.T) {
	t.Parallel()
	err := NewValidationErrorf("field %s invalid", "x")
	if err == nil {
		t.Fatal("NewValidationErrorf should not return nil")
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusBadRequest)
	}
	if err.GetCode() != "VALIDATION_ERROR" {
		t.Errorf("GetCode() = %q, want VALIDATION_ERROR", err.GetCode())
	}
	if err.Error() == "" {
		t.Error("Error() should not be empty")
	}
}
