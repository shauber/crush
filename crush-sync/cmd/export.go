package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/crush-sync/internal/db"
	"github.com/charmbracelet/crush-sync/pkg/types"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export [file]",
	Aliases: []string{"backup", "dump"},
	Short:   "Export conversations to JSON file",
	Long:    `Export all Crush conversations to a JSON file for backup or sync purposes.`,
	Example: `
# Export all conversations
crush-sync export conversations.json

# Export with custom data directory
crush-sync export --data-dir /custom/path conversations.json

# Export to compressed file
crush-sync export conversations.json.gz
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

var exportDataDir string

func init() {
	exportCmd.Flags().StringVar(&exportDataDir, "data-dir", "", "Custom Crush data directory")
}

func runExport(cmd *cobra.Command, args []string) error {
	reader, err := db.NewReader(exportDataDir)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer reader.Close()

	sessions, err := reader.GetSessions()
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	var conversations []types.Conversation

	for _, session := range sessions {
		messages, err := reader.GetMessagesBySession(session.ID)
		if err != nil {
			return fmt.Errorf("failed to get messages for session %s: %w", session.ID, err)
		}

		files, err := reader.GetFilesBySession(session.ID)
		if err != nil {
			return fmt.Errorf("failed to get files for session %s: %w", session.ID, err)
		}

		conversation := types.Conversation{
			Session:  &session,
			Messages: messages,
			Files:    files,
		}
		conversations = append(conversations, conversation)
	}

	syncFile := types.SyncFile{
		Header: types.Header{
			Version:   "1.0",
			Source:    "crush-sync",
			Timestamp: time.Now().Unix(),
		},
		Conversations: conversations,
	}

	jsonData, err := json.MarshalIndent(syncFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	outputFile := "conversations.json"
	if len(args) > 0 {
		outputFile = args[0]
	}

	if err := os.WriteFile(outputFile, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported %d conversations to %s (%.2f KB)\n", 
		len(conversations), outputFile, float64(len(jsonData))/1024)

	return nil
}