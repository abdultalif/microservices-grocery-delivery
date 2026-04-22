package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/message"
	"github.com/spf13/cobra"
)

var workerUpdateLocationCmd = &cobra.Command{
	Use:   "worker-update-location",
	Short: "Worker untuk update lokasi user dari order service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker update location sedang berjalan...")
		message.ConsumeUpdateUserLocation()
	},
}

func init() {
	rootCmd.AddCommand(workerUpdateLocationCmd)
}
