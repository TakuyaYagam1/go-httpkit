// Package middleware provides HTTP middleware for go-httpkit.
//
// Metrics records request count and duration (Prometheus). It accepts an optional
// Registerer and a PathFromRequest function (e.g. ChiPathFromRequest from httputil).
package middleware
