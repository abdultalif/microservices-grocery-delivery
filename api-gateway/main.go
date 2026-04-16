package main

import (
	"log"
	"net/http"
	"os"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/handlers"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/middleware/cors"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/middleware/logger"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/middleware/ratelimit"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	logrusLogger := logrus.New()
	logrusLogger.SetFormatter(&logrus.JSONFormatter{})

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(logger.Middleware())
	e.Use(cors.Middleware())
	e.Use(ratelimit.RedisMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")
	handlers.RegisterAllRoutes(api, ratelimit.RedisMiddleware())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrusLogger.Infof("API Gateway starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		logrusLogger.Fatal(err)
	}
}
