package payment

import (
	"os"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	paymentGroup := g.Group("/payments")

	// Payment routes
	paymentGroup.GET("", proxyHandler)
	paymentGroup.GET("/:id", proxyHandler)
	paymentGroup.POST("", proxyHandler)
	paymentGroup.PUT("/:id", proxyHandler)
	paymentGroup.POST("/callback", proxyHandler)
	paymentGroup.GET("/:id/status", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "http://localhost:8084"
	}

	// Hapus prefix /api/v1 dari path
	path := strings.TrimPrefix(c.Request().URL.Path, "/api/v1")
	if path == "" {
		path = "/"
	}

	return proxy.ForwardRequest(c, paymentServiceURL+path)
}
