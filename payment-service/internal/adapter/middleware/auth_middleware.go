package middleware

import (
	"encoding/json"
	"net/http"
	"payment-service/config"
	"payment-service/internal/adapter/handler/response"
	"payment-service/internal/core/domain/entity"
	"payment-service/internal/core/service"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareAuthInterface interface {
	CheckToken() echo.MiddlewareFunc
	CheckRole(allowedRoles ...string) echo.MiddlewareFunc
}

type middlewareAuth struct {
	cfg        *config.Config
	jwtService service.JwtServiceInterface
	redis      *redis.Client
}

// CheckToken implements middlewareAuthInterface.
func (m *middlewareAuth) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[middlewareAuth-1] CheckToken: %s", "missing or invalid token")
				return c.JSON(
					http.StatusUnauthorized,
					response.APIResponseError(http.StatusUnauthorized, "missing or invalid token"),
				)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			_, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				log.Errorf("[middlewareAuth-2] CheckToken: %s", err.Error())
				return c.JSON(
					http.StatusUnauthorized,
					response.APIResponseError(http.StatusUnauthorized, "invalid token"),
				)
			}

			getSession, err := m.redis.Get(c.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[middlewareAuth-3] CheckToken: session not found")
				return c.JSON(
					http.StatusUnauthorized,
					response.APIResponseError(http.StatusUnauthorized, "session expired or not found"),
				)
			}

			jwtUserData := entity.JwtUserData{}
			err = json.Unmarshal([]byte(getSession), &jwtUserData)
			if err != nil {
				log.Errorf("[middlewareAuth-4] CheckToken: failed to parse user data")
				return c.JSON(
					http.StatusInternalServerError,
					response.APIResponseError(http.StatusInternalServerError, "Failed to parse user data"),
				)
			}

			c.Set("user", jwtUserData)
			return next(c)
		}
	}
}

// CheckRole implements middlewareAuthInterface.
func (m *middlewareAuth) CheckRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userData, ok := c.Get("user").(entity.JwtUserData)
			if !ok {
				log.Errorf("[middlewareAuth-5] CheckRole: user not found in context")
				return c.JSON(
					http.StatusUnauthorized,
					response.APIResponseError(http.StatusUnauthorized, "Unauthorized"),
				)
			}

			for _, role := range allowedRoles {
				if userData.Role == role {
					return next(c)
				}
			}

			requiredRoles := strings.Join(allowedRoles, ", ")
			log.Warnf("[middlewareAuth-6] CheckRole: access denied for role %s, required: %s", userData.Role, requiredRoles)
			return c.JSON(
				http.StatusForbidden,
				response.APIResponseError(http.StatusForbidden, "User role mismatch: required "+requiredRoles),
			)
		}
	}
}

func NewmiddlewareAuth(cfg *config.Config, jwtService service.JwtServiceInterface, redis *redis.Client) MiddlewareAuthInterface {
	return &middlewareAuth{
		cfg:        cfg,
		jwtService: jwtService,
		redis:      redis,
	}
}
