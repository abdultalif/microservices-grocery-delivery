package middleware

import (
	"fmt"
	"net/http"
	"time"
	"user-service/internal/adapter/handler/response"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
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
			ip := c.RealIP()
			key := fmt.Sprintf("ratelimit:%s:%s", c.Path(), ip)

			val, err := r.redisClient.Incr(c.Request().Context(), key).Result()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.ResponseAPI(false, http.StatusInternalServerError, "internal rate limiter error", nil))
			}

			if val == 1 {
				r.redisClient.Expire(c.Request().Context(), key, time.Duration(windowSeconds)*time.Second)
			}

			if val > int64(limit) {
				return c.JSON(http.StatusTooManyRequests, response.ResponseAPI(false, http.StatusTooManyRequests, "too many requests", nil))
			}

			return next(c)
		}
	}
}
