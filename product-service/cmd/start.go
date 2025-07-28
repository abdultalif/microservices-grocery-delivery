package cmd

import (
	"product-service/internal/app"
	"product-service/utils/validator"

	"github.com/spf13/cobra"
)

var productValidator = validator.NewValidator()

var startCmd = &cobra.Command{
	Use: "start",
	Short: "Start",
	Long: "Start",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer()
	},

}