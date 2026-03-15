// Package middleware provides HTTP middleware for go-httpkit.
//
// Metrics records request count and duration (Prometheus); optional Registerer and PathFromRequest.
// Recoverer recovers panics and responds with 500; pass a go-logkit Logger for stack logging.
// Timeout cancels the request context after a duration and responds with 503 on timeout.
// SecurityHeaders sets common security response headers. RequestID sets or propagates X-Request-ID
// and stores it in context; use GetRequestID to read it.
package middleware
