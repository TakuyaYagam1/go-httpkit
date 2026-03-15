package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var ErrSSEClosed = errors.New("SSE writer closed")

// SSEWriter sends Server-Sent Events. Use NewSSEWriter to create; it returns (nil, false) if w does not implement http.Flusher.
type SSEWriter struct {
	w    http.ResponseWriter
	done bool
}

// NewSSEWriter configures w for SSE (Content-Type, Cache-Control, etc.) and returns a writer. Second return is false if w is not flushable.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	f.Flush()
	return &SSEWriter{w: w}, true
}

// Send writes an SSE message: optional "event: <event>\n" and "data: <line>\n" per line of data, then "\n". Flushes after write.
func (s *SSEWriter) Send(event, data string) error {
	if s.done {
		return ErrSSEClosed
	}
	if event != "" {
		_, err := s.w.Write([]byte("event: " + event + "\n"))
		if err != nil {
			return err
		}
	}
	for _, line := range strings.Split(data, "\n") {
		_, err := s.w.Write([]byte("data: " + line + "\n"))
		if err != nil {
			return err
		}
	}
	_, err := s.w.Write([]byte("\n"))
	if err != nil {
		return err
	}
	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// SendJSON marshals v to JSON and sends it as the data payload via Send.
func (s *SSEWriter) SendJSON(event string, v any) error {
	if s.done {
		return ErrSSEClosed
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Send(event, string(raw))
}

// Close marks the writer as done; subsequent Send/SendJSON calls are no-ops.
func (s *SSEWriter) Close() {
	s.done = true
}
