package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CategoryEntity, error)
	GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error)
	Create(ctx context.Context, req entity.CategoryEntity) error
	Delete(ctx context.Context, categoryID uuid.UUID) error
	Update(ctx context.Context, req entity.CategoryEntity) error

	GetAllPublished(ctx context.Context) ([]entity.CategoryEntity, error)
}

type CategoryRepository struct {
	db *gorm.DB
}

// Update implements CategoryRepositoryInterface.
func (c *CategoryRepository) Update(ctx context.Context, req entity.CategoryEntity) error {

	modelCategory := model.Category{
		ID:          req.ID,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Icon:        req.Icon,
		Status:      req.Status,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if err := c.db.Model(&model.Category{}).
		Where("id = ?", req.ID).
		Updates(&modelCategory).Error; err != nil {
		log.Errorf("[CategoryRepository-1] Update: %v", err)
		return err
	}

	return nil
}

// Delete implements CategoryRepositoryInterface.
func (c *CategoryRepository) Delete(ctx context.Context, categoryID uuid.UUID) error {
	modelCategory := model.Category{}

	if err := c.db.Preload("Products").Preload("Children").First(&modelCategory, "id = ?", categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("[CategoryRepository-1] Delete: Category not found")
			return errs.ErrCategoryNotFound
		}
		log.Errorf("[CategoryRepository-2] Delete: %v", err)
		return err
	}

	if len(modelCategory.Products) > 0 {
		log.Errorf("[CategoryRepository-3] Delete: Category has products, cannot delete")
		return errs.ErrCategoryHasProducts
	}

	if len(modelCategory.Children) > 0 {
		log.Errorf("[CategoryRepository-4] Delete: Category has children, cannot delete")
		return errs.ErrCategoryHasChildren
	}

	if err := c.db.Delete(&modelCategory).Error; err != nil {
		log.Errorf("[CategoryRepository-5] Delete: %v", err)
		return err
	}

	return nil
}

// Create implements CategoryRepositoryInterface.
func (c *CategoryRepository) Create(ctx context.Context, req entity.CategoryEntity) error {

	modelCategory := model.Category{
		ID:          uuid.New(),
		ParentID:    req.ParentID,
		Name:        req.Name,
		Icon:        req.Icon,
		Status:      req.Status,
		Slug:        req.Slug,
		Description: req.Description,
	}
	if err := c.db.Create(&modelCategory).Error; err != nil {
		log.Errorf("[CategoryRepository-1] Create: %v", err)
		return err
	}

	return nil
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
		log.Infof("[CategoryRepository-3] GetAll: No category found")
		return nil, 0, 0, errs.ErrCategoryNotFound
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
			log.Errorf("[CategoryRepository-1] GetBySlug: Category not found")
			return nil, errs.ErrCategoryNotFound
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
			log.Errorf("[CategoryRepository-1] GetByID: Category not found")
			return nil, errs.ErrCategoryNotFound
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

// GetAllPublished implements CategoryRepositoryInterface.
func (c *CategoryRepository) GetAllPublished(ctx context.Context) ([]entity.CategoryEntity, error) {
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
	return &CategoryRepository{db: db}
}
