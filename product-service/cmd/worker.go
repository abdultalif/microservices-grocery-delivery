package cmd

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/message"

	"github.com/spf13/cobra"
)

var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Jalanin worker untuk consume ke rabbitmq dan index ke elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker untuk consume ke rabbitmq dan index ke elasticsearch sedang berjalan")
		message.StartConsumer()
	},
}

func init() {
	rootCmd.AddCommand(WorkerCmd)
}
