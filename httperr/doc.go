// Package httperr provides HTTP-aware error types for use with JSON APIs.
//
// HTTPError carries status code and application error code (e.g. BAD_REQUEST, NOT_FOUND).
// CodeFromStatus maps HTTP status codes to default application codes. Use New for custom
// errors and NewValidationErrorf for validation failures. Sentinel errors (ErrInvalidID,
// ErrNotAuthenticated) are provided for common cases.
package httperr
