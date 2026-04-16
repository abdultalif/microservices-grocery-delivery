package product

import (
	"os"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterPublicRoutes(g *echo.Group) {
	productGroup := g.Group("/products")

	// Public routes
	productGroup.GET("/shop", proxyHandler)
	productGroup.GET("/home", proxyHandler)
	productGroup.GET("/home/:id", proxyHandler)
	productGroup.GET("/categories", proxyHandler)
	productGroup.GET("/categories/:id", proxyHandler)
	productGroup.GET("/search", proxyHandler)
}

func RegisterProtectedRoutes(g *echo.Group) {
	productGroup := g.Group("/admin/products")
	cartGroup := g.Group("/cart")

	// Protected product routes
	productGroup.GET("", proxyHandler)
	productGroup.POST("", proxyHandler)
	productGroup.PUT("/:id", proxyHandler)
	productGroup.DELETE("/:id", proxyHandler)
	productGroup.POST("/upload-image", proxyHandler)

	// Cart routes
	cartGroup.GET("", proxyHandler)
	cartGroup.POST("", proxyHandler)
	cartGroup.PUT("/:id", proxyHandler)
	cartGroup.DELETE("/:id", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8082"
	}

	// Hapus prefix /api/v1 dari path
	path := strings.TrimPrefix(c.Request().URL.Path, "/api/v1")
	if path == "" {
		path = "/"
	}

	return proxy.ForwardRequest(c, productServiceURL+path)
}
