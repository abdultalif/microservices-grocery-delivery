package service

import (
	"context"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"

	"github.com/google/uuid"
)

type CategoryServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetByID(ctx context.Context, categoryId uuid.UUID) (*entity.CategoryEntity, error)
}

type CategoryService struct {
	repository repository.CategoryRepositoryInterface
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
