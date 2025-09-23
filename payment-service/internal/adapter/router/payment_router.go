package router

import (
	"payment-service/config"
	"payment-service/internal/adapter/handler"
	"payment-service/internal/adapter/middleware"
	"payment-service/internal/core/service"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func NewRouterPaymentService(
	e *echo.Echo,
	paymentHandler handler.PaymentHandlerInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
	redisClient *redis.Client,
) {
	mid := middleware.NewmiddlewareAuth(cfg, jwtService, redisClient)

	api := e.Group("/api/v1", mid.CheckRole("Customer", "Super Admin"), mid.CheckToken())
	api.POST("/payments/web-hook", paymentHandler.MidtransWebHook)
	api.GET("/payments", paymentHandler.GetAllAdmin)
	api.GET("/payments/:paymentID", paymentHandler.GetDetail)

	apiCustomer := e.Group("/api/v1/auth", mid.CheckRole("Customer"), mid.CheckToken())
	apiCustomer.POST("payments", paymentHandler.Create)
	apiCustomer.GET("/payments", paymentHandler.GetAllCustomer)
	apiCustomer.GET("/payments/:paymentID", paymentHandler.GetDetail)
}
