package user

import (
	"os"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterPublicRoutes(g *echo.Group) {
	userGroup := g.Group("/auth")

	// Public routes authentication
	userGroup.POST("", proxyHandler)
	userGroup.POST("/login", proxyHandler)
	userGroup.GET("/verify-account", proxyHandler)
	userGroup.POST("/resend-verification", proxyHandler)
	userGroup.POST("/forgot-password", proxyHandler)
	userGroup.GET("/validate-forgot-token", proxyHandler)
	userGroup.PATCH("/reset-password", proxyHandler)
	userGroup.POST("/service-token", proxyHandler)

	// Public routes OAutherization
	g.GET("/oauth/google/login", proxyHandler)
	g.GET("/oauth/google/login/callback", proxyHandler)
	g.GET("/oauth/google/register", proxyHandler)
	g.GET("/oauth/google/register/callback", proxyHandler)

}

func RegisterProtectedRoutes(g *echo.Group) {
	userGroup := g.Group("/user")

	// Protected routes
	userGroup.GET("/profile", proxyHandler)
	userGroup.PATCH("/update-profile", proxyHandler)
	userGroup.PATCH("/change-password", proxyHandler)
	userGroup.POST("/upload-avatar", proxyHandler)

	// Admin routes for managing customers
	customerAdmin := g.Group("/admin")
	customerAdmin.GET("/customers", proxyHandler)
	customerAdmin.POST("/customers", proxyHandler)
	customerAdmin.PATCH("/customers/:id", proxyHandler)
	customerAdmin.GET("/customers/:id", proxyHandler)
	customerAdmin.DELETE("/customers/:id", proxyHandler)
	customerAdmin.PUT("/customers/:id/location", proxyHandler)

	// OAuth routes untuk mengelola akun yang terhubung dengan penyedia OAuth (misalnya Google)
	oauthGroup := g.Group("/oauth")
	oauthGroup.POST("/link", proxyHandler)
	oauthGroup.PATCH("/set-password", proxyHandler)
	oauthGroup.DELETE("/unlink/:provider_id", proxyHandler)
	oauthGroup.GET("/logout", proxyHandler)

	// Roles and permissions management (admin only)
	roleAdmin := g.Group("/admin")
	roleAdmin.GET("/role", proxyHandler)
	roleAdmin.POST("/role", proxyHandler)
	roleAdmin.PATCH("/role/:id", proxyHandler)
	roleAdmin.GET("/role/:id", proxyHandler)
	roleAdmin.DELETE("/role/:id", proxyHandler)

}

func proxyHandler(c echo.Context) error {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8080"
	}

	path := c.Request().URL.Path

	return proxy.ForwardRequest(c, userServiceURL+path)
}
