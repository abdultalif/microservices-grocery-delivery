package service

import (
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type JwtServiceInterface interface {
	GenerateToken(userID int64) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtService struct {
	jwtSecret string
	jwtIssuer string
}

// GenerateToken implements JwtServiceInterface.
func (j *jwtService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"issuer":  j.jwtIssuer,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.jwtSecret))
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
		jwtIssuer: cfg.App.JwtIssuer,
	}
}
