package product

import (
	"os"

	"github.com/abdultalif/microservices-grocery-delivery/api-gateway/utils/proxy"

	"github.com/labstack/echo/v4"
)

func RegisterPublicRoutes(g *echo.Group) {
	productGroup := g.Group("/products")
	categoryGroup := g.Group("/categories")

	// Public routes
	productGroup.GET("/shop", proxyHandler)
	productGroup.GET("/home", proxyHandler)
	productGroup.GET("/home/:productID", proxyHandler)

	categoryGroup.GET("/home", proxyHandler)
	categoryGroup.GET("/shop", proxyHandler)

}

func RegisterProtectedRoutes(g *echo.Group) {
	productGroup := g.Group("/admin/products")
	cartGroup := g.Group("/auth/cart")
	categoryGroup := g.Group("/admin/categories")

	// Protected category routes
	categoryGroup.PATCH("/:categoryID", proxyHandler)
	categoryGroup.GET("", proxyHandler)
	categoryGroup.GET("/:slug/slug", proxyHandler)
	categoryGroup.POST("", proxyHandler)
	categoryGroup.GET("/:categoryID", proxyHandler)
	categoryGroup.DELETE("/:categoryID", proxyHandler)

	// Protected product routes
	productGroup.GET("", proxyHandler)
	productGroup.POST("", proxyHandler)
	productGroup.GET("/:productID", proxyHandler)
	productGroup.PUT("/:productID", proxyHandler)
	productGroup.DELETE("/:productID", proxyHandler)
	productGroup.PATCH("/upload-photo/:productID", proxyHandler)

	// Cart routes
	cartGroup.GET("", proxyHandler)
	cartGroup.POST("", proxyHandler)
	cartGroup.DELETE("", proxyHandler)
	// cartGroup.PUT("/:id", proxyHandler)
	cartGroup.DELETE("/all", proxyHandler)
}

func proxyHandler(c echo.Context) error {
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8082"
	}

	path := c.Request().URL.Path

	return proxy.ForwardRequest(c, productServiceURL+path)
}
