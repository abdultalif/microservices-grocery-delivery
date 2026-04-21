package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	"github.com/spf13/cobra"
)

var workerSyncUserCmd = &cobra.Command{
	Use:   "worker-sync-user",
	Short: "Menjalankan worker untuk sync data buyer dari user service ke local DB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker sync user.updated sedang berjalan...")
		message.ConsumeUserUpdated()
	},
}

func init() {
	rootCmd.AddCommand(workerSyncUserCmd)
}
