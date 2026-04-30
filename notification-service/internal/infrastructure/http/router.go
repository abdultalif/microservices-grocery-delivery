package http

import (
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/config"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/handlers"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/middleware"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/services"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func NotifRouter(e *echo.Echo, notifHandler handlers.NotifHandlerInterface, wsHandler handlers.WebSocketHandlerInterface, cfg *config.Config, jwtService services.JwtServiceInterface, redis *redis.Client, midRateLimiter middleware.RateLimiterMiddlewareInterface, midAuth middleware.MiddlewareAuthInterface) {

	authNotif := e.Group("/api/v1/auth", midAuth.CheckToken(), midAuth.CheckRole("Customer"))

	authNotif.GET("/notifications", notifHandler.GetAll, midRateLimiter.RateLimiter(config.RateLimitGetAllNotifications, config.RateLimitWindowOneMinute))

	authNotif.GET("/notifications/:notificationID", notifHandler.GetByID, midRateLimiter.RateLimiter(config.RateLimitGetNotificationByID, config.RateLimitWindowOneMinute))

	authNotif.PUT("/notifications/:notificationID", notifHandler.MarkAsRead, midRateLimiter.RateLimiter(config.RateLimitGetNotificationByID, config.RateLimitWindowOneMinute))

	e.GET("/api/v1/ws", wsHandler.WebSocketHandler)

}
