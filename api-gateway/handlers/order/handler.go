package order

import (
	"os"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	orderGroup := g.Group("/orders")

	// Order routes
	orderGroup.GET("", proxyHandler)
	orderGroup.GET("/:id", proxyHandler)
	orderGroup.POST("", proxyHandler)
	orderGroup.PUT("/:id", proxyHandler)
	orderGroup.DELETE("/:id", proxyHandler)
	orderGroup.PUT("/:id/status", proxyHandler)
	orderGroup.GET("/:id/items", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://localhost:8083"
	}

	// Hapus prefix /api/v1 dari path
	path := strings.TrimPrefix(c.Request().URL.Path, "/api/v1")
	if path == "" {
		path = "/"
	}

	return proxy.ForwardRequest(c, orderServiceURL+path)
}
