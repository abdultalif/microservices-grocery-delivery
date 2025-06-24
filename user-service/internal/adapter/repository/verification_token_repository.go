package repository

import (
	"context"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type VerificationTokenRepositoryInterface interface {
	CreateVerification(ctx context.Context, req entity.VerificationTokenEnity) error
}

type VerificationTokenRepository struct {
	db *gorm.DB
}

// CreateVerification implements VerificationTokenRepositoryInterface.
func (v *VerificationTokenRepository) CreateVerification(ctx context.Context, req entity.VerificationTokenEnity) error {
	modelVerification := model.VerificationToken{
		UserID: req.UserID,
		Token: req.Token,
		TokenType: req.TokenType,
	}

	if err := v.db.Create(&modelVerification).Error; err != nil {
		log.Errorf("[VerificationTokenRepository-1] CreateVerification: %v", err)
	}

	return nil

}

func NewVerficationTokenRepository(db *gorm.DB) VerificationTokenRepositoryInterface {
	return &VerificationTokenRepository{db: db}
}
