package router

import (
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler"
	"product-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func CartRouter(e *echo.Echo, cartHandler handler.CartHandlerInterface, mid adapter.MiddlewareAdapterInterface, rate middleware.RateLimiterMiddlewareInterface) {

	cart := e.Group("api/v1/auth", mid.CheckToken(), mid.CheckRole("Customer"))

	cart.POST("/cart", cartHandler.AddToCart,
		rate.RateLimiter(RateLimitAddToCart, RateLimitWindowOneMinute))

	cart.GET("/cart", cartHandler.GetCart,
		rate.RateLimiter(RateLimitGetCart, RateLimitWindowOneMinute))

	cart.DELETE("/cart", cartHandler.RemoveFromCart,
		rate.RateLimiter(RateLimitRemoveFromCart, RateLimitWindowOneMinute))

	cart.DELETE("/cart/all", cartHandler.RemoveAllCart,
		rate.RateLimiter(RateLimitRemoveAllCart, RateLimitWindowOneMinute))
}
