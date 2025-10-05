package router

import (
	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/service"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func NewRouterUserService(
	e *echo.Echo,
	authHandler handler.AuthHandlerInterface,
	customerHandler handler.CustomerHandlerInterface,
	roleHandler handler.RoleHandlerInterface,
	userHandler handler.UserHandlerInterface,
	uploadHandler handler.UploadImageHandlerInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
	rateLimiter middleware.RateLimiterMiddlewareInterface,
	redisClient *redis.Client,
	oauthHandler handler.OAuthHandlerInterface,
) {
	mid := middleware.NewAuthMiddleware(cfg, jwtService, redisClient)

	OauthRouter(e, oauthHandler, mid, rateLimiter)
	AuthRouter(e, authHandler, rateLimiter)
	CustomerRouter(e, customerHandler, mid, rateLimiter)
	RoleRouter(e, roleHandler, mid, rateLimiter)
	UserRouter(e, userHandler, uploadHandler, mid, rateLimiter)
}
