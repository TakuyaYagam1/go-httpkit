// Package middleware provides HTTP middleware for use with go-httpkit
//
// # Client IP
//
// ClientIP resolves the client IP (using trusted proxy CIDRs for X-Real-IP and X-Forwarded-For) and stores it in the request context. CIDRs are parsed once at build time. Use GetClientIPFromContext in handlers to read it. Returns an error if all CIDR entries are invalid; empty or nil slice means no proxy trust
//
// # Logging
//
// Logger logs each request (method, path, redacted query, IP, user-agent, request_id) and after the handler adds status, latency_ms, and bytes. Log level: Info for 2xx, Warn for 4xx, Error for 5xx. Sensitive query params (token, password, secret, api_key, client_secret, refresh_token, access_token, authorization, state, code) are always redacted. Use WithRedactedParams to add extra names. Use WithSkipPaths to suppress logging for specific paths (e.g. /health, /metrics) - the handler still runs, only the log entry is omitted. If log is nil, the middleware is a no-op. CIDRs are parsed once at construction
//
// # Recoverer
//
// Recoverer recovers panics, logs the panic and stack trace (if log is non-nil), and responds with 500 JSON. Place it at the top of the middleware chain
//
// # Request ID
//
// RequestID sets or propagates X-Request-ID (from header or new UUID), validates format to prevent response splitting, and stores it in context. Use GetRequestID to read it
//
// # JSON requests
//
// RequireJSON validates application/json or +json Content-Type for POST/PUT/PATCH and requests that carry a body. When maxBodyBytes > 0, it wraps the body with http.MaxBytesReader before the handler runs
//
// # Security headers
//
// SecurityHeaders sets X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy, Content-Security-Policy, Cross-Origin-Opener-Policy, and Cross-Origin-Resource-Policy. Use WithHSTS for HTTPS-only services
//
// # Context timeout
//
// ContextTimeout attaches a deadline to r.Context() and runs the handler inline. When handlers return context.DeadlineExceeded, httputil.HandleError renders 503 JSON with code TIMEOUT
package middleware
