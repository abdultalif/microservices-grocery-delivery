package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	"github.com/spf13/cobra"
)

var workerSyncProductCmd = &cobra.Command{
	Use:   "worker-sync-product",
	Short: "Menjalankan worker untuk sync data product dari product service ke local DB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker sync product.updated sedang berjalan...")
		message.ConsumeProductUpdated()
	},
}

func init() {
	rootCmd.AddCommand(workerSyncProductCmd)
}
