package middleware

import (
	"fmt"
	"net/http"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type RateLimiterMiddlewareInterface interface {
	RateLimiter(limit int, windowSeconds int) echo.MiddlewareFunc
}

type rateLimiterMiddleware struct {
	redisClient *redis.Client
}

func NewRateLimiterMiddleware(redisClient *redis.Client) RateLimiterMiddlewareInterface {
	return &rateLimiterMiddleware{
		redisClient: redisClient,
	}
}

func (r *rateLimiterMiddleware) RateLimiter(limit int, windowSeconds int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			var key string
			user, ok := c.Get("user").(entity.JwtUserData)
			if ok && user.UserID != 0 {
				key = fmt.Sprintf("ratelimit:%s:user:%d", c.Path(), user.UserID)
			} else {
				ip := c.RealIP()
				key = fmt.Sprintf("ratelimit:%s:ip:%s", c.Path(), ip)
			}

			val, err := r.redisClient.Incr(c.Request().Context(), key).Result()
			if err != nil {
				log.Errorf("[rateLimiterMiddleware-3] RateLimiter: %v", err)
				return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, "internal rate limiter error"))
			}

			if val == 1 {
				log.Infof("[rateLimiterMiddleware-4] RateLimiter: setting key %s to expire in %d seconds", key, windowSeconds)
				r.redisClient.Expire(c.Request().Context(), key, time.Duration(windowSeconds)*time.Second)
			}

			if val > int64(limit) {
				log.Errorf("[rateLimiterMiddleware-5] RateLimiter: too many requests")
				return c.JSON(http.StatusTooManyRequests, response.APIResponseError(http.StatusTooManyRequests, "too many requests"))
			}

			return next(c)
		}
	}
}
