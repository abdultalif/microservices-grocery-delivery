package router

import (
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func CustomerRouter(
	e *echo.Echo,
	customerHandler handler.CustomerHandlerInterface,
	mid middleware.AuthMiddlewareInterface,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
) {
	CustomerAdmin := e.Group("/api/v1/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))

	CustomerAdmin.GET("/customers", customerHandler.GetCustomerAll,
		rateLimiter.RateLimiter(30, 60))

	CustomerAdmin.POST("/customers", customerHandler.CreateCustomer,
		rateLimiter.RateLimiter(10, 60))

	CustomerAdmin.PATCH("/customers/:id", customerHandler.UpdateCustomer,
		rateLimiter.RateLimiter(10, 60))

	CustomerAdmin.GET("/customers/:id", customerHandler.GetCustomerByID,
		rateLimiter.RateLimiter(20, 60))

	CustomerAdmin.DELETE("/customers/:id", customerHandler.DeleteCustomer,
		rateLimiter.RateLimiter(5, 60))
}
