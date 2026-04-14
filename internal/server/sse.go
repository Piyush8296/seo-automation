package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Hub manages Server-Sent Event channels keyed by audit ID.
// Each audit can have multiple concurrent browser subscribers (e.g. multiple tabs).
type Hub struct {
	mu      sync.RWMutex
	clients map[string][]chan ProgressEvent
}

// NewHub constructs an empty Hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[string][]chan ProgressEvent)}
}

// Subscribe registers a new channel for the given audit ID.
func (h *Hub) Subscribe(auditID string) chan ProgressEvent {
	ch := make(chan ProgressEvent, 64) // buffered so Broadcast never blocks
	h.mu.Lock()
	h.clients[auditID] = append(h.clients[auditID], ch)
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes and closes the channel.
func (h *Hub) Unsubscribe(auditID string, ch chan ProgressEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.clients[auditID]
	for i, c := range subs {
		if c == ch {
			h.clients[auditID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			return
		}
	}
}

// Broadcast sends an event to all subscribers of the given audit.
// It never blocks: slow clients are skipped (their buffered channel absorbs spikes).
func (h *Hub) Broadcast(auditID string, event ProgressEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.clients[auditID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// ServeSSE upgrades the HTTP response to an SSE stream and forwards events
// until the audit finishes or the client disconnects.
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request, auditID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported by this server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	ch := h.Subscribe(auditID)
	defer h.Unsubscribe(auditID, ch)

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

			if event.Type == "complete" || event.Type == "error" || event.Type == "cancelled" {
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}
