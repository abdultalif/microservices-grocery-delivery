package service

import (
	"context"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"
)

type CategoryServiceInterface interface {
	GetAllPublished(ctx context.Context) ([]entity.CategoryEntity, error)
}

type categoryService struct {
	repo repository.CategoryRepositoryInterface
}

// GetAllPublished implements CategoryServiceInterface.
func (c *categoryService) GetAllPublished(ctx context.Context) ([]entity.CategoryEntity, error) {
	return c.repo.GetAllPublished(ctx)
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &categoryService{repo: repo}
}
