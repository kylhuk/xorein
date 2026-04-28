package control

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const sseKeepalive = 30 * time.Second

// Event is an SSE event with a type name and arbitrary JSON data.
type Event struct {
	Type string
	Data any
}

// Mux fans out Events to all connected SSE subscribers.
type Mux struct {
	mu   sync.Mutex
	subs []chan Event
}

// Publish sends an event to all current subscribers (non-blocking; drops if slow).
func (m *Mux) Publish(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func (m *Mux) subscribe() (chan Event, func()) {
	ch := make(chan Event, 64)
	m.mu.Lock()
	m.subs = append(m.subs, ch)
	m.mu.Unlock()
	return ch, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		for i, s := range m.subs {
			if s == ch {
				m.subs = append(m.subs[:i], m.subs[i+1:]...)
				close(ch)
				return
			}
		}
	}
}

// ServeHTTP implements http.Handler for GET /v1/events.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, CodeUnsupported, "SSE not supported by transport")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsub := m.subscribe()
	defer unsub()

	// Spec 60 §5.2: send ready event immediately after connection.
	fmt.Fprintf(w, "event: ready\ndata: {\"version\":\"1\"}\n\n")
	f.Flush()

	tick := time.NewTicker(sseKeepalive)
	defer tick.Stop()

	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(e.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, data)
			f.Flush()
		case <-tick.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			f.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
