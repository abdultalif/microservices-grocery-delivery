package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"

	"github.com/spf13/cobra"
)

var workerUpdateStatusCmd = &cobra.Command{
	Use:   "worker-update-status-order",
	Short: "Menjalankan worker untuk consume RabbitMQ dan index ke Elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker untuk Order Indexing sedang berjalan...")
		message.ConsumeUpdateStatus()
	},
}
