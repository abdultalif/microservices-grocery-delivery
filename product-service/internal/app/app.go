package app

import (
	"context"
	"os"
	"os/signal"
	"product-service/config"
	"product-service/internal/adapter/handler"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/service"
	"syscall"
	"time"

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
	
	// customValidator := validator.NewValidator()
	// en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	// e.Validator = customValidator

	productRepository := repository.NewProductRepository(db.DB)

	productService := service.ProductServiceInterface(productRepository)
	jwtService := service.NewJwtService(cfg)

	apiGroup := e.Group("/api/v1")
	handler.NewProductHandler(apiGroup, productService, cfg, jwtService)
	
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