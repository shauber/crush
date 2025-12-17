package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/message"
)

type handlers struct {
	app *app.App
}

// Response wrapper types for proper JSON serialization
type sessionResponse struct {
	ID               string `json:"id"`
	ParentSessionID  string `json:"parent_session_id,omitempty"`
	Title            string `json:"title"`
	MessageCount     int64  `json:"message_count"`
	PromptTokens     int64  `json:"prompt_tokens,omitempty"`
	CompletionTokens int64  `json:"completion_tokens,omitempty"`
	SummaryMessageID string `json:"summary_message_id,omitempty"`
	Cost             float64 `json:"cost,omitempty"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

type messageResponse struct {
	ID               string `json:"id"`
	SessionID        string `json:"session_id"`
	Role             string `json:"role"`
	Content          string `json:"content"`
	Model            string `json:"model,omitempty"`
	Provider         string `json:"provider,omitempty"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
	IsSummaryMessage bool   `json:"is_summary_message"`
}

type createSessionRequest struct {
	Title string `json:"title"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

func toMessageResponse(msg message.Message) messageResponse {
	return messageResponse{
		ID:               msg.ID,
		SessionID:        msg.SessionID,
		Role:             string(msg.Role),
		Content:          msg.Content().String(),
		Model:            msg.Model,
		Provider:         msg.Provider,
		CreatedAt:        msg.CreatedAt,
		UpdatedAt:        msg.UpdatedAt,
		IsSummaryMessage: msg.IsSummaryMessage,
	}
}

func (h *handlers) createSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	title := req.Title
	if title == "" {
		title = "New Session"
	}

	sess, err := h.app.Sessions.Create(r.Context(), title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := sessionResponse{
		ID:               sess.ID,
		ParentSessionID:  sess.ParentSessionID,
		Title:            sess.Title,
		MessageCount:     sess.MessageCount,
		PromptTokens:     sess.PromptTokens,
		CompletionTokens: sess.CompletionTokens,
		SummaryMessageID: sess.SummaryMessageID,
		Cost:             sess.Cost,
		CreatedAt:        sess.CreatedAt,
		UpdatedAt:        sess.UpdatedAt,
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *handlers) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.app.Sessions.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]sessionResponse, len(sessions))
	for i, sess := range sessions {
		resp[i] = sessionResponse{
			ID:               sess.ID,
			ParentSessionID:  sess.ParentSessionID,
			Title:            sess.Title,
			MessageCount:     sess.MessageCount,
			PromptTokens:     sess.PromptTokens,
			CompletionTokens: sess.CompletionTokens,
			SummaryMessageID: sess.SummaryMessageID,
			Cost:             sess.Cost,
			CreatedAt:        sess.CreatedAt,
			UpdatedAt:        sess.UpdatedAt,
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *handlers) getSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sess, err := h.app.Sessions.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	resp := sessionResponse{
		ID:               sess.ID,
		ParentSessionID:  sess.ParentSessionID,
		Title:            sess.Title,
		MessageCount:     sess.MessageCount,
		PromptTokens:     sess.PromptTokens,
		CompletionTokens: sess.CompletionTokens,
		SummaryMessageID: sess.SummaryMessageID,
		Cost:             sess.Cost,
		CreatedAt:        sess.CreatedAt,
		UpdatedAt:        sess.UpdatedAt,
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *handlers) deleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.app.Sessions.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) sendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create user message
	msg, err := h.app.Messages.Create(r.Context(), sessionID, message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: req.Content}},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start agent processing (async)
	go func() {
		_, err := h.app.AgentCoordinator.Run(context.Background(), sessionID, req.Content)
		if err != nil {
			slog.Error("agent run failed", "error", err, "session", sessionID)
		}
	}()

	respondJSON(w, http.StatusOK, toMessageResponse(msg))
}

func (h *handlers) listMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	messages, err := h.app.Messages.List(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]messageResponse, len(messages))
	for i, msg := range messages {
		resp[i] = toMessageResponse(msg)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *handlers) cancelSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	h.app.AgentCoordinator.Cancel(sessionID)

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) sessionStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	status := map[string]interface{}{
		"busy":           h.app.AgentCoordinator.IsSessionBusy(sessionID),
		"queued_prompts": h.app.AgentCoordinator.QueuedPrompts(sessionID),
	}

	respondJSON(w, http.StatusOK, status)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
