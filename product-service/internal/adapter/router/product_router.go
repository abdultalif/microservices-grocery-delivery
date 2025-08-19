package router

import (
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler"

	"github.com/labstack/echo/v4"
)

func ProductRouter(e *echo.Echo, productHandler handler.ProductHandlerInterface, uploadHandler handler.UploadImageHandlerInterface, mid adapter.MiddlewareAdapterInterface) {
	
	product := e.Group("api/v1/products")
	product.Use(mid.CheckToken(), mid.CheckRole("Customer"))
	product.GET("/shop", productHandler.GetAllShop)
	product.GET("/home", productHandler.GetAllHome)
	product.GET("/home/:productID", productHandler.GetDetailHome)

	
	productAdmin := e.Group("api/v1/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))
	productAdmin.GET("/products", productHandler.GetAllAdmin)
	productAdmin.POST("/products", productHandler.Create)
	productAdmin.PUT("/products/:productID", productHandler.Update)
	productAdmin.GET("/products/:productID", productHandler.GetByID)
	productAdmin.DELETE("/products/:productID", productHandler.Delete)

	// ini upload image ke supabase
	productAdmin.PATCH("/upload-photo/:productID", uploadHandler.UploadImage)
}
