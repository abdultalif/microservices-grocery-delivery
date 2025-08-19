package router

import (
	"product-service/config"
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler"
	"product-service/internal/core/service"

	"github.com/labstack/echo/v4"
)

func NewRouter(
	e *echo.Echo,
	categoryHandler handler.CategoryHandlerInterface,
	productHandler handler.ProductHandlerInterface,
	uploadHandler handler.UploadImageHandlerInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
) {
	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)

	CategoryRouter(e, categoryHandler, mid)
	ProductRouter(e, productHandler, uploadHandler, mid)
}
