package streaming

import (
	"context"
	"log/slog"
	"sync"

	apistreaming "github.com/charmbracelet/crush/internal/api/streaming"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/pubsub"
)

// MessageSubscriber listens to message events and emits SSE tool events.
type MessageSubscriber struct {
	messages   message.Service
	sseHandler *apistreaming.SSEHandler

	// Track tool call states to detect transitions.
	mu             sync.RWMutex
	toolCallStates map[string]toolCallState
}

type toolCallState struct {
	started  bool
	finished bool
}

// NewMessageSubscriber creates a subscriber that bridges message pubsub to SSE.
func NewMessageSubscriber(messages message.Service, handler *apistreaming.SSEHandler) *MessageSubscriber {
	return &MessageSubscriber{
		messages:       messages,
		sseHandler:     handler,
		toolCallStates: make(map[string]toolCallState),
	}
}

// Start begins listening to message events and emitting SSE events.
// This should be called in a goroutine.
func (ms *MessageSubscriber) Start(ctx context.Context) {
	events := ms.messages.Subscribe(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			ms.handleMessageEvent(event)
		}
	}
}

func (ms *MessageSubscriber) handleMessageEvent(event pubsub.Event[message.Message]) {
	msg := event.Payload

	// Only process assistant messages (which contain tool calls).
	if msg.Role != message.Assistant {
		// Check for tool results in tool messages.
		if msg.Role == message.Tool {
			ms.handleToolResultMessage(msg)
		}
		return
	}

	// Process tool calls in the message.
	for _, tc := range msg.ToolCalls() {
		ms.processToolCall(msg.SessionID, tc)
	}
}

func (ms *MessageSubscriber) processToolCall(sessionID string, tc message.ToolCall) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	state, exists := ms.toolCallStates[tc.ID]

	// Tool just started (first time we see it).
	if !exists && !tc.Finished {
		ms.toolCallStates[tc.ID] = toolCallState{started: true, finished: false}
		ms.emitToolStart(sessionID, tc)
		return
	}

	// Tool just finished (transition from started to finished).
	if exists && state.started && !state.finished && tc.Finished {
		ms.toolCallStates[tc.ID] = toolCallState{started: true, finished: true}
		ms.emitToolEnd(sessionID, tc)
		return
	}

	// Tool appeared already finished (quick execution).
	if !exists && tc.Finished {
		ms.toolCallStates[tc.ID] = toolCallState{started: true, finished: true}
		ms.emitToolStart(sessionID, tc)
		ms.emitToolEnd(sessionID, tc)
	}
}

func (ms *MessageSubscriber) handleToolResultMessage(msg message.Message) {
	for _, tr := range msg.ToolResults() {
		ms.emitToolResult(msg.SessionID, tr)
	}
}

func (ms *MessageSubscriber) emitToolStart(sessionID string, tc message.ToolCall) {
	if ms.sseHandler == nil {
		return
	}

	payload := map[string]interface{}{
		"tool_call_id": tc.ID,
		"tool_name":    tc.Name,
		"arguments":    tc.Input,
		"status":       "started",
	}

	event, err := apistreaming.NewStreamingEvent(apistreaming.EventToolStart, sessionID).
		WithPayload(payload)
	if err != nil {
		slog.Error("failed to create tool start event", "error", err)
		return
	}

	ms.sseHandler.Broadcast(event)
}

func (ms *MessageSubscriber) emitToolEnd(sessionID string, tc message.ToolCall) {
	if ms.sseHandler == nil {
		return
	}

	payload := map[string]interface{}{
		"tool_call_id": tc.ID,
		"tool_name":    tc.Name,
		"status":       "completed",
	}

	event, err := apistreaming.NewStreamingEvent(apistreaming.EventToolEnd, sessionID).
		WithPayload(payload)
	if err != nil {
		slog.Error("failed to create tool end event", "error", err)
		return
	}

	ms.sseHandler.Broadcast(event)
}

func (ms *MessageSubscriber) emitToolResult(sessionID string, tr message.ToolResult) {
	if ms.sseHandler == nil {
		return
	}

	status := "completed"
	if tr.IsError {
		status = "error"
	}

	payload := map[string]interface{}{
		"tool_call_id": tr.ToolCallID,
		"tool_name":    tr.Name,
		"status":       status,
		"is_error":     tr.IsError,
		"has_content":  tr.Content != "",
	}

	eventType := apistreaming.EventToolEnd
	if tr.IsError {
		eventType = apistreaming.EventToolError
	}

	event, err := apistreaming.NewStreamingEvent(eventType, sessionID).
		WithPayload(payload)
	if err != nil {
		slog.Error("failed to create tool result event", "error", err)
		return
	}

	ms.sseHandler.Broadcast(event)
}

// Cleanup removes old tool call states to prevent memory leaks.
// Call this periodically or when sessions end.
func (ms *MessageSubscriber) Cleanup(toolCallIDs ...string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, id := range toolCallIDs {
		delete(ms.toolCallStates, id)
	}
}

// CleanupSession removes all tool call states (call periodically to prevent
// memory leaks).
func (ms *MessageSubscriber) CleanupAll() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.toolCallStates = make(map[string]toolCallState)
}
