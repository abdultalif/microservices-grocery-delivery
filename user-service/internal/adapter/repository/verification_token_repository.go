package repository

import (
	"context"
	"errors"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type VerificationTokenRepositoryInterface interface {
	GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEnity, error)
}

type VerificationTokenRepository struct {
	db *gorm.DB
}

// GetDataByToken implements VerificationTokenRepositoryInterface.
func (v *VerificationTokenRepository) GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEnity, error) {
	modelToken := model.VerificationToken{}

	if err := v.db.Where("token = ? AND token_type = ?", token, "email_verification").First(&modelToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("401")
			log.Errorf("[VerificationTokenRepository-1] GetDataByToken: %v", err)
			return nil, err
		}
		log.Errorf("[VerificationTokenRepository-2] GetDataByToken: %v", err)
		return nil, err
	}
	
	currentTime := time.Now()
	if currentTime.After(modelToken.ExpiresAt) {
		err := errors.New("401")
		log.Errorf("[VerificationTokenRepository-3] GetDataByToken: %v", err)
		return nil, err
	}  

	if err := v.db.Delete(&modelToken).Error; err != nil {
		log.Errorf("[VerificationTokenRepository-4] GetDataByToken: %v", err)
		return nil, err
	}

	return &entity.VerificationTokenEnity{
		ID: modelToken.ID,
		UserID: modelToken.UserID,
		Token: token,
		TokenType: modelToken.TokenType,
		ExpiresAt: modelToken.ExpiresAt,
	}, nil

}

func NewVerficationTokenRepository(db *gorm.DB) VerificationTokenRepositoryInterface {
	return &VerificationTokenRepository{db: db}
}
