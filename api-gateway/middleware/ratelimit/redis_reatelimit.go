package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

type RedisRateLimiter struct {
	redisClient *redis.Client
}

func NewRedisRateLimiter() *RedisRateLimiter {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return &RedisRateLimiter{redisClient: client}
}

func RedisMiddleware() echo.MiddlewareFunc {
	requestPersecond := getEnvAsFloat("RATE_LIMIT_REQUESTS_PER_SECOND", 10.0)
	burstSize := getEnvAsInt("RATE_LIMIT_BURST_SIZE", 20)
	windowSize := getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60)

	limiter := NewRedisRateLimiter()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/health" {
				return next(c)
			}

			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}

			allowed, remaining, resetTime, err := limiter.checkRateLimit(ip, int(requestPersecond), burstSize, windowSize)
			if err != nil {
				return next(c)
			}

			if !allowed {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"status":      "error",
					"message":     "Rate limit exceeded. Please try again later.",
					"code":        "RATE_LIMIT_EXCEEDED",
					"remaining":   remaining,
					"reset_time":  resetTime,
					"limit":       int(requestPersecond * float64(windowSize)),
					"window_size": windowSize,
				})
			}

			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(int(requestPersecond*float64(windowSize))))

			return next(c)
		}
	}
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func (rl *RedisRateLimiter) checkRateLimit(ip string, requestsPerSecond, burstSize, windowSize int) (bool, int, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", ip)

	now := time.Now().Unix()
	windowStart := now - int64(windowSize)

	pipe := rl.redisClient.Pipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	countCmd := pipe.ZCard(ctx, key)

	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: now,
	})

	pipe.Expire(ctx, key, time.Duration(windowSize)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := int(countCmd.Val())
	limit := int(float64(requestsPerSecond) * float64(windowSize))

	if count > limit {
		resetTime := now + int64(windowSize)
		return false, 0, resetTime, nil
	}

	remaining := limit - count
	resetTime := now + int64(windowSize)

	return true, remaining, resetTime, nil
}
