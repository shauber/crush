package cmd

import (
	"fmt"

	"github.com/charmbracelet/crush-sync/internal/db"
	"github.com/spf13/cobra"
)

var (
	statsDataDir string
)

var statsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"info", "status"},
	Short:   "Show conversation statistics",
	Long:    `Display statistics about conversations in your Crush database.`,
	Example: `
# Show default stats
crush-sync stats

# Show stats for custom data directory
crush-sync stats --data-dir /custom/path
`,
	Args: cobra.NoArgs,
	RunE: runStats,
}

func init() {
	statsCmd.Flags().StringVar(&statsDataDir, "data-dir", "", "Custom Crush data directory")
}

func runStats(cmd *cobra.Command, args []string) error {
	reader, err := db.NewReader(statsDataDir)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer reader.Close()

	sessions, err := reader.GetSessions()
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	var totalMessages, totalFiles int
	for _, session := range sessions {
		messages, err := reader.GetMessagesBySession(session.ID)
		if err != nil {
			return fmt.Errorf("failed to get messages: %w", err)
		}
		totalMessages += len(messages)

		files, err := reader.GetFilesBySession(session.ID)
		if err != nil {
			return fmt.Errorf("failed to get files: %w", err)
		}
		totalFiles += len(files)
	}

	fmt.Printf("Crush Database Statistics:\n")
	fmt.Printf("  Sessions: %d\n", len(sessions))
	fmt.Printf("  Messages: %d\n", totalMessages)
	fmt.Printf("  Files: %d\n", totalFiles)
	
	if len(sessions) > 0 {
		fmt.Printf("\nRecent Sessions:\n")
		maxToShow := 5
		if len(sessions) < maxToShow {
			maxToShow = len(sessions)
		}
		
		for i := 0; i < maxToShow; i++ {
			session := sessions[i]
			messages, _ := reader.GetMessagesBySession(session.ID)
			fmt.Printf("  • %s (%d messages)\n", session.Title, len(messages))
		}
		
		if len(sessions) > maxToShow {
			fmt.Printf("  ... and %d more\n", len(sessions)-maxToShow)
		}
	}

	return nil
}