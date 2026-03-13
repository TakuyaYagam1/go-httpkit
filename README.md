# go-httpkit

Simple HTTP utilities for Go projects (chi, JSON, validation, errors, metrics).

## Install

```bash
go get github.com/TakuyaYagam1/go-httpkit
```

## Packages

- **`httperr`** — Typed HTTP errors with status and code, `Unwrap()`, sentinels (`ErrInvalidID`, `ErrNotAuthenticated`), `NewValidationErrorf`.
- **`httputil`** — JSON render (`RenderOK`, `RenderCreated`, `RenderError`, …), `HandleError` (map `httperr` to JSON), request decode/validate (`DecodeAndValidate`, `DecodeJSON` with size limit and no trailing data), UUID params (`ParseUUID`, `ParseAuthUserID`), `GetClientIP` (trusted proxies), multipart limit, download helpers (`RenderJSONAttachment`, `RenderStream`, `RenderBytes`), search (`EscapeILIKE`, `ValidateSearchQ`), pagination (`ClampPage`, `ParseIntQuery`, `Ptr`).
- **`httputil/middleware`** — Prometheus metrics (`http_requests_total`, `http_request_duration_seconds`); path from request via `PathFromRequest` (e.g. chi route pattern).

Requires **go-chi/render** and **go-playground/validator** in your app for full use. Validator interface: `Validate(any) error`.
