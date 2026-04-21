package repository

import (
	"context"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/model"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LocalDataRepositoryInterface interface {
	UpsertBuyer(ctx context.Context, buyer entity.CustomerResponseEntity) error
	GetBuyer(ctx context.Context, buyerID int64) (*entity.CustomerResponseEntity, error)

	UpsertProduct(ctx context.Context, product entity.ProductResponseEntity) error
	GetProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductResponseEntity, error)
}

type localDataRepository struct {
	db *gorm.DB
}

// GetBuyer implements LocalDataRepositoryInterface.
func (l *localDataRepository) GetBuyer(ctx context.Context, buyerID int64) (*entity.CustomerResponseEntity, error) {
	var modelBuyer model.UserSnapshoot
	if err := l.db.WithContext(ctx).First(&modelBuyer, "id = ?", buyerID).Error; err != nil {
		return nil, err
	}
	return &entity.CustomerResponseEntity{
		ID:      modelBuyer.ID,
		Name:    modelBuyer.Name,
		Email:   modelBuyer.Email,
		Phone:   modelBuyer.Phone,
		Address: modelBuyer.Address,
		Lat:     modelBuyer.Lat,
		Lng:     modelBuyer.Lng,
	}, nil
}

// GetProduct implements LocalDataRepositoryInterface.
func (l *localDataRepository) GetProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductResponseEntity, error) {
	var modelProduct model.ProductSnapshot
	if err := l.db.WithContext(ctx).First(&modelProduct, "id = ?", productID).Error; err != nil {
		return nil, err
	}
	return &entity.ProductResponseEntity{
		ID:           modelProduct.ID,
		ProductName:  modelProduct.Name,
		RegulerPrice: float64(modelProduct.RegulerPrice),
		ProductImage: modelProduct.Image,
		Stock:        modelProduct.Stock,
	}, nil
}

// UpsertBuyer implements LocalDataRepositoryInterface.
func (l *localDataRepository) UpsertBuyer(ctx context.Context, buyer entity.CustomerResponseEntity) error {
	modelBuyer := model.UserSnapshoot{
		ID:      buyer.ID,
		Name:    buyer.Name,
		Email:   buyer.Email,
		Phone:   buyer.Phone,
		Address: buyer.Address,
		Lat:     buyer.Lat,
		Lng:     buyer.Lng,
	}

	err := l.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "email", "phone", "address", "lat", "lng", "updated_at"}),
	}).Create(&modelBuyer).Error

	if err != nil {
		log.Errorf("[UpsertBuyer-1] Failed to upsert buyer with ID %d: %v", buyer.ID, err)
		return err
	}
	return nil

}

// UpsertProduct implements LocalDataRepositoryInterface.
func (l *localDataRepository) UpsertProduct(ctx context.Context, product entity.ProductResponseEntity) error {
	modelProduct := model.ProductSnapshot{
		ID:           product.ID,
		Name:         product.ProductName,
		RegulerPrice: product.RegulerPrice,
		SalePrice:    product.SalePrice,
		Image:        product.ProductImage,
		Stock:        product.Stock,
		Weight:       product.Weight,
		Unit:         product.Unit,
	}
	err := l.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "reguler_price", "sale_price", "image", "stock", "weight", "unit"}),
		}).
		Create(&modelProduct).Error

	if err != nil {
		log.Errorf("[UpsertProduct-1] Failed to upsert product with ID %s: %v", product.ID, err)
		return err
	}

	return nil

}

func NewLocalDataRepository(db *gorm.DB) LocalDataRepositoryInterface {
	return &localDataRepository{db: db}
}
