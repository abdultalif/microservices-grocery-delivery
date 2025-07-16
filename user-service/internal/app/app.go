package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/config"
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/repository"
	"user-service/internal/adapter/storage"
	"user-service/internal/core/service"
	"user-service/utils/validator"

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

	// repositories
	userRepo := repository.NewUserRepository(db.DB)
	authRepo := repository.NewAuthRepository(db.DB)
	customerRepo := repository.NewCustomerRepository(db.DB)
	tokenRepo := repository.NewVerficationTokenRepository(db.DB)

	// services
	jwtService := service.NewJwtService(cfg)
	authService := service.NewAuthService(authRepo,cfg, jwtService, tokenRepo)
	customerService := service.NewCustomerService(customerRepo, authRepo,cfg, jwtService, tokenRepo)
	userService := service.NewUserService(userRepo, authRepo,cfg, jwtService, tokenRepo)
	roleService := service.NewServiceRole(repository.NewRoleRepository(db.DB))


	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	
	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator

	apiGroup := e.Group("/api/v1")
	handler.NewAuthHandler(apiGroup, authService)
	handler.NewCustomerHandler(apiGroup, customerService, userService, cfg, jwtService)
	handler.NewUserHandler(apiGroup, userService, cfg, jwtService)
	handler.NewRoleHandler(roleService, apiGroup, cfg, jwtService)
	handler.NewUploadImageHandler(apiGroup, userService, cfg, storage.NewSupabase(cfg), jwtService)
	
	e.Logger.Fatal(e.Start(":8080"))

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