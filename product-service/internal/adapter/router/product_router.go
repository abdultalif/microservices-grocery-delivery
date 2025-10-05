package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func ProductRouter(
	e *echo.Echo,
	productHandler handler.ProductHandlerInterface,
	uploadHandler handler.UploadImageHandlerInterface,
	mid adapter.MiddlewareAdapterInterface,
	rate middleware.RateLimiterMiddlewareInterface,
) {

	// Public product endpoints
	product := e.Group("api/v1/products")

	product.GET("/shop", productHandler.GetAllShop,
		rate.RateLimiter(RateLimitGetProductsShop, RateLimitWindowOneMinute))

	product.GET("/home", productHandler.GetAllHome,
		rate.RateLimiter(RateLimitGetProductsHome, RateLimitWindowOneMinute))

	product.GET("/home/:productID", productHandler.GetDetailHome,
		rate.RateLimiter(RateLimitGetProductDetailHome, RateLimitWindowOneMinute))

	// Admin product endpoints
	productAdmin := e.Group("api/v1/admin",
		mid.CheckToken(),
		mid.CheckRole("Super Admin"),
	)

	productAdmin.GET("/products", productHandler.GetAllAdmin,
		rate.RateLimiter(RateLimitGetAllProducts, RateLimitWindowOneMinute))

	productAdmin.POST("/products", productHandler.Create,
		rate.RateLimiter(RateLimitCreateProduct, RateLimitWindowOneMinute))

	productAdmin.PUT("/products/:productID", productHandler.Update,
		rate.RateLimiter(RateLimitUpdateProduct, RateLimitWindowOneMinute))

	productAdmin.GET("/products/:productID", productHandler.GetByID,
		rate.RateLimiter(RateLimitGetProductByID, RateLimitWindowOneMinute))

	productAdmin.DELETE("/products/:productID", productHandler.Delete,
		rate.RateLimiter(RateLimitDeleteProduct, RateLimitWindowOneMinute))

	productAdmin.PATCH("/upload-photo/:productID", uploadHandler.UploadImage,
		rate.RateLimiter(RateLimitUploadProductImg, RateLimitWindowOneMinute))
}
