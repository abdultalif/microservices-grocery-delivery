package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/service"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func NewRouter(
	e *echo.Echo,
	categoryHandler handler.CategoryHandlerInterface,
	productHandler handler.ProductHandlerInterface,
	uploadHandler handler.UploadImageHandlerInterface,
	cartHandler handler.CartHandlerInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
	redis *redis.Client,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
) {
	mid := adapter.NewMiddlewareAdapter(cfg, jwtService, redis)

	CartRouter(e, cartHandler, mid, rateLimiter)
	CategoryRouter(e, categoryHandler, mid, rateLimiter)
	ProductRouter(e, productHandler, uploadHandler, mid, rateLimiter)
}
