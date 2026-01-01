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

var (
	importDataDir string
	importNoBackup bool
	importDryRun  bool
)

var importCmd = &cobra.Command{
	Use:     "import [file]",
	Aliases: []string{ "load"},
	Short:   "Import conversations from JSON file",
	Long: `Import conversations from a JSON file created by crush-sync export.
Performs backup before import by default to ensure safety.`,
	Example: `
# Import from a file (creates backup automatically)
crush-sync import conversations.json

# Import without backup (dangerous)
crush-sync import --no-backup conversations.json

# Dry run to see what would be imported
crush-sync import --dry-run conversations.json

# Import with custom data directory
crush-sync import --data-dir /custom/path conversations.json
`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	importCmd.Flags().StringVar(&importDataDir, "data-dir", "", "Custom Crush data directory")
	importCmd.Flags().BoolVar(&importNoBackup, "no-backup", false, "Skip backup before import")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Show what would be imported without changes")
}

func runImport(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Read JSON file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse sync file
	var syncFile types.SyncFile
	if err := json.Unmarshal(jsonData, &syncFile); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate file format
	if syncFile.Header.Version != "1.0" {
		return fmt.Errorf("unsupported sync file version: %s", syncFile.Header.Version)
	}

	// Check dry run
	if importDryRun {
		fmt.Printf("Would import %d conversations:\n", len(syncFile.Conversations))
		for i, conv := range syncFile.Conversations {
			fmt.Printf("  %d. %s (%d messages, %d files)\n", 
				i+1, conv.Session.Title, len(conv.Messages), len(conv.Files))
		}
		return nil
	}

	// Create backup
	backupPath := ""
	if !importNoBackup {
		writer, err := db.NewWriter(importDataDir)
		if err != nil {
			return fmt.Errorf("failed to open database for backup: %w", err)
		}
		defer writer.Close()

		backupPath, err = writer.CreateBackup("")
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("Created backup: %s\n", backupPath)
	}

	// Import conversations
	writer, err := db.NewWriter(importDataDir)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer writer.Close()

	importedCount := 0
	totalMessages := 0
	totalFiles := 0

	for _, conv := range syncFile.Conversations {
		// Skip if conversation is empty
		if len(conv.Messages) == 0 {
			continue
		}

		// Generate new UUIDs to avoid conflicts
		newSessionID := generateUUID()
		oldSessionID := conv.Session.ID
		
		// Update session ID
		conv.Session.ID = newSessionID
		
		// Update message session IDs
		for i := range conv.Messages {
			conv.Messages[i].SessionID = newSessionID
			conv.Messages[i].ID = generateUUID() 
		}
		
		// Update file session IDs
		for i := range conv.Files {
			conv.Files[i].SessionID = newSessionID
			conv.Files[i].ID = generateUUID()
		}

		// Insert into database
		if err := writer.InsertConversation(&conv); err != nil {
			return fmt.Errorf("failed to import conversation %s: %w", oldSessionID, err)
		}

		importedCount++
		totalMessages += len(conv.Messages)
		totalFiles += len(conv.Files)
	}

	fmt.Printf("Successfully imported %d conversations (%d messages, %d files)\n", 
		importedCount, totalMessages, totalFiles)

	return nil
}

// Mock UUID generation - in practice, use proper UUID generation
func generateUUID() string {
	return fmt.Sprintf("sync-%d", time.Now().UnixNano())
}