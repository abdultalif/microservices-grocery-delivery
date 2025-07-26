package service

import (
	"context"
	"product-service/config"
	"product-service/internal/adapter/repository"

	"github.com/google/uuid"
)

type ProductServiceInterface interface {
	UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error
}
type ProductService struct {
	repo       repository.ProductRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
}

// UploadPhoto implements ProductServiceInterface.
func (p *ProductService) UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error {
	return p.repo.UploadPhoto(ctx, productID, photoURL)
}

func NewProductService(productRepository repository.ProductRepositoryInterface) ProductServiceInterface {
	return &ProductService{
		repo:       productRepository,
	}
}
