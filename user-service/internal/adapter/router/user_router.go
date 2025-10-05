package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func UserRouter(
	e *echo.Echo,
	userHandler handler.UserHandlerInterface,
	uploadHandler handler.UploadImageHandlerInterface,
	mid middleware.AuthMiddlewareInterface,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
) {

	userGroup := e.Group("/api/v1/user", mid.CheckToken(), mid.CheckRole("Customer", "Super Admin"))

	userGroup.GET("/profile", userHandler.GetProfileUser,
		rateLimiter.RateLimiter(RateLimitProfileView, RateLimitWindowOneMinute))

	userGroup.PATCH("/update-profile", userHandler.UpdateDataUser,
		rateLimiter.RateLimiter(RateLimitProfileUpdate, RateLimitWindowOneMinute))

	userGroup.PATCH("/change-password", userHandler.ChangePassword,
		rateLimiter.RateLimiter(RateLimitPasswordChange, RateLimitWindowFifteenMins))

	userGroup.POST("/upload-avatar", uploadHandler.UploadImage,
		rateLimiter.RateLimiter(RateLimitUploadAvatar, RateLimitMaxRequestsShort))

}
