package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SSEHandler handles Server-Sent Events connections
// It manages client connections and sends events

type SSEHandler struct {
	clients    map[chan<- *StreamingEvent]struct{}
	addClient    chan chan<- *StreamingEvent
	removeClient chan chan<- *StreamingEvent
	broadcast    chan *StreamingEvent
	shutdown     chan struct{}
	isShutdown   bool
}

// NewSSEHandler creates a new SSE handler instance
func NewSSEHandler() *SSEHandler {
	h := &SSEHandler{
		clients:      make(map[chan<- *StreamingEvent]struct{}),
		addClient:    make(chan chan<- *StreamingEvent),
		removeClient: make(chan chan<- *StreamingEvent),
		broadcast:    make(chan *StreamingEvent),
		shutdown:     make(chan struct{}),
		isShutdown:   false,
	}
	go h.run()
	return h
}

// run manages the internal state of the SSE handler
func (h *SSEHandler) run() {
	for {
		select {
		case client := <-h.addClient:
			h.clients[client] = struct{}{}
			
		case client := <-h.removeClient:
			delete(h.clients, client)
			close(client)
			
		case event := <-h.broadcast:
			h.broadcastToClients(event)
			
		case <-h.shutdown:
			// Close all client connections
			for client := range h.clients {
				close(client)
			}
			h.clients = make(map[chan<- *StreamingEvent]struct{})
			h.isShutdown = true
			return
		}
	}
}

// broadcastToClients sends an event to all connected clients
func (h *SSEHandler) broadcastToClients(event *StreamingEvent) {
	for client := range h.clients {
		select {
		case client <- event:
			// Event sent successfully
		default:
			// Client's buffer is full, remove the client
			delete(h.clients, client)
			close(client)
		}
	}
}

// HandleSSE handles incoming SSE connections
func (h *SSEHandler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if handler is shutting down
	if h.isShutdown {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}

	// Parse session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("Access-Control-Allow-Origin", "*")
	headers.Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create a connected event
	connectedEvent := NewStreamingEvent(EventConnected, sessionID).
		WithPayload(ConnectionEventPayload{
			Status: "connected",
		})

	// Create a channel for this client
	client := make(chan *StreamingEvent, 100)
	
	// Register client
	h.addClient <- client
	
	// Ensure client is removed on exit
	defer func() {
		h.removeClient <- client
	}()

	// Create a context that cancels when the client disconnects
	ctx := r.Context()
	
	// Send the initial connected event
	eventData, err := json.Marshal(connectedEvent)
	if err != nil {
		http.Error(w, "Failed to marshal event", http.StatusInternalServerError)
		return
	}

	// Write the initial event
	writer := bufio.NewWriter(w)
	_, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", connectedEvent.Type, string(eventData))
	if err != nil {
		return
	}
	writer.Flush()

	// Flush the response to ensure headers are sent
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Heartbeat interval
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Create heartbeat event
	heartbeatEvent := NewStreamingEvent(EventHeartbeat, sessionID).
		WithPayload(map[string]interface{}{
			"status": "alive",
		})

	// Main streaming loop
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
			
		case event := <-client:
			if event == nil {
				// Channel closed
				return
			}
			
			// Marshal event to JSON
			eventData, err := json.Marshal(event)
			if err != nil {
				continue
			}
			
			// Write SSE format
			_, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", event.Type, string(eventData))
			if err != nil {
				return
			}
			writer.Flush()
			flusher.Flush()
			
		case <-heartbeat.C:
			// Send heartbeat
			eventData, err := json.Marshal(heartbeatEvent)
			if err != nil {
				continue
			}
			
			_, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", EventHeartbeat, string(eventData))
			if err != nil {
				return
			}
			writer.Flush()
			flusher.Flush()
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *SSEHandler) Broadcast(event *StreamingEvent) {
	select {
	case h.broadcast <- event:
		// Event queued for broadcast
	case <-time.After(1 * time.Second):
		// Handler might be shutting down, ignore
	}
}

// BroadcastToSession sends an event only to clients subscribed to a specific session
// This is a basic implementation - in future we might filter by session_id in the broadcast channel
func (h *SSEHandler) BroadcastToSession(event *StreamingEvent) {
	h.Broadcast(event)
}

// GetConnectionCount returns the number of active connections
func (h *SSEHandler) GetConnectionCount() int {
	count := 0
	temp := make(chan int, 1)
	
	go func() {
		h.addClient <- make(chan *StreamingEvent) // Add dummy client to enter select
		defer func() { h.removeClient <- make(chan *StreamingEvent) }()
		
		count = 0
		for range h.clients {
			count++
		}
		temp <- count
	}()
	
	select {
	case result := <-temp:
		return result
	case <-time.After(100 * time.Millisecond):
		return -1 // Unable to get count
	}
}

// Shutdown gracefully shuts down the SSE handler
func (h *SSEHandler) Shutdown() {
	select {
	case h.shutdown <- struct{}{}:
		// Shutdown initiated
	case <-time.After(5 * time.Second):
		// Don't block indefinitely
	}
}