package http

import (
	"context"
	"log"
	"notification-service/internal/config"
	"notification-service/internal/handlers"
	"notification-service/internal/infrastructure/database"
	"notification-service/internal/infrastructure/messaging"
	adapter "notification-service/internal/middleware"
	"notification-service/internal/pkg"
	"notification-service/internal/repositories"
	"notification-service/internal/services"
	"notification-service/internal/worker"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("Database connection failed")
	}

	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	emailMessage := messaging.NewMessageEmail(cfg)
	rabbitMQ := worker.NewConsumeRabbitMQ(emailMessage)

	repoNotif := repositories.NewRepositoryNotif(db.DB)

	jwtService := services.NewJwtService(cfg)
	serviceNotif := services.NewServiceNotification(repoNotif)

	midAuth := adapter.NewmiddlewareAuth(cfg, jwtService, redisClient)
	midRateLimiter := adapter.NewRateLimiterMiddleware(redisClient)

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_NOTIFICATION)
		if err != nil {
			e.Logger.Fatalf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_NOTIFICATION, err)
		}
	}()

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_FORGOT_PASSWORD)
		if err != nil {
			e.Logger.Fatalf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_FORGOT_PASSWORD, err)
		}
	}()

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_UPDATE_STATUS_ORDER)
		if err != nil {
			e.Logger.Fatalf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_UPDATE_STATUS_ORDER, err)
		}
	}()

	notifHandler := handlers.NewNotifHandler(serviceNotif)
	NotifRouter(e, notifHandler, cfg, jwtService, redisClient, midRateLimiter, midAuth)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err = e.Start(":" + cfg.App.AppPort)
		if err != nil {
			log.Fatalf("[RunServer-2] %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	<-quit

	log.Print("[RunServer-3] Shutting down server of 5 second...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	e.Shutdown(ctx)

}
