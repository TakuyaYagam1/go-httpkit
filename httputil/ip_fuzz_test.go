package httputil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func FuzzGetClientIPWithNets(f *testing.F) {
	f.Add("10.0.0.2:1234", "203.0.113.1", "198.51.100.1, 10.0.0.1", "10.0.0.0/8")
	f.Add("192.168.1.10:443", "bad-ip", "1.2.3.4", "192.168.0.0/16")
	f.Add("203.0.113.9:80", "", "", "10.0.0.0/8")
	f.Fuzz(func(_ *testing.T, remoteAddr, xRealIP, xForwardedFor, extraCIDR string) {
		remoteAddr = limitFuzzString(remoteAddr, 256)
		xRealIP = limitFuzzString(xRealIP, 256)
		xForwardedFor = limitFuzzString(xForwardedFor, 512)
		extraCIDR = strings.TrimSpace(limitFuzzString(extraCIDR, 128))

		nets, err := ParseTrustedProxyCIDRs([]string{testCIDR10, "192.168.0.0/16", extraCIDR})
		if err != nil && nets == nil {
			nets = nil
		}
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r.RemoteAddr = remoteAddr
		r.Header.Set(headerXRealIP, xRealIP)
		r.Header.Set(headerXForwardedFor, xForwardedFor)

		_ = GetClientIPWithNets(r, nets)
	})
}

func limitFuzzString(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}
