package http

import (
	"notification-service/internal/config"
	"notification-service/internal/handlers"
	"notification-service/internal/middleware"
	"notification-service/internal/services"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func NotifRouter(e *echo.Echo, notifHandler handlers.NotifHandlerInterface, cfg *config.Config, jwtService services.JwtServiceInterface, redis *redis.Client, midRateLimiter middleware.RateLimiterMiddlewareInterface, midAuth middleware.MiddlewareAuthInterface) {

	authNotif := e.Group("/api/v1/auth", midAuth.CheckRole("Customer"), midAuth.CheckToken())

	authNotif.GET("/notifications", notifHandler.GetAll, midRateLimiter.RateLimiter(config.RateLimitGetAllNotifications, config.RateLimitWindowOneMinute))

	authNotif.GET("/notifications/:notificationID", notifHandler.GetByID, midRateLimiter.RateLimiter(config.RateLimitGetNotificationByID, config.RateLimitWindowOneMinute))

}
