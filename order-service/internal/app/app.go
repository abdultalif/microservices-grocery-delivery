package app

import (
	"context"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"

	"os"
	"os/signal"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/handler"
	httpclient "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/http_client"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	adapter "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/router"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/service"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/utils/validator"

	"syscall"
	"time"

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

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator
	publisherRabbitMQ := message.NewPublisherRabbitMQ(cfg)
	elasticseachInit, err := cfg.InitElasticsearch()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		return
	}

	httpClient := httpclient.NewHttpClient(cfg)
	orderRepo := repository.NewOrderRepository(db.DB)
	elasticRepo := repository.NewElasticRepository(elasticseachInit)

	orderService := service.NewOrderService(orderRepo, cfg, httpClient, publisherRabbitMQ, elasticRepo)
	jwtService := service.NewJwtService(cfg)

	orderHandler := handler.NewOrderHandler(orderService)

	midDistance := adapter.NewMiddlewareDistance(cfg, orderService, httpClient)
	midAuth := adapter.NewmiddlewareAuth(cfg, jwtService, redisClient)
	midRateLimiter := adapter.NewRateLimiterMiddleware(redisClient)

	router.OrderRouter(e, orderHandler, cfg, jwtService, redisClient, midRateLimiter, midDistance, midAuth)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err := e.Start(":" + cfg.App.AppPort)
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
