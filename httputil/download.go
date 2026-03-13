package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func sanitizeContentDispositionFilename(name string) string {
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		if r < 32 || r == 127 || r == '"' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	s := b.String()
	if s == "" {
		return "download"
	}
	return s
}

// RenderJSONAttachment encodes data as JSON and writes it as a downloadable attachment.
func RenderJSONAttachment[T any](w http.ResponseWriter, data T, filename string) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return fmt.Errorf("encode json attachment: %w", err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": sanitizeContentDispositionFilename(filename)}))
	_, err := w.Write(buf.Bytes())
	return err
}

// RenderStream writes a streaming response with Content-Disposition attachment. Caller must close rc.
func RenderStream(w http.ResponseWriter, contentType, filename string, rc io.Reader) error {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": sanitizeContentDispositionFilename(filename)}))
	_, err := io.Copy(w, rc)
	return err
}

// RenderBytes writes raw bytes with Content-Disposition attachment.
func RenderBytes(w http.ResponseWriter, contentType, filename string, data []byte) error {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": sanitizeContentDispositionFilename(filename)}))
	_, err := w.Write(data)
	return err
}
