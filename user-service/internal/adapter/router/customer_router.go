package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/middleware"

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
		rateLimiter.RateLimiter(RateLimitCustomerViewAll, RateLimitWindowOneMinute))

	CustomerAdmin.POST("/customers", customerHandler.CreateCustomer,
		rateLimiter.RateLimiter(RateLimitCustomerCreate, RateLimitWindowOneMinute))

	CustomerAdmin.PATCH("/customers/:id", customerHandler.UpdateCustomer,
		rateLimiter.RateLimiter(RateLimitCustomerUpdate, RateLimitWindowOneMinute))

	CustomerAdmin.GET("/customers/:id", customerHandler.GetCustomerByID,
		rateLimiter.RateLimiter(RateLimitCustomerViewByID, RateLimitWindowOneMinute))

	CustomerAdmin.DELETE("/customers/:id", customerHandler.DeleteCustomer,
		rateLimiter.RateLimiter(RateLimitCustomerDelete, RateLimitWindowOneMinute))

	CustomerAdmin.PUT("/customers/:id/location", customerHandler.UpdateLocationCustomer,
		rateLimiter.RateLimiter(RateLimitCustomerUpdateLocation, RateLimitWindowOneMinute))
}
