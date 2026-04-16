package handlers

import (
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/handlers/notification"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/handlers/order"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/handlers/payment"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/handlers/product"
	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/handlers/user"

	"github.com/labstack/echo/v4"
)

func RegisterPublicRoutes(g *echo.Group) {
	user.RegisterPublicRoutes(g)
	product.RegisterPublicRoutes(g)
}

func RegisterProtectedRoutes(g *echo.Group) {
	user.RegisterProtectedRoutes(g)
	product.RegisterProtectedRoutes(g)

	order.RegisterRoutes(g)
	payment.RegisterRoutes(g)

	notification.RegisterRoutes(g)
}

func RegisterAllRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	RegisterPublicRoutes(g)

	protected := g.Group("")
	protected.Use(jwtMiddleware)
	RegisterProtectedRoutes(protected)
}
