package middleware

import (
	"encoding/json"
	"net/http"
	"order-service/config"
	"order-service/internal/adapter/handler/response"
	"order-service/internal/core/domain/entity"
	"order-service/internal/core/service"
	"strings"

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
}

// CheckToken implements MiddlewareAdapterInterface.
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			redisConn := config.NewConfig().NewRedisClient()

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: %s", "missing or invalid token")
				return c.JSON(
					http.StatusUnauthorized, 
					response.ResponseAPI(false, http.StatusUnauthorized, "missing or invalid token", nil),
				)
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			
			_, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", err.Error())
				return c.JSON(
					http.StatusUnauthorized, 
					response.ResponseAPI(false, http.StatusUnauthorized, "invalid token", nil),
				)
			}

			getSession, err := redisConn.Get(c.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-3] CheckToken: session not found")
				return c.JSON(
					http.StatusUnauthorized, 
					response.ResponseAPI(false, http.StatusUnauthorized, "session expired or not found", nil),
				)
			}

			jwtUserData := entity.JwtUserData{}
			err = json.Unmarshal([]byte(getSession), &jwtUserData)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-4] CheckToken: failed to parse user data")
				return c.JSON(
					http.StatusInternalServerError, 
					response.ResponseAPI(false, http.StatusInternalServerError, "Failed to parse user data", nil),
				)
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
			userData, ok := c.Get("user").(entity.JwtUserData)
			if !ok {
				log.Errorf("[MiddlewareAdapter-5] CheckRole: user not found in context")
				return c.JSON(
					http.StatusUnauthorized, 
					response.ResponseAPI(false, http.StatusUnauthorized, "Unauthorized", nil),
				)
			}
			
			for _, role := range allowedRoles {
				if userData.Role == role {
					return next(c)
				}
			}
			
			requiredRoles := strings.Join(allowedRoles, ", ")
			log.Warnf("[MiddlewareAdapter-6] CheckRole: access denied for role %s, required: %s", userData.Role, requiredRoles)
			return c.JSON(
				http.StatusForbidden, 
				response.ResponseAPI(false, http.StatusForbidden, "User role mismatch: required " + requiredRoles, nil),
			)
		}
	}
}

func NewMiddlewareAdapter(cfg *config.Config, jwtService service.JwtServiceInterface) MiddlewareAdapterInterface {
	return &middlewareAdapter{
		cfg:        cfg,
		jwtService: jwtService,
	}
}