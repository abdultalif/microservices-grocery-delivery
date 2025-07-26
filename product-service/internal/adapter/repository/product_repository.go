package repository

import (
	"context"
	"product-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error
}

type ProductRepository struct {
	db *gorm.DB
}

// UploadPhoto implements ProductRepositoryInterface.
func (p *ProductRepository) UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error {
	if err := p.db.Model(&model.Product{}).Where("id = ?", productID).Update("image", photoURL).Error; err != nil {
		log.Errorf("[UserRepository-1] UpdatePhoto: %v", err)
		return err
	}
	return nil
}



func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}
