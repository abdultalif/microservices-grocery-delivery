package cmd

import (
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/app"
	"github.com/spf13/cobra"
)

// Command untuk menjalankan REST API (Echo)
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start REST API server",
	Long:  "Run the REST API server for user-service",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
