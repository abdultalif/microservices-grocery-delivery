package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/config"
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/logger"
	adapter "user-service/internal/adapter/middleware"
	"user-service/internal/adapter/repository"
	"user-service/internal/adapter/router"
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

	redisClient, err := cfg.NewRedisClient()
	if err != nil {
		log.Fatalf("[RunServer-1] failed to connect redis: %v", err)
		return
	}

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[RunServer-2] failed to connect postgres: %v", err)
		return
	}

	logFile, err := os.OpenFile("oauth_activity.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("[RunServer] failed to open log file: %v", err)
	}
	defer logFile.Close()

	fileLogger := logger.NewLogger(logFile)

	userRepo := repository.NewUserRepository(db.DB)
	authRepo := repository.NewAuthRepository(db.DB)
	customerRepo := repository.NewCustomerRepository(db.DB)
	tokenRepo := repository.NewVerficationTokenRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)
	oauth := repository.NewOAuthRepository(db.DB)

	jwtService := service.NewJwtService(cfg)
	authService := service.NewAuthService(authRepo, cfg, jwtService, tokenRepo)
	customerService := service.NewCustomerService(customerRepo, authRepo, cfg, jwtService, tokenRepo)
	userService := service.NewUserService(userRepo, authRepo, cfg, jwtService, tokenRepo)
	roleService := service.NewServiceRole(roleRepo)
	oauthService := service.NewOAuthService(userRepo, oauth, cfg, jwtService, authRepo, fileLogger)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator

	authHandler := handler.NewAuthHandler(authService, cfg, jwtService)
	customerHandler := handler.NewCustomerHandler(customerService, userService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	uploadHandler := handler.NewUploadImageHandler(userService, storage.NewSupabase(cfg))
	oauthHandler := handler.NewOAuthHandler(oauthService, cfg)

	router.NewRouterUserService(
		e,
		authHandler,
		customerHandler,
		roleHandler,
		userHandler,
		uploadHandler,
		cfg,
		jwtService,
		adapter.NewRateLimiterMiddleware(redisClient),
		redisClient,
		oauthHandler,
	)

	port := cfg.App.AppPort
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[RunServer-3] server start failed: %v", err)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Print("[RunServer-4] Shutting down server in 5 seconds...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("[RunServer-5] server forced to shutdown: %v", err)
	}
}
