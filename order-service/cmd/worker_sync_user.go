package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	"github.com/spf13/cobra"
)

var workerSyncUserCmd = &cobra.Command{
	Use:   "worker-sync-user",
	Short: "Menjalankan worker untuk sync data buyer dari user service ke local DB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Worker sync user started...")

		cfg := config.NewConfig()

		// Context ini yang mengontrol hidup-matinya consumer.
		// Saat SIGTERM/SIGINT diterima, ctx dibatalkan → consumer berhenti bersih.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		// Consumer jalan di goroutine, bukan blocking di sini.
		go message.ConsumeUserEvents(ctx, *cfg)

		// Block di sini sampai sinyal shutdown diterima.
		<-quit
		fmt.Println("Worker sync user shutting down...")
		cancel() // hentikan consumer loop dengan bersih
	},
}

func init() {
	rootCmd.AddCommand(workerSyncUserCmd)
}
