package middleware

const (
	keyMethod = "method"
	keyPath   = "path"
	keyStatus = "status"

	contentTypeJSON           = "application/json"
	internalErrorResponseJSON = `{"code":"INTERNAL_ERROR","message":"Internal server error"}`
)
