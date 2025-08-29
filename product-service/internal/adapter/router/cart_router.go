package router

import (
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler"

	"github.com/labstack/echo/v4"
)

func CartRouter(e *echo.Echo, cartHandler handler.CartHandlerInterface, mid adapter.MiddlewareAdapterInterface) {

	cart := e.Group("api/v1/auth", mid.CheckToken(), mid.CheckRole("Customer"))
	cart.POST("/cart", cartHandler.AddToCart)
	cart.GET("/cart", cartHandler.GetCart)
}
