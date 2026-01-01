package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/crush-sync/internal/db"
	"github.com/spf13/cobra"
)

var (
	backupDataDir string
)

var backupCmd = &cobra.Command{
	Use:     "backup [directory]",
	Aliases: []string{"safe"},
	Short:   "Create database backup",
	Long: `Create a timestamped backup of the Crush database.
Backups are created in the specified directory or current directory with timestamp.`,
	Example: `
# Backup to current directory
crush-sync backup

# Backup to specific directory
crush-sync backup /path/to/backups/

# Backup with custom data directory
crush-sync backup --data-dir /custom/path ~/backups/
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBackup,
}

func init() {
	backupCmd.Flags().StringVar(&backupDataDir, "data-dir", "", "Custom Crush data directory")
}

func runBackup(cmd *cobra.Command, args []string) error {
	reader, err := db.NewReader(backupDataDir)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer reader.Close()

	backupDir := "."
	if len(args) > 0 {
		backupDir = args[0]
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get the actual database path
	dataDir := backupDataDir
	if dataDir == "" {
		dataDir = ".crush"
	}
	dbPath := filepath.Join(dataDir, "crush.db")

	// Create backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	backupFile := fmt.Sprintf("crush-backup-%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupFile)

	// Copy database file
	contents, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database file: %w", err)
	}

	if err := os.WriteFile(backupPath, contents, 0o644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	// Get file info for feedback
	stat, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("failed to get backup info: %w", err)
	}

	fmt.Printf("Created backup: %s (%.2f MB)\n", 
		backupPath, float64(stat.Size())/1024/1024)

	return nil
}