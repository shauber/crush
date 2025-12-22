package streaming

import (
	"encoding/json"
	"time"
)

// EventType represents different types of streaming events
type EventType string

const (
	// Connection events
	EventConnected    EventType = "connected"
	EventDisconnected EventType = "disconnected"
	EventHeartbeat    EventType = "heartbeat"

	// Message events
	EventMessageStart    EventType = "message_start"
	EventMessageProgress EventType = "message_progress"
	EventMessageEnd      EventType = "message_end"

	// Tool execution events
	EventToolStart    EventType = "tool_start"
	EventToolProgress EventType = "tool_progress"
	EventToolEnd      EventType = "tool_end"
	EventToolError    EventType = "tool_error"

	// Agent events
	EventAgentStart    EventType = "agent_start"
	EventAgentProgress EventType = "agent_progress"
	EventAgentEnd      EventType = "agent_end"
	EventAgentError    EventType = "agent_error"

	// Session events
	EventSessionCreated  EventType = "session_created"
	EventSessionUpdated  EventType = "session_updated"
	EventSessionDeleted  EventType = "session_deleted"
	EventSessionArchived EventType = "session_archived"
)

// StreamingEvent represents a single streaming event
type StreamingEvent struct {
	// EventType identifies the type of event
	Type EventType `json:"type"`
	
	// SessionID identifies which session this event belongs to
	SessionID string `json:"session_id"`
	
	// MessageID identifies which message this event is related to (if applicable)
	MessageID string `json:"message_id,omitempty"`
	
	// Timestamp when the event was created
	Timestamp time.Time `json:"timestamp"`
	
	// Sequence number for ordering events within a stream
	Sequence int64 `json:"sequence"`
	
	// Payload contains event-specific data
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ConnectionEventPayload for connection-related events
type ConnectionEventPayload struct {
	ClientID    string `json:"client_id,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	IP          string `json:"ip,omitempty"`
	Status      string `json:"status,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Connections int    `json:"connections,omitempty"`
}

// MessageEventPayload for message-related events
type MessageEventPayload struct {
	Content    string `json:"content,omitempty"`
	Delta      string `json:"delta,omitempty"`
	Role       string `json:"role,omitempty"`
	TokenCount int    `json:"token_count,omitempty"`
}

// ToolEventPayload for tool execution events
type ToolEventPayload struct {
	ToolName    string      `json:"tool_name,omitempty"`
	Arguments   interface{} `json:"arguments,omitempty"`
	Result      interface{} `json:"result,omitempty"`
	Status      string      `json:"status,omitempty"`
	Error       string      `json:"error,omitempty"`
	DurationMs  int64       `json:"duration_ms,omitempty"`
	OperationID string      `json:"operation_id,omitempty"`
}

// AgentEventPayload for agent execution events
type AgentEventPayload struct {
	AgentName   string      `json:"agent_name,omitempty"`
	Status      string      `json:"status,omitempty"`
	Thought     string      `json:"thought,omitempty"`
	Action      string      `json:"action,omitempty"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	DurationMs  int64       `json:"duration_ms,omitempty"`
	AgentID     string      `json:"agent_id,omitempty"`
}

// SessionEventPayload for session-related events
type SessionEventPayload struct {
	SessionID   string    `json:"session_id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Model       string    `json:"model,omitempty"`
	Provider    string    `json:"provider,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
}

// NewStreamingEvent creates a new streaming event with the current timestamp
func NewStreamingEvent(eventType EventType, sessionID string) *StreamingEvent {
	return &StreamingEvent{
		Type:      eventType,
		SessionID: sessionID,
		Timestamp: time.Now().UTC(),
	}
}

// WithMessageID sets the message ID for the event
func (e *StreamingEvent) WithMessageID(messageID string) *StreamingEvent {
	e.MessageID = messageID
	return e
}

// WithSequence sets the sequence number for the event
func (e *StreamingEvent) WithSequence(sequence int64) *StreamingEvent {
	e.Sequence = sequence
	return e
}

// WithPayload sets the payload for the event
func (e *StreamingEvent) WithPayload(payload interface{}) (*StreamingEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	e.Payload = data
	return e, nil
}