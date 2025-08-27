package router

import (
	"order-service/config"
	"order-service/internal/adapter/handler"
	"order-service/internal/adapter/middleware"
	"order-service/internal/core/service"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func OrderRouter(e *echo.Echo, orderHandler handler.OrderHandlerInterface, cfg *config.Config, JwtService service.JwtServiceInterface, redis *redis.Client) {

	mid := middleware.NewmiddlewareAuth(cfg, JwtService, redis)

	orderCustomer := e.Group("/api/v1/auth", mid.CheckToken(), mid.CheckRole("Customer"))
	orderCustomer.POST("/orders", orderHandler.Create)
	orderCustomer.GET("/orders", orderHandler.GetAllCustomer)
	orderCustomer.GET("/orders/:orderID", orderHandler.GetDetailCustomer)
	orderCustomer.GET("/orders/:orderCode/code", orderHandler.GetOrderByOrderCode)

	orderAdmin := e.Group("/api/v1/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))
	orderAdmin.GET("/orders", orderHandler.GetAll)
	orderAdmin.DELETE("/orders/:orderID", orderHandler.DeleteOrderByID)
	orderAdmin.GET("/orders/:orderID", orderHandler.GetByID)
	orderAdmin.PUT("/orders/:orderID/status", orderHandler.UpdateStatus)
}
