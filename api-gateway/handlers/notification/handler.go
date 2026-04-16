package notification

import (
	"os"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	notificationGroup := g.Group("/notifications")

	// Notification routes
	notificationGroup.GET("", proxyHandler)
	notificationGroup.GET("/:id", proxyHandler)
	notificationGroup.POST("", proxyHandler)
	notificationGroup.PUT("/:id", proxyHandler)
	notificationGroup.DELETE("/:id", proxyHandler)
	notificationGroup.PUT("/:id/read", proxyHandler)
	notificationGroup.GET("/unread", proxyHandler)

	// WebSocket route
	notificationGroup.GET("/ws", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8085"
	}

	// Hapus prefix /api/v1 dari path
	path := strings.TrimPrefix(c.Request().URL.Path, "/api/v1")
	if path == "" {
		path = "/"
	}

	return proxy.ForwardRequest(c, notificationServiceURL+path)
}
