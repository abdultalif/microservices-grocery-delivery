package router

import (
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func AuthRouter(
	e *echo.Echo,
	authHandler handler.AuthHandlerInterface,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
) {
	authGroup := e.Group("/api/v1")

	authGroup.POST("/auth/login", authHandler.SignIn,
		rateLimiter.RateLimiter(5, 600))

	authGroup.POST("/auth", authHandler.CreateUserAccount,
		rateLimiter.RateLimiter(10, 3600))

	authGroup.GET("/auth/verify-account", authHandler.VerifyAccount,
		rateLimiter.RateLimiter(5, 1800))

	authGroup.POST("/auth/forgot-password", authHandler.ForgotPassword,
		rateLimiter.RateLimiter(3, 900))

	authGroup.GET("/auth/validate-forgot-token", authHandler.ValidateForgotPasswordToken,
		rateLimiter.RateLimiter(5, 600))

	authGroup.PATCH("/auth/reset-password", authHandler.UpdatePassword,
		rateLimiter.RateLimiter(3, 900))

	// Service token lebih longgar
	authGroup.POST("/auth/service-token", authHandler.GenerateServiceToken,
		rateLimiter.RateLimiter(20, 60))

}
