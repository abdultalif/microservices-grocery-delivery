package router

import (
	"product-service/config"
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler"
	"product-service/internal/adapter/middleware"
	"product-service/internal/core/service"

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
