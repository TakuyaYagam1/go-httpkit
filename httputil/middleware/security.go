package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCSP               = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'"
	defaultPermissionsPolicy = "geolocation=(), microphone=(), camera=()"
	defaultCOOP              = "same-origin"
	defaultCORP              = "same-origin"
)

type securityOpts struct {
	csp               string
	permissionsPolicy string
	coop              string
	corp              string
	hsts              hstsOpts
}

type hstsOpts struct {
	maxAge            time.Duration
	includeSubDomains bool
	preload           bool
}

// SecurityOption configures SecurityHeaders
type SecurityOption func(*securityOpts)

// WithCSP sets the Content-Security-Policy header. Empty string leaves CSP unset for this middleware
func WithCSP(csp string) SecurityOption {
	return func(o *securityOpts) { o.csp = csp }
}

// WithPermissionsPolicy sets the Permissions-Policy header. Empty string leaves the header unset.
func WithPermissionsPolicy(policy string) SecurityOption {
	return func(o *securityOpts) { o.permissionsPolicy = policy }
}

// WithCrossOriginOpenerPolicy sets the Cross-Origin-Opener-Policy header. Empty string leaves the header unset.
func WithCrossOriginOpenerPolicy(policy string) SecurityOption {
	return func(o *securityOpts) { o.coop = policy }
}

// WithCrossOriginResourcePolicy sets the Cross-Origin-Resource-Policy header. Empty string leaves the header unset.
func WithCrossOriginResourcePolicy(policy string) SecurityOption {
	return func(o *securityOpts) { o.corp = policy }
}

// WithHSTS sets the Strict-Transport-Security header. Use only for HTTPS-only services.
func WithHSTS(maxAge time.Duration, includeSubDomains, preload bool) SecurityOption {
	return func(o *securityOpts) {
		o.hsts = hstsOpts{maxAge: maxAge, includeSubDomains: includeSubDomains, preload: preload}
	}
}

// SecurityHeaders returns middleware that sets common security headers. Options override defaults and HSTS is enabled only via WithHSTS.
func SecurityHeaders(opts ...SecurityOption) func(http.Handler) http.Handler {
	cfg := securityOpts{
		csp:               defaultCSP,
		permissionsPolicy: defaultPermissionsPolicy,
		coop:              defaultCOOP,
		corp:              defaultCORP,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			if cfg.permissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.permissionsPolicy)
			}
			if cfg.coop != "" {
				w.Header().Set("Cross-Origin-Opener-Policy", cfg.coop)
			}
			if cfg.corp != "" {
				w.Header().Set("Cross-Origin-Resource-Policy", cfg.corp)
			}
			if cfg.csp != "" {
				w.Header().Set("Content-Security-Policy", cfg.csp)
			}
			if hsts := hstsHeader(cfg.hsts); hsts != "" {
				w.Header().Set("Strict-Transport-Security", hsts)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hstsHeader(cfg hstsOpts) string {
	maxAge := int64(cfg.maxAge / time.Second)
	if maxAge <= 0 {
		return ""
	}
	parts := []string{"max-age=" + strconv.FormatInt(maxAge, 10)}
	if cfg.includeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	if cfg.preload {
		parts = append(parts, "preload")
	}
	return strings.Join(parts, "; ")
}
