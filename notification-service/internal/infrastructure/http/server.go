package httpinfra

import (
	"notification-service/internal/config"
	"notification-service/internal/infrastructure/messaging"
	"notification-service/internal/pkg"
	"notification-service/internal/worker"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func StartHTTPServer() {
	e := echo.New()

	// global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	cfg := config.NewConfig()

	emailMessage := messaging.NewMessageEmail(cfg)
	rabbitMQ := worker.NewConsumeRabbitMQ(emailMessage)

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_NOTIFICATION)
		if err != nil {
			e.Logger.Fatalf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_NOTIFICATION, err)
		}
	}()

	e.Logger.Fatal(e.Start(":" + cfg.App.AppPort))

}
