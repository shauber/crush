package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/crush-sync/pkg/types"
	_ "github.com/ncruces/go-sqlite3/driver"
)

const (
	dataDirDefault = ".crush"
	dbFilename     = "crush.db"
)

type Reader struct {
	db *sql.DB
}

func NewReader(dataDir string) (*Reader, error) {
	if dataDir == "" {
		dataDir = dataDirDefault
	}

	dbPath := filepath.Join(dataDir, dbFilename)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Reader{db: db}, nil
}

func (r *Reader) Close() error {
	return r.db.Close()
}

func (r *Reader) GetSessions() ([]types.Session, error) {
	query := `
		SELECT id, parent_session_id, title, message_count, 
		       prompt_tokens, completion_tokens, cost, updated_at, 
		       created_at, summary_message_id, todos
		FROM sessions
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var s types.Session
		if err := rows.Scan(
			&s.ID, &s.ParentSessionID, &s.Title, &s.MessageCount,
			&s.PromptTokens, &s.CompletionTokens, &s.Cost,
			&s.UpdatedAt, &s.CreatedAt, &s.SummaryMessageID,
			&s.Todos,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *Reader) GetMessagesBySession(sessionID string) ([]types.Message, error) {
	query := `
		SELECT id, session_id, role, parts, model, created_at,
		       updated_at, finished_at, provider, is_summary_message
		FROM messages
		WHERE session_id = ?
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var m types.Message
		if err := rows.Scan(
			&m.ID, &m.SessionID, &m.Role, &m.Parts, &m.Model,
			&m.CreatedAt, &m.UpdatedAt, &m.FinishedAt,
			&m.Provider, &m.IsSummaryMessage,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (r *Reader) GetFilesBySession(sessionID string) ([]types.File, error) {
	query := `
		SELECT id, session_id, path, content, version, created_at, updated_at
		FROM files
		WHERE session_id = ?
		ORDER BY path ASC`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []types.File
	for rows.Next() {
		var f types.File
		if err := rows.Scan(
			&f.ID, &f.SessionID, &f.Path, &f.Content,
			&f.Version, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, f)
	}

	return files, nil
}