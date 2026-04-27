package repository

import (
	"context"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/model"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LocalDataRepositoryInterface interface {
	UpsertBuyer(ctx context.Context, buyer entity.CustomerResponseEntity) error
	GetBuyer(ctx context.Context, buyerID int64) (*entity.CustomerResponseEntity, error)
	DeleteBuyer(ctx context.Context, buyerID int64) error
}

type localDataRepository struct {
	db *gorm.DB
}

// DeleteBuyer implements LocalDataRepositoryInterface.
func (r *localDataRepository) DeleteBuyer(ctx context.Context, buyerID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", buyerID).
		Delete(&model.UserSnapshot{}).Error
}

// GetBuyer implements LocalDataRepositoryInterface.
func (l *localDataRepository) GetBuyer(ctx context.Context, buyerID int64) (*entity.CustomerResponseEntity, error) {
	var modelBuyer model.UserSnapshot
	if err := l.db.WithContext(ctx).First(&modelBuyer, "id = ?", buyerID).Error; err != nil {
		log.Errorf("[GetBuyer-1] Failed to get buyer with ID %d: %v", buyerID, err)
		return nil, err
	}
	return &entity.CustomerResponseEntity{
		ID:      modelBuyer.ID,
		Name:    modelBuyer.Name,
		Email:   modelBuyer.Email,
		Address: modelBuyer.Address,
	}, nil
}

// UpsertBuyer implements LocalDataRepositoryInterface.
func (l *localDataRepository) UpsertBuyer(ctx context.Context, buyer entity.CustomerResponseEntity) error {
	modelBuyer := model.UserSnapshot{
		ID:      buyer.ID,
		Name:    buyer.Name,
		Email:   buyer.Email,
		Address: buyer.Address,
	}

	err := l.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "email", "address", "updated_at"}),
	}).Create(&modelBuyer).Error

	if err != nil {
		log.Errorf("[UpsertBuyer-1] Failed to upsert buyer with ID %d: %v", buyer.ID, err)
		return err
	}
	return nil

}

func NewLocalDataRepository(db *gorm.DB) LocalDataRepositoryInterface {
	return &localDataRepository{db: db}
}
