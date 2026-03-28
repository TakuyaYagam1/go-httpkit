package httputil

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBoolQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		query   string
		key     string
		wantVal bool
		wantOk  bool
	}{
		{"missing", "", "x", false, false},
		{"true", "x=true", "x", true, true},
		{"1", "x=1", "x", true, true},
		{"yes", "x=yes", "x", true, true},
		{"false", "x=false", "x", false, true},
		{"0", "x=0", "x", false, true},
		{"no", "x=no", "x", false, true},
		{"invalid", "x=maybe", "x", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r *http.Request
			if tt.query != "" {
				r = requireReq(t, "GET", "http://a/?"+tt.query)
			} else {
				r = requireReq(t, "GET", "http://a/")
			}
			val, ok := ParseBoolQuery(r, tt.key)
			assert.Equal(t, tt.wantVal, val)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestParseEnumQuery(t *testing.T) {
	t.Parallel()
	type Status string
	allowed := []Status{"active", "done"}
	r := requireReq(t, "GET", "http://a/?s=active")
	v, ok := ParseEnumQuery(r, "s", allowed)
	assert.True(t, ok)
	assert.Equal(t, Status("active"), v)
	r = requireReq(t, "GET", "http://a/?s=invalid")
	_, ok = ParseEnumQuery(r, "s", allowed)
	assert.False(t, ok)
	r = requireReq(t, "GET", "http://a/")
	_, ok = ParseEnumQuery(r, "s", allowed)
	assert.False(t, ok)
}

func TestParseSortQuery(t *testing.T) {
	t.Parallel()
	allowed := []string{"score", "name", "created_at"}
	r := requireReq(t, "GET", "http://a/?sort=score")
	f, d, ok := ParseSortQuery(r, allowed)
	assert.True(t, ok)
	assert.Equal(t, "score", f)
	assert.Equal(t, "asc", d)
	r = requireReq(t, "GET", "http://a/?sort=-score")
	f, d, ok = ParseSortQuery(r, allowed)
	assert.True(t, ok)
	assert.Equal(t, "score", f)
	assert.Equal(t, "desc", d)
	r = requireReq(t, "GET", "http://a/?sort=unknown")
	_, _, ok = ParseSortQuery(r, allowed)
	assert.False(t, ok)
	r = requireReq(t, "GET", "http://a/")
	_, _, ok = ParseSortQuery(r, allowed)
	assert.False(t, ok)
	r = requireReq(t, "GET", "http://a/?sort=name:asc")
	f, d, ok = ParseSortQuery(r, allowed)
	assert.True(t, ok)
	assert.Equal(t, "name", f)
	assert.Equal(t, "asc", d)
	r = requireReq(t, "GET", "http://a/?sort=-name")
	f, d, ok = ParseSortQuery(r, allowed)
	assert.True(t, ok)
	assert.Equal(t, "name", f)
	assert.Equal(t, "desc", d)
}

func TestParseSortQuery_MixedNotationRejected(t *testing.T) {
	t.Parallel()
	allowed := []string{"name"}
	r := requireReq(t, "GET", "http://a/?sort=-name:asc")
	_, _, ok := ParseSortQuery(r, allowed)
	assert.False(t, ok, "mixed -prefix and :dir should be rejected")
	r = requireReq(t, "GET", "http://a/?sort=-name:desc")
	_, _, ok = ParseSortQuery(r, allowed)
	assert.False(t, ok, "mixed -prefix and :dir should be rejected")
}

func TestParseTimeQuery(t *testing.T) {
	t.Parallel()
	layout := time.RFC3339
	ts := "2025-03-15T12:00:00Z"
	r := requireReq(t, "GET", "http://a/?at="+ts)
	tm, ok := ParseTimeQuery(r, "at", layout)
	assert.True(t, ok)
	assert.True(t, tm.Equal(time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)))
	r = requireReq(t, "GET", "http://a/?at=invalid")
	_, ok = ParseTimeQuery(r, "at", layout)
	assert.False(t, ok)
}

func requireReq(t *testing.T, _, url string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	require.NoError(t, err)
	return r
}
