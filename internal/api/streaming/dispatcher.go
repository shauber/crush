package streaming

import (
	"sync"
	"sync/atomic"
)

// Dispatcher manages event distribution across SSE handlers and other consumers
type Dispatcher struct {
	sseHandler *SSEHandler
	
	// Mapping to track active session streams
	sessionStreams map[string]*sessionStream
	t streamsMutex sync.RWMutex
	
	// Global sequence counter for event ordering
	sequenceCounter int64
}

// sessionStream represents a single session's event stream
type sessionStream struct {
	sessionID string
	lastEvent int64
	events   map[string]*StreamingEvent // MessageID -> latest event
	mutex   sync.RWMutex
}

// EventDispatcher provides public interface for event emission
type EventDispatcher interface {
	// Emit sends an event to all connected clients for the session
	Emit(event *StreamingEvent)
	
	// EmitToSession sends an event to a specific session
	EmitToSession(sessionID string, event *StreamingEvent)
	
	// EmitWithSource sends an event with source information
	EmitWithSource(event *StreamingEvent, source string)
	
	// GetLastEvent returns the last event for a message in a session
	GetLastEvent(sessionID, messageID string) *StreamingEvent
	
	// GetStreamCount returns the number of active stream connections
	GetStreamCount() int
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		sseHandler:      NewSSEHandler(),
		sessionStreams:  make(map[string]*sessionStream),
		sequenceCounter: 0,
	}
}

// NewDispatcherWithSSE creates a new dispatcher with existing SSE handler
func NewDispatcherWithSSE(sseHandler *SSEHandler) *Dispatcher {
	return &Dispatcher{
		sseHandler:      sseHandler,
		sessionStreams:  make(map[string]*sessionStream),
		sequenceCounter: 0,
	}
}

// getNextSequence returns the next sequence number
func (d *Dispatcher) getNextSequence() int64 {
	return atomic.AddInt64(&d.sequenceCounter, 1)
}

// getOrCreateSessionStream gets or creates a session stream
func (d *Dispatcher) getOrCreateSessionStream(sessionID string) *sessionStream {
	d.streamsMutex.RLock()
	if stream, exists := d.sessionStreams[sessionID]; exists {
		d.streamsMutex.RUnlock()
		return stream
	}
	d.streamsMutex.RUnlock()

	d.streamsMutex.Lock()
	defer d.streamsMutex.Unlock()
	
	// Double-check after lock
	if stream, exists := d.sessionStreams[sessionID]; exists {
		return stream
	}
	
	stream := &sessionStream{
		sessionID: sessionID,
		events:    make(map[string]*StreamingEvent),
	}
	d.sessionStreams[sessionID] = stream
	return stream
}

// Emit sends an event to the global stream (deprecated - use EmitToSession)
func (d *Dispatcher) Emit(event *StreamingEvent) {
	d.EmitToSession(event.SessionID, event)
}

// EmitToSession sends an event to a specific session
func (d *Dispatcher) EmitToSession(sessionID string, event *StreamingEvent) {
	if event.SessionID == "" {
		event.SessionID = sessionID
	}
	
	// Ensure sequence is set
	if event.Sequence == 0 {
		event.Sequence = d.getNextSequence()
	}
	
	// Store event for the session
	stream := d.getOrCreateSessionStream(sessionID)
	stream.addEvent(event)
	
	// Broadcast via SSE
	d.sseHandler.Broadcast(event)
}

// EmitWithSource sends an event with additional source information
// 'source' can be used for debugging or filtering purposes
func (d *Dispatcher) EmitWithSource(event *StreamingEvent, source string) {
	// You could log or tag the event with source here
	// For now, just emit as normal
	d.EmitToSession(event.SessionID, event)
}

// addEvent adds an event to the session stream
func (s *sessionStream) addEvent(event *StreamingEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.lastEvent = event.Sequence
	
	// Store by message ID if provided
	if event.MessageID != "" {
		s.events[event.MessageID] = event
	}
}

// GetLastEvent returns the last event for a specific message in a session
func (d *Dispatcher) GetLastEvent(sessionID, messageID string) *StreamingEvent {
	d.streamsMutex.RLock()
	stream, exists := d.sessionStreams[sessionID]
	d.streamsMutex.RUnlock()
	
	if !exists {
		return nil
	}
	
	return stream.getEvent(messageID)
}

// getEvent gets a specific event from the session stream
func (s *sessionStream) getEvent(messageID string) *StreamingEvent {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.events[messageID]
}

// GetStreamCount returns the total number of active SSE connections
func (d *Dispatcher) GetStreamCount() int {
	return d.sseHandler.GetConnectionCount()
}

// GetSessionCount returns the number of active session streams
func (d *Dispatcher) GetSessionCount() int {
	d.streamsMutex.RLock()
	defer d.streamsMutex.RUnlock()
	return len(d.sessionStreams)
}

// CleanupSession removes a session stream and its events
func (d *Dispatcher) CleanupSession(sessionID string) {
	d.streamsMutex.Lock()
	defer d.streamsMutex.Unlock()
	
	if stream, exists := d.sessionStreams[sessionID]; exists {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		
		// Clear all events
		for k := range stream.events {
			delete(stream.events, k)
		}
		
		delete(d.sessionStreams, sessionID)
	}
}

// GetSSEHandler returns the underlying SSE handler for direct SSE access
func (d *Dispatcher) GetSSEHandler() *SSEHandler {
	return d.sseHandler
}

// Shutdown gracefully shuts down the dispatcher
func (d *Dispatcher) Shutdown() {
	d.sseHandler.Shutdown()
	
	d.streamsMutex.Lock()
	defer d.streamsMutex.Unlock()
	
	// Clear all session streams
	for sessionID := range d.sessionStreams {
		if stream := d.sessionStreams[sessionID]; stream != nil {
			stream.mutex.Lock()
			for k := range stream.events {
				delete(stream.events, k)
			}
			stream.mutex.Unlock()
		}
	}
	
	// Clear the map
	for k := range d.sessionStreams {
		delete(d.sessionStreams, k)
	}
}

// HealthCheck checks if the dispatcher is running properly
func (d *Dispatcher) HealthCheck() map[string]interface{} {
	d.streamsMutex.RLock()
	sessionCount := len(d.sessionStreams)
	d.streamsMutex.RUnlock()
	
	connectionCount := d.GetStreamCount()
	
	return map[string]interface{}{
		"session_streams":   sessionCount,
		"sse_connections":   connectionCount,
		"total_events_sent": atomic.LoadInt64(&d.sequenceCounter),
		"status":           "healthy",
		"sse_handler":      "running",
	}
}