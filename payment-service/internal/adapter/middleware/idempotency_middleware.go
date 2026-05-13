package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	IdempotencyKeyHeader  = "Idempotency-Key"
	IdempotencyKeyPrefix  = "idempotency:"
	IdempotencyKeyTTL     = 24 * time.Hour
	IdempotencyLockTTL    = 1 * time.Minute
	IdempotencyLockSuffix = ":lock"
)

type CachedResponse struct {
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
}

func IdempotencyMiddleware(cfg config.Config) echo.MiddlewareFunc {
	redisClient, _ := cfg.NewRedisClient()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Ambil Idempotency-Key dari header
			idempotencyKey := c.Request().Header.Get(IdempotencyKeyHeader)
			if idempotencyKey == "" {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"code":    http.StatusBadRequest,
					"message": "Idempotency-Key header is required",
					"data":    nil,
				})
			}

			redisKey := IdempotencyKeyPrefix + idempotencyKey
			lockKey := redisKey + IdempotencyLockSuffix

			// Cek apakah response sudah ada di Redis (request sebelumnya sudah selesai)
			cached, err := redisClient.Get(ctx, redisKey).Result()
			if err == nil {
				// Cache hit → kembalikan response yang sama tanpa proses ulang
				log.Infof("[IdempotencyMiddleware] Cache hit for key: %s", idempotencyKey)

				var cachedResponse CachedResponse
				if err := json.Unmarshal([]byte(cached), &cachedResponse); err != nil {
					log.Errorf("[IdempotencyMiddleware] Failed to unmarshal cached response: %v", err)
					return next(c) // fallback: proses normal jika cache corrupt
				}

				c.Response().Header().Set("X-Idempotent-Replayed", "true")
				return c.JSONBlob(cachedResponse.StatusCode, cachedResponse.Body)
			}

			// Cek apakah sedang diproses (concurrent request dengan key yang sama)
			// Gunakan SET NX (set if not exists) sebagai distributed lock
			locked, err := redisClient.SetNX(ctx, lockKey, "1", IdempotencyLockTTL).Result()
			if err != nil {
				log.Errorf("[IdempotencyMiddleware] Redis lock error: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"code":    http.StatusInternalServerError,
					"message": "Internal server error",
					"data":    nil,
				})
			}

			if !locked {
				// Request lain sedang memproses key yang sama → tolak
				log.Warnf("[IdempotencyMiddleware] Duplicate concurrent request for key: %s", idempotencyKey)
				return c.JSON(http.StatusConflict, map[string]interface{}{
					"success": false,
					"code":    http.StatusConflict,
					"message": "Request with this Idempotency-Key is currently being processed, please wait",
					"data":    nil,
				})
			}

			// Pastikan lock selalu dilepas setelah handler selesai
			defer releaseLock(ctx, cfg, lockKey)

			// Inject custom ResponseWriter untuk menangkap response
			crw := newCaptureResponseWriter(c.Response())
			c.Response().Writer = crw

			// Jalankan handler asli
			handlerErr := next(c)

			// Simpan response ke Redis setelah handler sukses
			if handlerErr == nil && crw.statusCode < 500 {
				cachePayload := CachedResponse{
					StatusCode: crw.statusCode,
					Body:       crw.body,
				}

				cacheBytes, err := json.Marshal(cachePayload)
				if err != nil {
					log.Errorf("[IdempotencyMiddleware] Failed to marshal response for caching: %v", err)
				} else {
					if err := redisClient.Set(ctx, redisKey, cacheBytes, IdempotencyKeyTTL).Err(); err != nil {
						log.Errorf("[IdempotencyMiddleware] Failed to cache response: %v", err)
					} else {
						log.Infof("[IdempotencyMiddleware] Response cached for key: %s", idempotencyKey)
					}
				}
			}

			return handlerErr
		}
	}
}

// releaseLock menghapus distributed lock dari Redis
func releaseLock(ctx context.Context, config config.Config, lockKey string) {
	client, _ := config.NewRedisClient()
	if err := client.Del(ctx, lockKey).Err(); err != nil {
		log.Errorf("[IdempotencyMiddleware] Failed to release lock %s: %v", lockKey, err)
	}
}

// -------------------------------------------------------
// captureResponseWriter: menangkap status code & body
// yang ditulis oleh handler agar bisa disimpan ke Redis
// -------------------------------------------------------

type captureResponseWriter struct {
	echo.Response
	wrappedWriter http.ResponseWriter
	statusCode    int
	body          []byte
}

func newCaptureResponseWriter(r *echo.Response) *captureResponseWriter {
	return &captureResponseWriter{
		Response:      *r,
		wrappedWriter: r.Writer,
		statusCode:    http.StatusOK,
	}
}

func (crw *captureResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.wrappedWriter.WriteHeader(code)
}

func (crw *captureResponseWriter) Write(b []byte) (int, error) {
	crw.body = append(crw.body, b...)
	return crw.wrappedWriter.Write(b)
}

func (crw *captureResponseWriter) Header() http.Header {
	return crw.wrappedWriter.Header()
}
