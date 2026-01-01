package types

import (
	"database/sql"
)

// Conversation represents a complete Crush conversation including session, messages, and files
type Conversation struct {
	Session *Session   `json:"session"`
	Messages []Message `json:"messages"`
	Files   []File    `json:"files"`
}

// Session represents a Crush session
type Session struct {
	ID               string         `json:"id"`
	ParentSessionID  sql.NullString `json:"parent_session_id"`
	Title            string         `json:"title"`
	MessageCount     int64          `json:"message_count"`
	PromptTokens     int64          `json:"prompt_tokens"`
	CompletionTokens int64          `json:"completion_tokens"`
	Cost             float64        `json:"cost"`
	UpdatedAt        int64          `json:"updated_at"`
	CreatedAt        int64          `json:"created_at"`
	SummaryMessageID sql.NullString `json:"summary_message_id"`
	Todos            sql.NullString `json:"todos"`
}

// Message represents a Crush message
type Message struct {
	ID               string         `json:"id"`
	SessionID        string         `json:"session_id"`
	Role             string         `json:"role"`
	Parts            string         `json:"parts"`
	Model            sql.NullString `json:"model"`
	CreatedAt        int64          `json:"created_at"`
	UpdatedAt        int64          `json:"updated_at"`
	FinishedAt       sql.NullInt64  `json:"finished_at"`
	Provider         sql.NullString `json:"provider"`
	IsSummaryMessage int64          `json:"is_summary_message"`
}

// File represents a Crush file
type File struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Version   int64  `json:"version"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Header represents sync file metadata
type Header struct {
	Version   string `json:"version"`
	Source    string `json:"source"`
	Timestamp int64  `json:"timestamp"`
}

// SyncFile represents the complete sync file structure
type SyncFile struct {
	Header       Header        `json:"header"`
	Conversations []Conversation `json:"conversations"`
}