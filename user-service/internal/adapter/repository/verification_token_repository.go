package repository

import (
	"context"
	"errors"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type VerificationTokenRepositoryInterface interface {
	CreateVerification(ctx context.Context, req entity.VerificationTokenEnity) error
	GetDataByToken(ctx context.Context, token string, tokenType string) (*entity.VerificationTokenEnity, error)
	GetDataWithoutDelete(ctx context.Context, token string) (*entity.VerificationTokenEnity, error)
}

type VerificationTokenRepository struct {
	db *gorm.DB
}

// GetDataWithoutDelete implements VerificationTokenRepositoryInterface.
func (v *VerificationTokenRepository) GetDataWithoutDelete(ctx context.Context, token string) (*entity.VerificationTokenEnity, error) {
	modelToken := model.VerificationToken{}

	if err := v.db.Where("token = ? AND token_type = ?", token, "forgot_password").First(&modelToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrTokenExpired
			log.Errorf("[VerificationTokenRepository-1] GetDataWithoutDelete: %v", err)
			return nil, err
		}
		log.Errorf("[VerificationTokenRepositor-2] GetDataWithoutDelete: %v", err)
		return nil, err
	}

	if time.Now().After(modelToken.ExpiresAt) {
		err := errs.ErrInvalidToken
		log.Errorf("[VerificationTokenRepository-3] GetDataWithoutDelete: expired")
		return nil, err
	}

	return &entity.VerificationTokenEnity{
		ID:        modelToken.ID,
		UserID:    modelToken.UserID,
		Token:     modelToken.Token,
		TokenType: modelToken.TokenType,
		ExpiresAt: modelToken.ExpiresAt,
	}, nil
}

// GetDataByToken implements VerificationTokenRepositoryInterface.
func (v *VerificationTokenRepository) GetDataByToken(ctx context.Context, token string, tokenType string) (*entity.VerificationTokenEnity, error) {

	modelToken := model.VerificationToken{}

	if err := v.db.Where("token = ? AND token_type = ?", token, tokenType).First(&modelToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrTokenExpired
			log.Errorf("[VerificationTokenRepository-1] GetDataByToken: %v", err)
			return nil, err
		}
		log.Errorf("[VerificationTokenRepository-2] GetDataByToken: %v", err)
		return nil, err
	}

	currentTime := time.Now()
	if currentTime.After(modelToken.ExpiresAt) {
		err := errs.ErrInvalidToken
		log.Errorf("[VerificationTokenRepository-3] GetDataByToken: %v", err)
		return nil, err
	}

	if err := v.db.Delete(&modelToken).Error; err != nil {
		log.Errorf("[VerificationTokenRepository-4] GetDataByToken: %v", err)
		return nil, err
	}

	return &entity.VerificationTokenEnity{
		ID:        modelToken.ID,
		UserID:    modelToken.UserID,
		Token:     token,
		TokenType: modelToken.TokenType,
		ExpiresAt: modelToken.ExpiresAt,
	}, nil
}

// CreateVerification implements VerificationTokenRepositoryInterface.
func (v *VerificationTokenRepository) CreateVerification(ctx context.Context, req entity.VerificationTokenEnity) error {
	modelVerification := model.VerificationToken{
		UserID:    req.UserID,
		Token:     req.Token,
		TokenType: req.TokenType,
		ExpiresAt: req.ExpiresAt,
	}

	if err := v.db.Create(&modelVerification).Error; err != nil {
		log.Errorf("[VerificationTokenRepository-1] CreateVerification: %v", err)
	}

	return nil
}

func NewVerficationTokenRepository(db *gorm.DB) VerificationTokenRepositoryInterface {
	return &VerificationTokenRepository{db: db}
}
