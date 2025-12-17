package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/crush/internal/pubsub"
)

func (h *handlers) streamSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	// Verify session exists
	_, err := h.app.Sessions.Get(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to message events
	messageEvents := h.app.Messages.Subscribe(r.Context())

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"session_id\":\"%s\"}\n\n", sessionID)
	flusher.Flush()

	// Stream events
	for {
		select {
		case <-r.Context().Done():
			return

		case event := <-messageEvents:
			msg := event.Payload

			// Only send events for this session
			if msg.SessionID != sessionID {
				continue
			}

			eventType := "message_updated"
			switch event.Type {
			case pubsub.CreatedEvent:
				eventType = "message_created"
			case pubsub.DeletedEvent:
				eventType = "message_deleted"
			}

			data, _ := json.Marshal(map[string]interface{}{
				"type":       eventType,
				"message_id": msg.ID,
				"role":       msg.Role,
				"content":    msg.Content().String(),
				"created_at": msg.CreatedAt,
			})

			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
			flusher.Flush()
		}
	}
}
