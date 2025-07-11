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
	GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error)
}

type CategoryRepository struct {
	db *gorm.DB
}

// GetAll implements CategoryRepositoryInterface.
func (c *CategoryRepository) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error) {
	modelCategories := []model.Category{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := c.db.Preload("Products").
		Where("name ILIKE ? OR slug ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	if err := sqlMain.Model(&modelCategories).Count(&countData).Error; err != nil {
		log.Errorf("[CategoryRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))
	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelCategories).Error; err != nil {
		log.Errorf("[CategoryRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelCategories) == 0 {
		err := errors.New("404")
		log.Infof("[CategoryRepository-3] GetAll: No category found")
		return nil, 0, 0, err
	}

	entities := []entity.CategoryEntity{}
	for _, val := range modelCategories {
		productEntities := []entity.ProductEntity{}
		for _, prd := range val.Products {
			productEntities = append(productEntities, entity.ProductEntity{
				ID:           prd.ID,
				CategorySlug: val.Slug,
				ParentID:     prd.ParentID,
				Name:         prd.Name,
				Image:        prd.Image,
			})
		}
		// status := "Published"
		// if val.Status == false {
		// 	status = "Unpublished"
		// }

		entities = append(entities, entity.CategoryEntity{
			ID:          val.ID,
			ParentID:    val.ParentID,
			Name:        val.Name,
			Icon:        val.Icon,
			Status:      val.Status,
			Slug:        val.Slug,
			Description: val.Description,
			Products:    productEntities,
		})
	}

	return entities, countData, int64(totalPage), nil

}

// GetBySlug implements CategoryRepositoryInterface.
func (c *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error) {
	modelCategory := model.Category{}

	if err := c.db.First(&modelCategory, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errors.New("404")
			log.Errorf("[CategoryRepository-1] GetBySlug: Category not found")
			return nil, err
		}
		log.Errorf("[CategoryRepository-2] GetBySlug: %v", err)
		return nil, err
	}

	return &entity.CategoryEntity{
		ID:          modelCategory.ID,
		ParentID:    modelCategory.ParentID,
		Name:        modelCategory.Name,
		Icon:        modelCategory.Icon,
		Status:      modelCategory.Status,
		Slug:        modelCategory.Slug,
		Description: modelCategory.Description,
	}, nil

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
		ID:          modelCategory.ID,
		ParentID:    modelCategory.ParentID,
		Name:        modelCategory.Name,
		Icon:        modelCategory.Icon,
		Status:      modelCategory.Status,
		Slug:        modelCategory.Slug,
		Description: modelCategory.Description,
	}, nil
}


func NewCategoryRepository(db *gorm.DB) CategoryRepositoryInterface {
	return &CategoryRepository{db: db}
}
