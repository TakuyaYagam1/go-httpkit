package middleware

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

type statusWriter struct {
	http.ResponseWriter
	mu           sync.Mutex
	status       int
	bytesWritten int
	headerSent   bool
}

func (w *statusWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.headerSent {
		return
	}
	w.headerSent = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	if !w.headerSent {
		w.headerSent = true
		if w.status == 0 {
			w.status = http.StatusOK
		}
		w.ResponseWriter.WriteHeader(w.status)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	w.mu.Unlock()
	return n, err
}

func (w *statusWriter) Status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *statusWriter) BytesWritten() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.bytesWritten
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) ReadFrom(r io.Reader) (int64, error) {
	w.mu.Lock()
	if !w.headerSent {
		w.headerSent = true
		if w.status == 0 {
			w.status = http.StatusOK
		}
	}
	w.mu.Unlock()
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(r)
		w.mu.Lock()
		w.bytesWritten += int(n)
		w.mu.Unlock()
		return n, err
	}
	n, err := io.Copy(w.ResponseWriter, r)
	w.mu.Lock()
	w.bytesWritten += int(n)
	w.mu.Unlock()
	return n, err
}

func (w *statusWriter) claimHeaderSent() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.headerSent {
		return false
	}
	w.headerSent = true
	return true
}
