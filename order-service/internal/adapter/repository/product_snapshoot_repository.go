package repository

import (
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/model"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type ProductSnapshootRepositoryInterface interface {
	Create(req entity.ProductCustomerResponse) error
	GetByID(productID uuid.UUID) (*model.ProductSnapshoot, error)
}
type ProductSnapshootRepository struct {
	db *gorm.DB
}

func (p *ProductSnapshootRepository) GetByID(productID uuid.UUID) (*model.ProductSnapshoot, error) {
	var productSnapshoot model.ProductSnapshoot

	if err := p.db.Where("id = ?", productID).First(&productSnapshoot).Error; err != nil {
		log.Errorf("[ProductSnapshootRepository-1] GetByID: %v", err)
	}

	return &productSnapshoot, nil

}

func (p *ProductSnapshootRepository) Create(req entity.ProductCustomerResponse) error {
	modelProductSnapshoot := model.ProductSnapshoot{
		ID:           req.ID,
		Name:         req.Name,
		Stock:        req.Stock,
		Image:        req.Image,
		RegulerPrice: req.RegulerPrice,
		SalePrice:    req.SalePrice,
		Unit:         req.Unit,
		Weight:       req.Weight,
		CreatedAt:    req.CreatedAt,
	}

	if err := p.db.FirstOrCreate(&modelProductSnapshoot, &model.ProductSnapshoot{ID: req.ID}).Error; err != nil {
		log.Errorf("[ProductSnapshootRepository-1] create: %v", err)
		return err
	}

	if len(req.Child) > 0 {
		for _, child := range req.Child {
			childModel := model.ProductSnapshoot{
				ID:           child.ID,
				Name:         child.Name,
				Image:        child.Image,
				Stock:        child.Stock,
				RegulerPrice: child.RegulerPrice,
				SalePrice:    child.SalePrice,
				Unit:         child.Unit,
				Weight:       child.Weight,
				CreatedAt:    child.CreatedAt,
			}

			if err := p.db.FirstOrCreate(&childModel, &model.ProductSnapshoot{ID: child.ID}).Error; err != nil {
				log.Errorf("[ProductSnapshootRepository-2] create: %v", err)
				return err
			}
		}
	}

	return nil

}

func NewProductSnapshootRepository(db *gorm.DB) ProductSnapshootRepositoryInterface {
	return &ProductSnapshootRepository{
		db: db,
	}
}
