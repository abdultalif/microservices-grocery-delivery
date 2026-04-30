package notification

import (
	"os"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	notificationGroup := g.Group("/auth/notifications")

	// Notification routes
	notificationGroup.GET("", proxyHandler)
	notificationGroup.GET("/:notificationID", proxyHandler)
	notificationGroup.PUT("/:notificationID", proxyHandler)

	// WebSocket route
	notificationGroup.GET("/ws", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8085"
	}

	path := c.Request().URL.Path

	return proxy.ForwardRequest(c, notificationServiceURL+path)
}
