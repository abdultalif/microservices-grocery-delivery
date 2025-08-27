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
		rateLimiter.RateLimiter(RateLimitLogin, RateLimitMaxRequestsShort))

	authGroup.POST("/auth", authHandler.CreateUserAccount,
		rateLimiter.RateLimiter(RateLimitCreateAccount, RateLimitWindowOneHour))

	authGroup.GET("/auth/verify-account", authHandler.VerifyAccount,
		rateLimiter.RateLimiter(RateLimitVerifyAccount, RateLimitMaxRequestsLong))

	authGroup.POST("/auth/forgot-password", authHandler.ForgotPassword,
		rateLimiter.RateLimiter(RateLimitForgotPassword, RateLimitWindowFifteenMins))

	authGroup.GET("/auth/validate-forgot-token", authHandler.ValidateForgotPasswordToken,
		rateLimiter.RateLimiter(RateLimitValidateForgotToken, RateLimitMaxRequestsShort))

	authGroup.PATCH("/auth/reset-password", authHandler.UpdatePassword,
		rateLimiter.RateLimiter(RateLimitResetPassword, RateLimitWindowFifteenMins))

	authGroup.POST("/auth/service-token", authHandler.GenerateServiceToken,
		rateLimiter.RateLimiter(RateLimitServiceToken, RateLimitWindowOneMinute))

}
