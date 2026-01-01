package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/message"
)

type handlers struct {
	app *app.App
}

// Response wrapper types for proper JSON serialization
type sessionResponse struct {
	ID               string  `json:"id"`
	ParentSessionID  string  `json:"parent_session_id,omitempty"`
	Title            string  `json:"title"`
	MessageCount     int64   `json:"message_count"`
	PromptTokens     int64   `json:"prompt_tokens,omitempty"`
	CompletionTokens int64   `json:"completion_tokens,omitempty"`
	SummaryMessageID string  `json:"summary_message_id,omitempty"`
	Cost             float64 `json:"cost,omitempty"`
	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
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

type setModelRequest struct {
	ModelType string `json:"model_type"` // "large" or "small"
	Model     string `json:"model"`      // model ID
	Provider  string `json:"provider"`   // provider ID
}

func toMessageResponse(msg message.Message) messageResponse {
	content := getMessageContent(msg)
	return messageResponse{
		ID:               msg.ID,
		SessionID:        msg.SessionID,
		Role:             string(msg.Role),
		Content:          content,
		Model:            msg.Model,
		Provider:         msg.Provider,
		CreatedAt:        msg.CreatedAt,
		UpdatedAt:        msg.UpdatedAt,
		IsSummaryMessage: msg.IsSummaryMessage,
	}
}

// getMessageContent extracts the appropriate content from a message based on its role.
// For tool role messages, it serializes the tool results. For all other roles, it returns text content.
func getMessageContent(msg message.Message) string {
	if msg.Role == message.Tool {
		// For tool role messages, serialize tool results as JSON
		toolResults := msg.ToolResults()
		if len(toolResults) == 0 {
			return ""
		}
		// If there's only one tool result, return its content directly
		if len(toolResults) == 1 {
			return toolResults[0].Content
		}
		// For multiple tool results, serialize as JSON array
		data, err := json.Marshal(toolResults)
		if err != nil {
			return ""
		}
		return string(data)
	}
	// For all other roles, return text content
	return msg.Content().String()
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

func (h *handlers) listModels(w http.ResponseWriter, r *http.Request) {
	cfg := h.app.Config()
	models := map[string]interface{}{
		"selected": map[string]interface{}{
			"large": cfg.Models[config.SelectedModelTypeLarge],
			"small": cfg.Models[config.SelectedModelTypeSmall],
		},
		"recent": cfg.RecentModels,
	}
	respondJSON(w, http.StatusOK, models)
}

func (h *handlers) getProviders(w http.ResponseWriter, r *http.Request) {
	cfg := h.app.Config()
	providers := make([]map[string]interface{}, 0)
	for prov := range cfg.Providers.Seq() {
		provider := map[string]interface{}{
			"id":      prov.ID,
			"name":    prov.Name,
			"type":    prov.Type,
			"disable": prov.Disable,
			"models":  prov.Models,
		}
		// Remove sensitive data
		providers = append(providers, provider)
	}
	respondJSON(w, http.StatusOK, providers)
}

func (h *handlers) setModel(w http.ResponseWriter, r *http.Request) {
	var req setModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg := h.app.Config()
	modelType := config.SelectedModelType(req.ModelType)
	if modelType != config.SelectedModelTypeLarge && modelType != config.SelectedModelTypeSmall {
		http.Error(w, "Invalid model type", http.StatusBadRequest)
		return
	}

	// Validate that the model exists for the provider
	modelConfig := cfg.GetModel(req.Provider, req.Model)
	if modelConfig == nil {
		http.Error(w, "Model not found for provider", http.StatusNotFound)
		return
	}

	newModel := config.SelectedModel{
		Provider: req.Provider,
		Model:    req.Model,
	}

	if err := cfg.UpdatePreferredModel(modelType, newModel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Model updated successfully",
		"model":   newModel,
	})
}

// getConfig returns the full application configuration with optional schema metadata.
func (h *handlers) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.app.Config()

	// Check for include_schema query parameter
	includeSchema := r.URL.Query().Get("include_schema") == "true"

	// Convert to API response with redaction
	resp := ToConfigResponse(cfg, includeSchema)

	respondJSON(w, http.StatusOK, resp)
}

// updateConfig handles bulk configuration updates.
func (h *handlers) updateConfig(w http.ResponseWriter, r *http.Request) {
	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the update request
	if err := ValidateConfigUpdate(&req); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	cfg := h.app.Config()

	// Apply the configuration updates
	if err := ApplyConfigUpdate(cfg, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply config update: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Emit SSE event for config_updated (Phase C)

	// Return the updated configuration
	resp := ToConfigResponse(cfg, false)
	respondJSON(w, http.StatusOK, resp)
}

// updateConfigField handles single field configuration updates.
func (h *handlers) updateConfigField(w http.ResponseWriter, r *http.Request) {
	var req FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the field update request
	if err := ValidateFieldUpdate(&req); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	cfg := h.app.Config()

	// Apply the field update
	if err := ApplyFieldUpdate(cfg, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply field update: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Emit SSE event for config_updated (Phase C)

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Field updated successfully",
		"path":    req.Path,
		"value":   req.Value,
	})
}
