package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func CategoryRouter(e *echo.Echo, categoryHandler handler.CategoryHandlerInterface, mid adapter.MiddlewareAdapterInterface, rate middleware.RateLimiterMiddlewareInterface) {

	category := e.Group("api/v1/categories")

	category.GET("/home", categoryHandler.GetAllHome,
		rate.RateLimiter(RateLimitGetCategoriesHome, RateLimitWindowOneMinute))

	category.GET("/shop", categoryHandler.GetAllShop,
		rate.RateLimiter(RateLimitGetCategoriesShop, RateLimitWindowOneMinute))

	categoryAdmin := e.Group("api/v1/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))

	categoryAdmin.PATCH("/categories/:categoryId", categoryHandler.Update,
		rate.RateLimiter(RateLimitUpdateCategory, RateLimitWindowOneMinute))

	categoryAdmin.GET("/categories", categoryHandler.GetAll,
		rate.RateLimiter(RateLimitGetAllCategories, RateLimitWindowOneMinute))

	categoryAdmin.GET("/categories/:slug/slug", categoryHandler.GetBySlug,
		rate.RateLimiter(RateLimitGetCategoryBySlug, RateLimitWindowOneMinute))

	categoryAdmin.POST("/categories", categoryHandler.Create,
		rate.RateLimiter(RateLimitCreateCategory, RateLimitWindowOneMinute))

	categoryAdmin.GET("/categories/:categoryId", categoryHandler.GetByID,
		rate.RateLimiter(RateLimitGetCategoryByID, RateLimitWindowOneMinute))

	categoryAdmin.DELETE("/categories/:categoryId", categoryHandler.Delete,
		rate.RateLimiter(RateLimitDeleteCategory, RateLimitWindowOneMinute))
}
