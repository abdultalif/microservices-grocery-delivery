package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/handler"
	httpclient "github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/http_client"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/message"
	adapter "github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/router"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/service"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/utils/validator"

	"github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func RunServer() {
	cfg := config.NewConfig()
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[RunServer-1] %v", err)
		return
	}

	redisClient, err := cfg.NewRedisClient()
	if err != nil {
		log.Fatalf("[RunServer-1] failed to connect redis: %v", err)
		return
	}

	paymentRepo := repository.NewPaymentRepository(db.DB)
	httpClient := httpclient.NewHttpClient(cfg)
	midtrans := httpclient.NewMidtransClient(cfg)

	rabbitConn, err := cfg.NewRabbitMQ()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		panic(err)
	}

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		panic(err)
	}

	publisherRabbitMQ := message.NewPublishRabbitMQ(
		rabbitChannel,
	)

	paymentService := service.NewPaymentService(paymentRepo, cfg, httpClient, midtrans, publisherRabbitMQ)
	jwtService := service.NewJwtService(cfg)

	paymentHandler := handler.NewPaymentHandler(paymentService)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(adapter.GatewayValidationMiddleware())

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator

	e.GET("/api/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	router.NewRouterPaymentService(e, paymentHandler, cfg, jwtService, adapter.NewRateLimiterMiddleware(redisClient), redisClient)

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
