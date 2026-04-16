package user

import (
	"os"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterPublicRoutes(g *echo.Group) {
	userGroup := g.Group("/users")

	// Public routes
	userGroup.POST("/register", proxyHandler)
	userGroup.POST("/signin", proxyHandler)
	userGroup.POST("/verify-email", proxyHandler)
	userGroup.POST("/resend-verification", proxyHandler)
	userGroup.POST("/forgot-password", proxyHandler)
	userGroup.POST("/reset-password", proxyHandler)
}

func RegisterProtectedRoutes(g *echo.Group) {
	userGroup := g.Group("/users")

	// Protected routes
	userGroup.GET("/profile", proxyHandler)
	userGroup.PUT("/profile", proxyHandler)
	userGroup.PUT("/password", proxyHandler)
	userGroup.POST("/upload-avatar", proxyHandler)
	userGroup.GET("/roles", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8081"
	}

	// Hapus prefix /api/v1/users dari path
	path := strings.TrimPrefix(c.Request().URL.Path, "/api/v1/users")
	if path == "" {
		path = "/"
	}

	return proxy.ForwardRequest(c, userServiceURL+path)
}
