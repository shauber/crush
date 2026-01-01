package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/crush-sync/pkg/types"
	_ "github.com/ncruces/go-sqlite3/driver"
)

type Writer struct {
	db *sql.DB
}

func NewWriter(dataDir string) (*Writer, error) {
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

	return &Writer{db: db}, nil
}

func (w *Writer) Close() error {
	return w.db.Close()
}

func (w *Writer) InsertConversation(conv *types.Conversation) error {
	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert session
	if _, err := tx.Exec(`
		INSERT INTO sessions (
			id, parent_session_id, title, message_count,
			prompt_tokens, completion_tokens, cost, updated_at,
			created_at, summary_message_id, todos
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, conv.Session.ID, conv.Session.ParentSessionID, conv.Session.Title,
		conv.Session.MessageCount, conv.Session.PromptTokens,
		conv.Session.CompletionTokens, conv.Session.Cost,
		conv.Session.UpdatedAt, conv.Session.CreatedAt,
		conv.Session.SummaryMessageID, conv.Session.Todos); err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	// Insert messages
	for _, msg := range conv.Messages {
		if _, err := tx.Exec(`
			INSERT INTO messages (
				id, session_id, role, parts, model,
				created_at, updated_at, finished_at,
				provider, is_summary_message
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, msg.ID, msg.SessionID, msg.Role, msg.Parts, msg.Model,
			msg.CreatedAt, msg.UpdatedAt, msg.FinishedAt,
			msg.Provider, msg.IsSummaryMessage); err != nil {
			return fmt.Errorf("failed to insert message: %w", err)
		}
	}

	// Insert files
	for _, file := range conv.Files {
		if _, err := tx.Exec(`
			INSERT INTO files (
				id, session_id, path, content, version,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, file.ID, file.SessionID, file.Path, file.Content,
			file.Version, file.CreatedAt, file.UpdatedAt); err != nil {
			return fmt.Errorf("failed to insert file: %w", err)
		}
	}

	return tx.Commit()
}

func (w *Writer) CreateBackup(backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = "."
	}

	timestamp := time.Now().Format("2006-01-02-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("crush-backup-%s.db", timestamp))

	sourcePath := filepath.Join(dataDirDefault, dbFilename)
	
	// Copy database file
	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source database: %w", err)
	}

	if err := os.WriteFile(backupPath, contents, 0o644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}