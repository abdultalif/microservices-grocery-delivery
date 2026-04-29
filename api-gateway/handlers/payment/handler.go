package payment

import (
	"os"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {

	g.POST("/payments/web-hook", proxyHandler)

	// Payment Admin routes
	paymentAdmin := g.Group("/payments")
	paymentAdmin.GET("", proxyHandler)
	paymentAdmin.GET("/:paymentID", proxyHandler)

	paymentCustomer := g.Group("/auth/payments")
	paymentCustomer.POST("", proxyHandler)
	paymentCustomer.GET("", proxyHandler)
	paymentCustomer.GET("/:paymentID", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "http://localhost:8084"
	}

	path := c.Request().URL.Path

	return proxy.ForwardRequest(c, paymentServiceURL+path)
}
