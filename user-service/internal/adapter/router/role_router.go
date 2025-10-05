package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/middleware"

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
		rateLimiter.RateLimiter(RateLimitRoleViewAll, RateLimitWindowOneMinute))

	RoleAdmin.POST("/role", roleHandler.CreateRole,
		rateLimiter.RateLimiter(RateLimitRoleCreate, RateLimitWindowOneMinute))

	RoleAdmin.DELETE("/role/:id", roleHandler.DeleteRole,
		rateLimiter.RateLimiter(RateLimitRoleDelete, RateLimitWindowOneMinute))

	RoleAdmin.GET("/role/:id", roleHandler.GetRoleByID,
		rateLimiter.RateLimiter(RateLimitRoleViewByID, RateLimitWindowOneMinute))

	RoleAdmin.PATCH("/role/:id", roleHandler.UpdateRole,
		rateLimiter.RateLimiter(RateLimitRoleUpdate, RateLimitWindowOneMinute))

}
