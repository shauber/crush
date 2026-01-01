package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crush-sync",
	Short: "Sync Crush conversations between instances",
	Long: `crush-sync is a standalone tool for synchronizing Crush conversations
between different machine instances. It provides export, import, and sync
capabilities for seamless conversation portability.

Find your Crush database at: ~/.crush/crush.db

Usage:
  crush-sync export [file]          # Export conversations to JSON
  crush-sync import [file]          # Import conversations from JSON
  crush-sync backup [dir]           # Create backup of database
  crush-sync stats                  # Show conversation stats
`,
	Example: `
# Export all conversations
crush-sync export my-conversations.json

# Import from backup
crush-sync import my-conversations.json

# Create daily backup
crush-sync backup ~/crush-backups/
`,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(statsCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}