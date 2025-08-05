package service

import (
	"context"
	"product-service/internal/adapter/message"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
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
	publisherRabbitMQ message.PublishRabbitMQInterface
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

	// 👈 Revisi: Ambil data lengkap agar bisa dipublish
	getProductByID, err := p.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-2] Create: %v", err)
		return err
	}

	productToPublish := *getProductByID

	// 👈 Revisi: Publish ke RabbitMQ
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

func NewProductService(repo repository.ProductRepositoryInterface, publisherRabbitMQ message.PublishRabbitMQInterface) ProductServiceInterface {
	return &productService{
		repo: repo,
		publisherRabbitMQ: publisherRabbitMQ,
	}
}
