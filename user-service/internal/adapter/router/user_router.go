package router

import (
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/middleware"

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
		rateLimiter.RateLimiter(30, 60))
	userGroup.PATCH("/update-profile", userHandler.UpdateDataUser,
		rateLimiter.RateLimiter(5, 60))
	userGroup.PATCH("/change-password", userHandler.ChangePassword,
		rateLimiter.RateLimiter(3, 900))
	userGroup.POST("/upload-avatar", uploadHandler.UploadImage,
		rateLimiter.RateLimiter(3, 600))

}
