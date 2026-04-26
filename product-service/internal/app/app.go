package app

import (
	"context"
	"os"
	"os/signal"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler"
	adapter "github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/router"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/storage"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/service"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/utils/validator"

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

	redis, err := cfg.NewRedisClient()
	if err != nil {
		log.Fatalf("[RunServer-1] failed to connect redis: %v", err)
		return
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(adapter.GatewayValidationMiddleware())

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator
	elasticseachInit, err := cfg.InitElasticsearch()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		return
	}

	productRepository := repository.NewProductRepository(db.DB, elasticseachInit)
	categoryRepo := repository.NewCategoryRepository(db.DB)
	cartRepo := repository.NewCartRedisRepository(redis)

	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepository, categoryRepo)
	jwtService := service.NewJwtService(cfg)
	cartService := service.NewCartService(cartRepo, productRepository)

	uploadHandler := handler.NewUploadImageHandler(productService, storage.NewSupabase(cfg))
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	cartHandler := handler.NewCartHandler(cartService, productService)

	router.NewRouter(
		e,
		categoryHandler,
		productHandler,
		uploadHandler,
		cartHandler,
		cfg,
		jwtService,
		redis,
		adapter.NewRateLimiterMiddleware(redis))

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
