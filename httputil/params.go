package httputil

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/TakuyaYagam1/go-httpkit/httperr"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// GetUserID returns the user ID from context (set by auth middleware).
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// ParseUUID parses id from path/query and writes error response on failure.
func ParseUUID(w http.ResponseWriter, r *http.Request, id string) (uuid.UUID, bool) {
	if id == "" {
		HandleError(w, r, httperr.ErrInvalidID)
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		HandleError(w, r, httperr.ErrInvalidID)
		return uuid.Nil, false
	}
	return parsed, true
}

// ParseUUIDField parses value as UUID for a given field name (validation error).
func ParseUUIDField(w http.ResponseWriter, r *http.Request, value, field string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		HandleError(w, r, httperr.NewValidationErrorf("invalid %s", field))
		return uuid.Nil, false
	}
	return parsed, true
}

// ParseAuthUserID returns the authenticated user's UUID from context or writes error.
func ParseAuthUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID := GetUserID(r.Context())
	if userID == "" {
		HandleError(w, r, httperr.ErrNotAuthenticated)
		return uuid.Nil, false
	}
	return ParseUUID(w, r, userID)
}
