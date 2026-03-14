package httputil

import (
	"net/http"
	"strconv"
)

const MaxPage = 10000

// ClampPage returns page clamped to [1, MaxPage].
func ClampPage(p *int) int {
	if p == nil || *p < 1 {
		return 1
	}
	if *p > MaxPage {
		return MaxPage
	}
	return *p
}

// ClampPerPage returns perPage clamped to [defaultVal, maxVal], or defaultVal if nil/<=0.
func ClampPerPage(p *int, defaultVal, maxVal int) int {
	if p == nil || *p <= 0 {
		return defaultVal
	}
	if *p > maxVal {
		return maxVal
	}
	return *p
}

// ClampLimit returns limit clamped to [defaultVal, maxVal], or defaultVal if nil/<=0.
func ClampLimit(p *int, defaultVal, maxVal int) int {
	return ClampPerPage(p, defaultVal, maxVal)
}

// ParseIntQuery parses the first query parameter key as positive int; returns nil if missing or invalid.
func ParseIntQuery(r *http.Request, key string) *int {
	q := r.URL.Query().Get(key)
	if q == "" {
		return nil
	}
	n, err := strconv.Atoi(q)
	if err != nil || n < 1 {
		return nil
	}
	return &n
}

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T {
	return &v
}

func TotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	n := int(total) / perPage
	if int(total)%perPage != 0 {
		n++
	}
	return n
}

type PaginationMeta struct {
	Page       int
	PerPage    int
	Total      int
	TotalPages int
}

func NewPaginationMeta(page, perPage int, total int64) PaginationMeta {
	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: TotalPages(total, perPage),
	}
}
