package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"

	"github.com/spf13/cobra"
)

var workerUpdatePaymentOrderCmd = &cobra.Command{
	Use:   "worker-update-payment-order",
	Short: "Menjalankan worker untuk consume RabbitMQ dan index ke Elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(
			"Worker payment event started...",
		)

		message.ConsumePaymentEvent()
	},
}

func init() {
	rootCmd.AddCommand(
		workerUpdatePaymentOrderCmd,
	)
}
