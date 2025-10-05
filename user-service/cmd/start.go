package cmd

import (
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/app"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/utils/validator"

	"github.com/spf13/cobra"
)

var userValidator = validator.NewValidator()

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start",
	Long:  "Start",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer()
	},
}
