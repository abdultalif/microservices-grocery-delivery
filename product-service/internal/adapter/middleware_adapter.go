package adapter

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler/response"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/service"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareAdapterInterface interface {
	CheckToken() echo.MiddlewareFunc
	CheckRole(allowedRoles ...string) echo.MiddlewareFunc
}

type middlewareAdapter struct {
	cfg        *config.Config
	jwtService service.JwtServiceInterface
	redis      *redis.Client
}

// CheckToken implements MiddlewareAdapterInterface.
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			respErr := response.ResponseDefault{}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: %s", "missing or invalid token")
				respErr.Message = "missing or invalid token"
				respErr.Data = nil
				respErr.Code = http.StatusUnauthorized
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			_, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", err.Error())
				respErr.Message = "invalid token"
				respErr.Data = nil
				respErr.Code = http.StatusUnauthorized
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			getSession, err := m.redis.Get(c.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-3] CheckToken: session not found")
				respErr.Message = "session expired or not found"
				respErr.Code = http.StatusUnauthorized
				respErr.Success = false
				respErr.Data = nil
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			jwtUserData := entity.JwtUserData{}
			err = json.Unmarshal([]byte(getSession), &jwtUserData)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-4] CheckToken: failed to parse user data")
				respErr.Message = "Failed to parse user data"
				respErr.Code = http.StatusInternalServerError
				respErr.Success = false
				respErr.Data = nil
				return c.JSON(http.StatusInternalServerError, respErr)
			}

			c.Set("user", jwtUserData)
			return next(c)
		}
	}
}

// CheckRole implements MiddlewareAdapterInterface.
func (m *middlewareAdapter) CheckRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			respErr := response.ResponseDefault{}
			userData, ok := c.Get("user").(entity.JwtUserData)
			if !ok {
				log.Errorf("[MiddlewareAdapter-5] CheckRole: user not found in context")
				respErr.Message = "unauthorized"
				respErr.Code = http.StatusUnauthorized
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			for _, role := range allowedRoles {
				if userData.Role == role {
					return next(c)
				}
			}

			requiredRoles := strings.Join(allowedRoles, ", ")
			log.Warnf("[MiddlewareAdapter-6] CheckRole: access denied for role %s, required: %s", userData.Role, requiredRoles)
			respErr.Message = "User role mismatch: required " + requiredRoles
			respErr.Code = http.StatusForbidden
			return c.JSON(http.StatusForbidden, respErr)
		}
	}
}

func NewMiddlewareAdapter(cfg *config.Config, jwtService service.JwtServiceInterface, redis *redis.Client) MiddlewareAdapterInterface {
	return &middlewareAdapter{
		cfg:        cfg,
		jwtService: jwtService,
		redis:      redis,
	}
}
