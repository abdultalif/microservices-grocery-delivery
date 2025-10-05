package service

import (
	"context"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/message"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type ProductServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error
	GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error)
	Delete(ctx context.Context, productID uuid.UUID) error
	Create(ctx context.Context, req entity.ProductEntity) error
	Update(ctx context.Context, req entity.ProductEntity) error

	GetAllHome(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	SearchProducts(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
}

type productService struct {
	repo              repository.ProductRepositoryInterface
	publisherRabbitMQ message.PublishRabbitMQInterface
}

// UploadPhoto implements ProductServiceInterface.
func (p *productService) UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error {
	return p.repo.UploadPhoto(ctx, productID, photoURL)
}

// Create implements ProductServiceInterface.
func (p *productService) Create(ctx context.Context, req entity.ProductEntity) error {
	if err := p.repo.CheckCategoryExists(ctx, req.CategorySlug); err != nil {
		return err
	}

	productID, err := p.repo.Create(ctx, req)
	if err != nil {
		log.Errorf("[ProductService-1] Create: %v", err)
		return err
	}

	getProductByID, err := p.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-2] Create: %v", err)
		return err
	}

	productToPublish := *getProductByID

	if err := p.publisherRabbitMQ.PublishProductToQueue(productToPublish); err != nil {
		log.Errorf("[ProductService-3] Create: %v", err)
		return err
	}

	return nil
}

// Update implements ProductServiceInterface.
func (p *productService) Update(ctx context.Context, req entity.ProductEntity) error {
	if err := p.repo.CheckCategoryExists(ctx, req.CategorySlug); err != nil {
		return err
	}

	err := p.repo.Update(ctx, req)
	if err != nil {
		log.Errorf("[ProductService-1] Update: %v", err)
		return err
	}

	getProductByID, err := p.GetByID(ctx, req.ID)
	if err != nil {
		log.Errorf("[ProductService-2] Update: %v", err)
	}

	if err := p.publisherRabbitMQ.PublishProductToQueue(*getProductByID); err != nil {
		log.Errorf("[ProductService-3] Update: %v", err)
	}

	return nil
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

// SearchProducts implements ProductServiceInterface.
func (p *productService) SearchProducts(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	return p.repo.SearchProduct(ctx, query)
}

// GetAllHome implements ProductServiceInterface.
func (p *productService) GetAllHome(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	return p.repo.GetAllHome(ctx, query)
}

func NewProductService(repo repository.ProductRepositoryInterface, publisherRabbitMQ message.PublishRabbitMQInterface) ProductServiceInterface {
	return &productService{
		repo:              repo,
		publisherRabbitMQ: publisherRabbitMQ,
	}
}
