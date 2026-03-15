// Package httputil provides HTTP helpers: error handling and JSON responses, request
// decoding and validation, pagination (ClampPage, ClampPerPage, ParseIntQuery), query
// parsing (ParseBoolQuery, ParseEnumQuery, ParseSortQuery, ParseTimeQuery), client IP
// resolution with proxy headers, search sanitization (EscapeILIKE), multipart limits,
// file download responses, chi route pattern (ChiPathFromRequest), SSE (SSEWriter,
// NewSSEWriter), and health checks (Checker, HealthHandler).
package httputil
