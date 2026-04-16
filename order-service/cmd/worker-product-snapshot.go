package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	"github.com/spf13/cobra"
)

var workerProductSnapShotCmd = &cobra.Command{
	Use:   "worker-product-snapshot",
	Short: "Menjalankan worker untuk consume RabbitMQ dan index ke Elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker untuk produk snapshot indexing sedang berjalan...")
		message.ConsumeFromProduct()
	},
}
