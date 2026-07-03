package httputil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func FuzzDecodeJSON(f *testing.F) {
	f.Add([]byte(`{"name":"alice","age":30}`))
	f.Add([]byte(`{"name":"bob"} garbage`))
	f.Add([]byte(`not-json`))
	f.Add([]byte(`{"nested":{"x":[1,2,3]}}`))
	f.Fuzz(func(_ *testing.T, data []byte) {
		if len(data) > MaxRequestBodySize {
			data = data[:MaxRequestBodySize]
		}
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
		var out map[string]any
		_ = DecodeJSON(r, &out)
	})
}
