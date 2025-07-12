package service

import (
	"context"
	"errors"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"
	"product-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type CategoryServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetByID(ctx context.Context, categoryId uuid.UUID) (*entity.CategoryEntity, error)
	GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error)
	Create(ctx context.Context, req entity.CategoryEntity) error
	Delete(ctx context.Context, categoryID uuid.UUID) error
}

type CategoryService struct {
	repository repository.CategoryRepositoryInterface
}

// Create implements CategoryServiceInterface.
func (c *CategoryService) Create(ctx context.Context, req entity.CategoryEntity) error {
	slug := conv.GenerateSlug(req.Name)
	result, err := c.repository.GetBySlug(ctx, slug)
	if err != nil {

		log.Errorf("[CategoryService-1] Create: %v", err)
		return err
	}

	if result != nil {
		err = errors.New("409")
		log.Errorf("[CategoryService-2] Create: Category with slug %s already exists", slug)
		return err
	}

	req.Slug = slug
	err = c.repository.Create(ctx, req)
	if err != nil {
		log.Errorf("[CategoryService-3] Create: %v", err)
		return err
	}

	return nil
}

// Delete implements CategoryServiceInterface.
func (c *CategoryService) Delete(ctx context.Context, categoryID uuid.UUID) error {
	return c.repository.Delete(ctx, categoryID)
}

// GetBySlug implements CategoryServiceInterface.
func (c *CategoryService) GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error) {
	return c.repository.GetBySlug(ctx, slug)
}

// GetAll implements CategoryServiceInterface.
func (c *CategoryService) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error) {
	return c.repository.GetAll(ctx, query)
}

// GetByID implements CategoryServiceInterface.
func (c *CategoryService) GetByID(ctx context.Context, categoryId uuid.UUID) (*entity.CategoryEntity, error) {
	return c.repository.GetByID(ctx, categoryId)
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &CategoryService{
		repository: repo,
	}
}
