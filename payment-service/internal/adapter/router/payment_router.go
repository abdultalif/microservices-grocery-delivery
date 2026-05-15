package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/service"

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

	// e.POST("/api/payments/web-hook", paymentHandler.MidtransWebHook)
	api := e.Group("api/v1", mid.CheckToken(), mid.CheckRole("Super Admin"))
	api.GET("/payments", paymentHandler.GetAllAdmin)
	api.GET("/payments/:paymentID", paymentHandler.GetDetail)

	apiCustomer := e.Group("api/v1/auth", mid.CheckToken(), mid.CheckRole("Customer"))
	apiCustomer.POST("/payments", paymentHandler.Create, middleware.IdempotencyMiddleware(*cfg))
	apiCustomer.GET("/payments", paymentHandler.GetAllCustomer)
	apiCustomer.GET("/payments/:paymentID", paymentHandler.GetDetail)
	apiCustomer.POST("/payments/cancel", paymentHandler.CancelPayment)
}
