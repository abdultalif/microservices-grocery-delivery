package services

import (
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type JwtServiceInterface interface {
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtService struct {
	jwtSecret string
}

// ValidateToken implements JwtServiceInterface.
func (j *jwtService) ValidateToken(encodeToken string) (*jwt.Token, error) {
	return jwt.Parse(encodeToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(j.jwtSecret), nil
	})
}

func NewJwtService(cfg *config.Config) JwtServiceInterface {
	return &jwtService{
		jwtSecret: cfg.App.JwtSecret,
	}
}
