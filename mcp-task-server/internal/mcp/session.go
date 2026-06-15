package mcp

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// Session represents one connected MCP client.
// Each client gets its own session; the session owns the SSE writer.
type Session struct {
	ID string

	w       http.ResponseWriter
	flusher http.Flusher
	mu      sync.Mutex
	closed  atomic.Bool
}

func NewSession(w http.ResponseWriter) (*Session, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support flushing")
	}
	return &Session{
		ID:      uuid.NewString(),
		w:       w,
		flusher: flusher,
	}, nil
}

// SendEvent writes a single SSE event to the client.
// Safe to call from multiple goroutines.
func (s *Session) SendEvent(eventType string, data []byte) error {
	if s.closed.Load() {
		return fmt.Errorf("session closed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// SSE wire format:
	//   event: <type>\n
	//   data: <json>\n
	//   \n
	_, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventType, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *Session) Close() {
	s.closed.Store(true)
}
