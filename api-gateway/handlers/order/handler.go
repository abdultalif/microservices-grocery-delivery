package order

import (
	"os"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	orderCustomer := g.Group("/auth/orders")

	// Order Customer routes
	orderCustomer.POST("", proxyHandler)
	orderCustomer.GET("", proxyHandler)
	orderCustomer.GET("/:orderID", proxyHandler)
	orderCustomer.GET("/:orderCode/code", proxyHandler)

	// Order Admin routes
	orderAdmin := g.Group("/admin/orders")
	orderAdmin.GET("", proxyHandler)
	orderAdmin.GET("/:orderID", proxyHandler)
	orderAdmin.DELETE("/:orderID", proxyHandler)
	orderAdmin.PUT("/:orderID/status", proxyHandler)

	orderPublic := g.Group("/public/orders")
	orderPublic.GET("/:orderID/code", proxyHandler)
	orderPublic.PUT("/:orderID/status", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://localhost:8083"
	}

	path := c.Request().URL.Path

	return proxy.ForwardRequest(c, orderServiceURL+path)
}
