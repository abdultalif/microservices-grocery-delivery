package app

import (
	"context"
	"os"
	"os/signal"
	"product-service/config"
	"product-service/internal/adapter/handler"
	"product-service/internal/adapter/message"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/service"
	"product-service/utils/validator"
  "product-service/internal/adapter/storage"

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

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	
	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator
	publisherRabbitMQ := message.NewPublishRabbitMQ(cfg)

	productRepository := repository.NewProductRepository(db.DB)
	categoryRepo := repository.NewCategoryRepository(db.DB)
  repoUpload := repository.NewProductRepository(db.DB)
	
  serviceUpload := service.NewProductService(repoUpload)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepository, publisherRabbitMQ)
	jwtService := service.NewJwtService(cfg)

	apiGroup := e.Group("/api/v1")
  handler.NewUploadImageHandler(apiGroup, serviceUpload, cfg, storage.NewSupabase(cfg), jwtService)
	handler.NewProductHandler(apiGroup, productService, cfg, jwtService)
  handler.NewCategoryHandler(apiGroup, categoryService, cfg, jwtService)
	
	go func () {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err = e.Start(":"+cfg.App.AppPort)
		if err != nil {
			log.Fatalf("[RunServer-2] %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	<-quit

	log.Print("[RunServer-3] Shutting down server of 5 second...")
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	e.Shutdown(ctx)
}