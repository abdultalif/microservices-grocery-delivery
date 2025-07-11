package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CategoryEntity, error)
}

type CategoryRepository struct {
	db *gorm.DB
}

// GetByID implements CategoryRepositoryInterface.
func (c *CategoryRepository) GetByID(ctx context.Context, CategoryID uuid.UUID) (*entity.CategoryEntity, error) {
	
	modelCategory := model.Category{}

	if err := c.db.First(&modelCategory, "id = ?", CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errors.New("404")
			log.Errorf("[CategoryRepository-1] GetByID: Category not found")
			return nil, err
		}
		log.Errorf("[CategoryRepository-2] GetByID: %v", err)
		return nil, err
	}

	return &entity.CategoryEntity{
		ID:          CategoryID,
		ParentID:    modelCategory.ParentID,
		Name:        modelCategory.Name,
		Icon:        modelCategory.Icon,
		Status:      modelCategory.Status,
		Slug:        modelCategory.Slug,
		Description: modelCategory.Description,
	}, nil
}

// GetAll implements CategoryRepositoryInterface.
func (c *CategoryRepository) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error) {
	modelCatrogries := []model.Category{}

	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := c.db.Preload("Products").Where("name ILIKE ? OR status ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")

	if err := sqlMain.Model(&modelCatrogries).Count(&countData).Error; err != nil {
		log.Errorf("[CategoryRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))
	if err := sqlMain.Offset(int(offset)).Limit(int(query.Limit)).Order(order).Find(&modelCatrogries).Error; err != nil {
		log.Errorf("[CategoryRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelCatrogries) == 0 {
		err := errors.New("404")
		log.Errorf("[CategoryRepository-3] GetAll: Category not found")
		return nil, 0, 0, err
	}

	entities := []entity.CategoryEntity{}
	for _, category := range modelCatrogries {
		productEntities := []entity.ProductEntity{}
		for _, product := range category.Products {
			productEntities = append(productEntities, entity.ProductEntity{
				ID:           product.ID,
				ParentID:     product.ParentID,
				CategorySlug: product.CategorySlug,
				Name:         product.Name,
				Image:        product.Image,
			})
		}
		entities = append(entities, entity.CategoryEntity{
			ID:          category.ID,
			ParentID:    category.ParentID,
			Name:        category.Name,
			Icon:        category.Icon,
			Status:      category.Status,
			Slug:        category.Slug,
			Description: category.Description,
			Products:    productEntities,
		})
	}

	return entities, countData, int64(totalPage), nil

}

func NewCategoryRepository(db *gorm.DB) CategoryRepositoryInterface {
	return &CategoryRepository{db: db}
}
