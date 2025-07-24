package service

import (
	"context"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"

	"github.com/google/uuid"
)

type ProductServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error)
	Delete(ctx context.Context, productID uuid.UUID) error
	Create(ctx context.Context, req entity.ProductEntity) error
	Update(ctx context.Context, req entity.ProductEntity) error
}

type productService struct {
	repo repository.ProductRepositoryInterface
}

// Create implements ProductServiceInterface.
func (p *productService) Create(ctx context.Context, req entity.ProductEntity) error {
	return p.repo.Create(ctx, req)
}

// Update implements ProductServiceInterface.
func (p *productService) Update(ctx context.Context, req entity.ProductEntity) error {
	return p.repo.Update(ctx, req)
}

// Delete implements ProductServiceInterface.
func (p *productService) Delete(ctx context.Context, productID uuid.UUID) error {
	return p.repo.Delete(ctx, productID)
}

// GetByID implements ProductServiceInterface.
func (p *productService) GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error) {
	return p.repo.GetByID(ctx, productID)
}

// GetAll implements ProductServiceInterface.
func (p *productService) GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	return p.repo.GetAll(ctx, query)
}

func NewProductService(repo repository.ProductRepositoryInterface) ProductServiceInterface {
	return &productService{repo: repo}
}
