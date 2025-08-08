package repository

import (
	"context"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"

	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {

	GetAllPublished(ctx context.Context) ([]entity.CategoryEntity, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// GetAllPublished implements CategoryRepositoryInterface.
func (c *categoryRepository) GetAllPublished(ctx context.Context) ([]entity.CategoryEntity, error) {
	modelCategories := []model.Category{}

	if err := c.db.Select("id, parent_id, name, icon, slug, status").Where("status = ?", "Published").Find(&modelCategories).Error; err != nil {
		log.Errorf("[CategoryRepository-1] GetAllPublished: %v", err)
		return nil, err
	}

	if len(modelCategories) == 0 {
		err := errs.ErrCategoryNotFound
		log.Infof("[CategoryRepository-2] GetAllPublished: No category found")
		return nil, err
	}

	entities := []entity.CategoryEntity{}
	for _, val := range modelCategories {
		entities = append(entities, entity.CategoryEntity{
			ID:       val.ID,
			ParentID: val.ParentID,
			Name:     val.Name,
			Icon:     val.Icon,
			Status:   val.Status,
			Slug:     val.Slug,
		})
	}

	return entities, nil
}


func NewCategoryRepository(db *gorm.DB) CategoryRepositoryInterface {
	return &categoryRepository{db: db}
}
