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
	UpdateBuyerLocation(ctx context.Context, buyerID int64, lat, lng string) error
	DeleteBuyer(ctx context.Context, buyerID int64) error

	UpsertProduct(ctx context.Context, product entity.ProductSnapshot) error
	GetProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductSnapshot, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
}

type localDataRepository struct {
	db *gorm.DB
}

// DeleteProduct implements LocalDataRepositoryInterface.
func (l *localDataRepository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}

	tx := l.db.WithContext(ctx).Begin()

	// 1. hapus child dulu
	if err := tx.
		Where("parent_id = ?", id).
		Delete(&model.ProductSnapshot{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. baru hapus parent
	if err := tx.
		Where("id = ?", id).
		Delete(&model.ProductSnapshot{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteBuyer implements LocalDataRepositoryInterface.
func (r *localDataRepository) DeleteBuyer(ctx context.Context, buyerID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", buyerID).
		Delete(&model.UserSnapshoot{}).Error
}

func (r *localDataRepository) UpdateBuyerLocation(ctx context.Context, buyerID int64, lat, lng string) error {
	result := r.db.WithContext(ctx).
		Model(&model.UserSnapshoot{}).
		Where("id = ?", buyerID).
		Updates(map[string]interface{}{
			"lat": lat,
			"lng": lng,
		})

	if result.Error != nil {
		log.Errorf("[UpdateBuyerLocation] Failed to update location for buyer %d: %v", buyerID, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Warnf("[UpdateBuyerLocation] Buyer %d not found in snapshot DB", buyerID)
		return gorm.ErrRecordNotFound
	}

	return nil
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
func (l *localDataRepository) GetProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductSnapshot, error) {
	var modelProduct model.ProductSnapshot
	if err := l.db.WithContext(ctx).First(&modelProduct, "id = ?", productID).Error; err != nil {
		return nil, err
	}
	return &entity.ProductSnapshot{
		ID:           modelProduct.ID,
		Name:         modelProduct.Name,
		RegulerPrice: modelProduct.RegulerPrice,
		SalePrice:    modelProduct.SalePrice,
		Image:        modelProduct.Image,
		Stock:        modelProduct.Stock,
		Weight:       modelProduct.Weight,
		Unit:         modelProduct.Unit,
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
		DoUpdates: clause.AssignmentColumns([]string{"id", "name", "email", "phone", "address", "lat", "lng", "updated_at"}),
	}).Create(&modelBuyer).Error

	if err != nil {
		log.Errorf("[UpsertBuyer-1] Failed to upsert buyer with ID %d: %v", buyer.ID, err)
		return err
	}
	return nil

}

// UpsertProduct implements LocalDataRepositoryInterface.
func (l *localDataRepository) UpsertProduct(ctx context.Context, product entity.ProductSnapshot) error {
	modelProduct := model.ProductSnapshot{
		ID:           product.ID,
		ParentID:     product.ParentID,
		Name:         product.Name,
		Image:        product.Image,
		Stock:        product.Stock,
		RegulerPrice: product.RegulerPrice,
		SalePrice:    product.SalePrice,
		Unit:         product.Unit,
		Weight:       product.Weight,
	}

	err := l.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "stock", "image",
				"reguler_price", "sale_price",
				"unit", "weight", "parent_id",
			}),
		}).
		Create(&modelProduct).Error

	if err != nil {
		log.Errorf("[UpsertProduct-1] %v", err)
		return err
	}

	if len(product.Child) > 0 {

		err := l.db.WithContext(ctx).
			Where("parent_id = ?", product.ID).
			Delete(&model.ProductSnapshot{}).Error
		if err != nil {
			log.Errorf("[UpsertProduct-2] delete child error: %v", err)
			return err
		}

		var childs []model.ProductSnapshot

		for _, val := range product.Child {
			childs = append(childs, model.ProductSnapshot{
				ID:           val.ID,
				ParentID:     &product.ID,
				Name:         product.Name,
				Image:        val.Image,
				RegulerPrice: val.RegulerPrice,
				SalePrice:    val.SalePrice,
				Unit:         val.Unit,
				Weight:       val.Weight,
				Stock:        val.Stock,
			})
		}

		err = l.db.WithContext(ctx).Create(&childs).Error
		if err != nil {
			log.Errorf("[UpsertProduct-3] insert child error: %v", err)
			return err
		}
	}

	return nil
}

func NewLocalDataRepository(db *gorm.DB) LocalDataRepositoryInterface {
	return &localDataRepository{db: db}
}
