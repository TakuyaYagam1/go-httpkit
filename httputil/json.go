package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/wahrwelt-kit/go-httpkit/httperr"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(data); err != nil {
		writeInternalJSONError(w)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func writeInternalJSONError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", mimeApplicationJSON)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"code":"` + httperr.CodeInternalError + `","message":"` + msgInternalServerError + `"}` + "\n"))
}
