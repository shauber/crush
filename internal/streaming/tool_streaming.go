package streaming

import (
	"context"
	"log/slog"
	"time"

	"charm.land/fantasy"

	apistreaming "github.com/charmbracelet/crush/internal/api/streaming"
)

// sessionIDContextKey matches the key used in agent/tools package.
type sessionIDContextKey string

const sessionIDKey sessionIDContextKey = "session_id"

// getSessionFromContext retrieves the session ID from the context.
func getSessionFromContext(ctx context.Context) string {
	sessionID := ctx.Value(sessionIDKey)
	if sessionID == nil {
		return ""
	}
	s, ok := sessionID.(string)
	if !ok {
		return ""
	}
	return s
}

// toolStreaming wraps fantasy.AgentTool with streaming capabilities.
type toolStreaming struct {
	fantasy.AgentTool
	sseHandler *apistreaming.SSEHandler
}

// NewToolStreaming creates a streaming-enabled wrapper for an AgentTool.
func NewToolStreaming(base fantasy.AgentTool, sseHandler *apistreaming.SSEHandler) fantasy.AgentTool {
	return &toolStreaming{
		AgentTool:  base,
		sseHandler: sseHandler,
	}
}

// getActiveHandler returns the SSE handler to use - either the explicit one
// or the global handler if available.
func (ts *toolStreaming) getActiveHandler() *apistreaming.SSEHandler {
	if ts.sseHandler != nil {
		return ts.sseHandler
	}
	return apistreaming.GlobalSSEHandler
}

// Run implements AgentTool interface with streaming.
func (ts *toolStreaming) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	handler := ts.getActiveHandler()
	if handler == nil {
		return ts.AgentTool.Run(ctx, call)
	}

	// Get session ID from context using the proper context key.
	sessionID := getSessionFromContext(ctx)
	if sessionID == "" {
		return ts.AgentTool.Run(ctx, call)
	}

	info := ts.AgentTool.Info()

	// Stream tool start
	ts.streamToolStart(sessionID, info.Name, call.Input)

	// Execute the tool via type assertion with fantasy standards
	start := time.Now()
	result, err := ts.AgentTool.Run(ctx, call)

	// Stream result
	duration := time.Since(start)
	if err != nil {
		ts.streamToolError(sessionID, info.Name, err.Error(), duration)
	} else {
		ts.streamToolComplete(sessionID, info.Name, result, duration)
	}

	return result, err
}

func (ts *toolStreaming) streamToolStart(sessionID, toolName string, params string) {
	handler := ts.getActiveHandler()
	if handler == nil {
		return
	}

	payload := map[string]interface{}{
		"tool_name": toolName,
		"arguments": params,
		"status":    "started",
	}

	event, err := apistreaming.NewStreamingEvent(apistreaming.EventToolStart, sessionID).
		WithPayload(payload)
	if err != nil {
		slog.Error("failed to create streaming event", "error", err)
		return
	}

	handler.Broadcast(event)
}

func (ts *toolStreaming) streamToolComplete(sessionID, toolName string, result interface{}, duration time.Duration) {
	handler := ts.getActiveHandler()
	if handler == nil {
		return
	}

	payload := map[string]interface{}{
		"tool_name":   toolName,
		"result":      result,
		"status":      "completed",
		"duration_ms": duration.Milliseconds(),
	}

	event, err := apistreaming.NewStreamingEvent(apistreaming.EventToolEnd, sessionID).
		WithPayload(payload)
	if err != nil {
		slog.Error("failed to create streaming event", "error", err)
		return
	}

	handler.Broadcast(event)
}

func (ts *toolStreaming) streamToolError(sessionID, toolName string, errorText string, duration time.Duration) {
	handler := ts.getActiveHandler()
	if handler == nil {
		return
	}

	payload := map[string]interface{}{
		"tool_name":   toolName,
		"error":       errorText,
		"status":      "error",
		"duration_ms": duration.Milliseconds(),
	}

	event, err := apistreaming.NewStreamingEvent(apistreaming.EventToolError, sessionID).
		WithPayload(payload)
	if err != nil {
		slog.Error("failed to create streaming event", "error", err)
		return
	}

	handler.Broadcast(event)
}

// WrapTools applies streaming to a slice of tools using an explicit handler.
func WrapTools(tools []fantasy.AgentTool, handler *apistreaming.SSEHandler) []fantasy.AgentTool {
	if handler == nil {
		return tools
	}

	wrapped := make([]fantasy.AgentTool, len(tools))
	for i, tool := range tools {
		wrapped[i] = NewToolStreaming(tool, handler)
	}
	return wrapped
}

// WrapToolsWithGlobal applies streaming to a slice of tools using the global
// SSE handler. This is useful when tools need to automatically stream to
// clients when the server is running, without requiring explicit handler
// passing.
func WrapToolsWithGlobal(toolList []fantasy.AgentTool) []fantasy.AgentTool {
	wrapped := make([]fantasy.AgentTool, len(toolList))
	for i, tool := range toolList {
		wrapped[i] = NewToolStreaming(tool, nil) // Will use GlobalSSEHandler
	}
	return wrapped
}

// Required interface methods
func (ts *toolStreaming) Info() fantasy.ToolInfo {
	return ts.AgentTool.Info()
}

func (ts *toolStreaming) ProviderOptions() fantasy.ProviderOptions {
	return ts.AgentTool.ProviderOptions()
}

func (ts *toolStreaming) SetProviderOptions(opts fantasy.ProviderOptions) {
	ts.AgentTool.SetProviderOptions(opts)
}
