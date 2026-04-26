package service

import (
	"context"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/message"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/utils"

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
	repo         repository.ProductRepositoryInterface
	repoCategory repository.CategoryRepositoryInterface
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

	getProductByID, err := p.repo.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-2] Create: %v", err)
		return err
	}

	reqProductSnapShot := entity.ProductEntity{
		ID:           getProductByID.ID,
		Name:         getProductByID.Name,
		Stock:        getProductByID.Stock,
		Image:        getProductByID.Image,
		RegulerPrice: getProductByID.RegulerPrice,
		SalePrice:    getProductByID.SalePrice,
		Unit:         getProductByID.Unit,
		Weight:       getProductByID.Weight,
		CreatedAt:    getProductByID.CreatedAt,
	}

	go func() {
		err := message.PublishProduct(
			utils.PRODUCT_CREATED_EVENT,
			utils.PRODUCT_CREATED_RK,
			reqProductSnapShot,
		)
		if err != nil {
			log.Errorf("[ProductService-Publish] %v", err)
		}
	}()

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

	go message.PublishProduct(utils.PRODUCT_UPDATED_EVENT, utils.PRODUCT_UPDATED_RK, *getProductByID)

	return nil
}

// Delete implements ProductServiceInterface.
func (p *productService) Delete(ctx context.Context, productID uuid.UUID) error {
	err := p.repo.Delete(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-1] Delete: %v", err)
		return err
	}

	go message.PublishProduct(utils.PRODUCT_DELETED_EVENT, utils.PRODUCT_DELETED_RK, entity.ProductEntity{ID: productID})

	return nil
}

// GetByID implements ProductServiceInterface.
func (p *productService) GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error) {
	result, err := p.repo.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-1] GetByID: %v", err)
		return nil, err
	}

	resultCat, err := p.repoCategory.GetBySlug(ctx, result.CategorySlug)
	if err != nil {
		log.Errorf("[ProductService-2] GetByID: %v", err)
		return nil, err
	}

	result.CategoryName = resultCat.Name

	return result, nil

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

func NewProductService(repo repository.ProductRepositoryInterface, repoCategory repository.CategoryRepositoryInterface) ProductServiceInterface {
	return &productService{
		repo:         repo,
		repoCategory: repoCategory,
	}
}
