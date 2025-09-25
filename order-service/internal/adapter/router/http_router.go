package router

import (
	"order-service/config"
	"order-service/internal/adapter/handler"
	"order-service/internal/adapter/middleware"
	"order-service/internal/core/service"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func OrderRouter(
	e *echo.Echo,
	orderHandler handler.OrderHandlerInterface,
	cfg *config.Config,
	JwtService service.JwtServiceInterface,
	redis *redis.Client,
	rate middleware.RateLimiterMiddlewareInterface,
	midDistance middleware.MiddlewareDistanceInterface,
	mid middleware.MiddlewareAuthInterface) {

	// Public
	e.GET("/api/v1/public/orders/:orderCode/code", orderHandler.GetPublicOrderIDByOrderCode)

	// Customer
	orderCustomer := e.Group("/api/v1/auth", mid.CheckToken(), mid.CheckRole("Customer"))

	orderCustomer.POST("/orders", orderHandler.Create, midDistance.DistanceCheck(),
		rate.RateLimiter(RateLimitCreateOrder, RateLimitWindowOneMinute))

	orderCustomer.GET("/orders", orderHandler.GetAllCustomer,
		rate.RateLimiter(RateLimitGetAllCustomer, RateLimitWindowOneMinute))

	orderCustomer.GET("/orders/:orderID", orderHandler.GetDetailCustomer,
		rate.RateLimiter(RateLimitGetDetailCustomer, RateLimitWindowOneMinute))

	orderCustomer.GET("/orders/:orderCode/code", orderHandler.GetOrderByOrderCode,
		rate.RateLimiter(RateLimitGetOrderByCode, RateLimitWindowOneMinute))

	// Admin
	orderAdmin := e.Group("/api/v1/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))

	orderAdmin.GET("/orders", orderHandler.GetAll,
		rate.RateLimiter(RateLimitGetAllAdmin, RateLimitWindowOneMinute))

	orderAdmin.DELETE("/orders/:orderID", orderHandler.DeleteOrderByID,
		rate.RateLimiter(RateLimitDeleteOrderByID, RateLimitWindowOneMinute))

	orderAdmin.GET("/orders/:orderID", orderHandler.GetByID,
		rate.RateLimiter(RateLimitGetOrderByIDAdmin, RateLimitWindowOneMinute))

	orderAdmin.PUT("/orders/:orderID/status", orderHandler.UpdateStatus,
		rate.RateLimiter(RateLimitUpdateOrderStatus, RateLimitWindowOneMinute))
}
