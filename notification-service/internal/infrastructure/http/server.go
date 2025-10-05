package http

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/config"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/handlers"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/infrastructure/database"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/infrastructure/messaging"
	adapter "github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/pkg"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/repositories"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/services"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/worker"

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

	repoNotif := repositories.NewRepositoryNotif(db.DB)

	jwtService := services.NewJwtService(cfg)
	serviceNotif := services.NewServiceNotification(repoNotif)

	emailMessage := messaging.NewMessageEmail(cfg)
	rabbitMQ := worker.NewConsumeRabbitMQ(emailMessage, repoNotif, serviceNotif)

	midAuth := adapter.NewmiddlewareAuth(cfg, jwtService, redisClient)
	midRateLimiter := adapter.NewRateLimiterMiddleware(redisClient)

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_VERIFICATION)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_VERIFICATION, err)
		}
	}()

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_FORGOT_PASSWORD)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_FORGOT_PASSWORD, err)
		}
	}()

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_CREATE_CUSTOMER)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_CREATE_CUSTOMER, err)
		}
	}()

	// go func() {
	// 	err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_UPDATE_CUSTOMER)
	// 	if err != nil {
	// 		e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_UPDATE_CUSTOMER, err)
	// 	}
	// }()

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.NOTIF_EMAIL_UPDATE_STATUS_ORDER)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", pkg.NOTIF_EMAIL_UPDATE_STATUS_ORDER, err)
		}
	}()

	go func() {
		err := rabbitMQ.ConsumeMessage(pkg.PUSH_NOTIF)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", pkg.PUSH_NOTIF, err)
		}
	}()

	notifHandler := handlers.NewNotifHandler(serviceNotif)
	wsHandle := handlers.NewWebSocketHandler(e)
	NotifRouter(e, notifHandler, wsHandle, cfg, jwtService, redisClient, midRateLimiter, midAuth)

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
