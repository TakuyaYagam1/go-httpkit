package httputil

import (
	"net/http"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		remoteAddr         string
		headers            map[string]string
		trustedProxyCIDRs  []string
		want               string
	}{
		{"no proxy", "192.168.1.1:12345", nil, nil, "192.168.1.1"},
		{"no trusted CIDRs ignores headers", "192.168.1.1:12345", map[string]string{"X-Real-IP": "10.0.0.1"}, nil, "192.168.1.1"},
		{"trusted proxy X-Real-IP", "10.0.0.2:80", map[string]string{"X-Real-IP": "203.0.113.1"}, []string{"10.0.0.0/8"}, "203.0.113.1"},
		{"trusted proxy X-Forwarded-For", "10.0.0.2:80", map[string]string{"X-Forwarded-For": "203.0.113.2"}, []string{"10.0.0.0/8"}, "203.0.113.2"},
		{"X-Real-IP preferred over X-Forwarded-For", "10.0.0.2:80", map[string]string{"X-Real-IP": "1.2.3.4", "X-Forwarded-For": "5.6.7.8"}, []string{"10.0.0.0/8"}, "1.2.3.4"},
		{"untrusted proxy uses remote", "192.168.1.1:80", map[string]string{"X-Real-IP": "10.0.0.1"}, []string{"10.0.0.0/8"}, "192.168.1.1"},
		{"first of X-Forwarded-For", "10.0.0.2:80", map[string]string{"X-Forwarded-For": " 203.0.113.3 , 10.0.0.1 "}, []string{"10.0.0.0/8"}, "203.0.113.3"},
	}
	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = tt.remoteAddr
		for k, v := range tt.headers {
			r.Header.Set(k, v)
		}
		got := GetClientIP(r, tt.trustedProxyCIDRs)
		if got != tt.want {
			t.Errorf("%s: GetClientIP() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
