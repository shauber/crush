package cmd

import (
	"github.com/charmbracelet/crush/internal/api"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Crush as HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := setupApp(cmd)
		if err != nil {
			return err
		}
		defer app.Shutdown()

		port, _ := cmd.Flags().GetInt("port")
		return api.Serve(app, port)
	},
}

func init() {
	serverCmd.Flags().IntP("port", "p", 8080, "HTTP port")
}
