package jwt

import (
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/health" {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "Authorization header required")
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return echo.NewHTTPError(401, "Invalid authorization header")
			}

			secretKey := os.Getenv("JWT_SECRET")
			if secretKey == "" {
				secretKey = "your-secret-key"
			}

			token, err := jwt.ParseWithClaims(tokenParts[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})

			if err != nil {
				return echo.NewHTTPError(401, "Invalid token")
			}

			claims, ok := token.Claims.(*Claims)
			if !ok {
				return echo.NewHTTPError(401, "Invalid token claims")
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}
