package router

import (
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler"

	"github.com/labstack/echo/v4"
)

func CategoryRouter(e *echo.Echo, categoryHandler handler.CategoryHandlerInterface, mid adapter.MiddlewareAdapterInterface) {
	
	category := e.Group("api/v1/categories")
	category.Use(mid.CheckToken(), mid.CheckRole("Customer"))
	category.GET("/home", categoryHandler.GetAllHome)
	category.GET("/shop", categoryHandler.GetAllShop)

	
	categoryAdmin := e.Group("api/v1/admin")
	categoryAdmin.Use(mid.CheckToken(), mid.CheckRole("Super Admin"))
	categoryAdmin.PATCH("/categories/:categoryId", categoryHandler.Update)
	categoryAdmin.GET("/categories", categoryHandler.GetAll)
	categoryAdmin.GET("/categories/:slug/slug", categoryHandler.GetBySlug)
	categoryAdmin.POST("/categories", categoryHandler.Create)
	categoryAdmin.GET("/categories/:categoryId", categoryHandler.GetByID)
	categoryAdmin.DELETE("/categories/:categoryId", categoryHandler.Delete)
}
