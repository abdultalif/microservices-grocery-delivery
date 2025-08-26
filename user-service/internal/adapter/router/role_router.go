package router

import (
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func RoleRouter(
	e *echo.Echo,
	roleHandler handler.RoleHandlerInterface,
	mid middleware.AuthMiddlewareInterface,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
) {
	RoleAdmin := e.Group("/api/v1/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))

	RoleAdmin.GET("/role", roleHandler.GetAllRole,
		rateLimiter.RateLimiter(20, 60))

	RoleAdmin.POST("/role", roleHandler.CreateRole,
		rateLimiter.RateLimiter(3, 60))

	RoleAdmin.DELETE("/role/:id", roleHandler.DeleteRole,
		rateLimiter.RateLimiter(3, 60))

	RoleAdmin.GET("/role/:id", roleHandler.GetRoleByID,
		rateLimiter.RateLimiter(20, 60))

	RoleAdmin.PATCH("/role/:id", roleHandler.UpdateRole,
		rateLimiter.RateLimiter(5, 60))

}
