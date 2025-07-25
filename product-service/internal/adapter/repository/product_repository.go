package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error)
	Create(ctx context.Context, req entity.ProductEntity) error
	Update(ctx context.Context, req entity.ProductEntity) error
	CheckCategoryExists(ctx context.Context, slug string) error
	Delete(ctx context.Context, productID uuid.UUID) error
}

type ProductRepository struct {
	db *gorm.DB
}

// CheckCategoryExists implements ProductRepositoryInterface.
func (p *ProductRepository) CheckCategoryExists(ctx context.Context, slug string) error {
	var count int64
	if err := p.db.WithContext(ctx).Model(&model.Category{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		log.Errorf("[ProductRepository-CheckCategoryExists] %v", err)
		return err
	}

	if count == 0 {
		return errs.ErrCategoryNotFound
	}
	return nil
}

// Delete implements ProductRepositoryInterface.
func (p *ProductRepository) Delete(ctx context.Context, productID uuid.UUID) error {
	modelProduct := model.Product{}

	if err := p.db.Preload("Childs").First(&modelProduct, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("[ProductRepository-1] Delete: Product not found")
			return errs.ErrProductNotFound
		}
		log.Errorf("[ProductRepository-2] Delete: %v", err)
		return err
	}

	log.Errorf("childs: %v", len(modelProduct.Childs))
	if len(modelProduct.Childs) > 0 {
		log.Errorf("[ProductRepository-3] Delete: Product has children, cannot delete")
		return errs.ErrProductHasChildren
	}

	if err := p.db.Delete(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-4] Delete: %v", err)
		return err
	}

	return nil

}

// Update implements ProductRepositoryInterface.
func (p *ProductRepository) Update(ctx context.Context, req entity.ProductEntity) error {
	var existingProduct model.Product

	// Cari produk utama
	if err := p.db.WithContext(ctx).Where("id = ?", req.ID).First(&existingProduct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		log.Errorf("[ProductRepository-Update-1] %v", err)
		return err
	}

	// Update data produk utama
	existingProduct.Name = req.Name
	existingProduct.CategorySlug = req.CategorySlug
	existingProduct.Description = req.Description
	existingProduct.Image = req.Image
	existingProduct.RegulerPrice = req.RegulerPrice
	existingProduct.SalePrice = req.SalePrice
	existingProduct.Unit = req.Unit
	existingProduct.Weight = req.Weight
	existingProduct.Stock = req.Stock
	existingProduct.Variant = req.Variant
	existingProduct.Status = req.Status

	if err := p.db.WithContext(ctx).Save(&existingProduct).Error; err != nil {
		log.Errorf("[ProductRepository-Update-2] %v", err)
		return err
	}

	// Hapus semua child yang lama (jika ada)
	if err := p.db.WithContext(ctx).Where("parent_id = ?", req.ID).Delete(&model.Product{}).Error; err != nil {
		log.Errorf("[ProductRepository-Update-3] Delete Childs: %v", err)
		return err
	}

	// Simpan ulang child baru
	if len(req.Child) > 0 {
		children := []model.Product{}
		for _, child := range req.Child {
			children = append(children, model.Product{
				CategorySlug: req.CategorySlug,
				ParentID:     &req.ID,
				Name:         req.Name,
				Description:  req.Description,
				Image:        child.Image,
				RegulerPrice: child.RegulerPrice,
				SalePrice:    child.SalePrice,
				Unit:         req.Unit,
				Weight:       child.Weight,
				Stock:        child.Stock,
				Variant:      req.Variant,
				Status:       req.Status,
			})
		}
		if err := p.db.WithContext(ctx).Create(&children).Error; err != nil {
			log.Errorf("[ProductRepository-Update-4] Create Childs: %v", err)
			return err
		}
	}

	return nil

}

// Create implements ProductRepositoryInterface.
func (p *ProductRepository) Create(ctx context.Context, req entity.ProductEntity) error {
	modelProduct := model.Product{
		CategorySlug: req.CategorySlug,
		ParentID:     req.ParentID,
		Name:         req.Name,
		Image:        req.Image,
		Description:  req.Description,
		RegulerPrice: req.RegulerPrice,
		SalePrice:    req.SalePrice,
		Unit:         req.Unit,
		Weight:       req.Weight,
		Variant:      req.Variant,
		Status:       req.Status,
	}

	if err := p.db.Where("name = ?", req.Name).First(&model.Product{}).Error; err == nil {
		return errs.ErrProductAlreadyExists
	}

	if err := p.db.Create(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-1] Create: %v", err)
		return err
	}

	if len(req.Child) > 0 {
		modelProductChild := []model.Product{}
		for _, val := range req.Child {
			modelProductChild = append(modelProductChild, model.Product{
				CategorySlug: req.CategorySlug,
				ParentID:     &modelProduct.ID,
				Name:         req.Name,
				Image:        val.Image,
				Description:  req.Description,
				RegulerPrice: val.RegulerPrice,
				SalePrice:    val.SalePrice,
				Unit:         req.Unit,
				Weight:       val.Weight,
				Stock:        val.Stock,
				Variant:      req.Variant,
				Status:       req.Status,
			})
		}

		if err := p.db.Create(&modelProductChild).Error; err != nil {
			log.Errorf("[ProductRepository-2] Create: %v", err)
			return err
		}
	}

	return nil

}

// GetByID implements ProductRepositoryInterface.
func (p *ProductRepository) GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error) {
	modelProduct := model.Product{}
	if err := p.db.WithContext(ctx).Preload("Category").First(&modelProduct, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrProductNotFound
		}
		log.Errorf("[ProductRepository-1] GetByID: %v", err)
		return nil, err
	}

	modelParent := []model.Product{}
	err := p.db.WithContext(ctx).Preload("Category").Where("parent_id = ?", modelProduct.ID).Find(&modelParent).Error
	if err != nil {
		log.Errorf("[ProductRepository-2] GetByID: %v", err)
		return nil, err
	}

	childEntities := []entity.ProductEntity{}
	for _, val := range modelParent {
		childEntities = append(childEntities, entity.ProductEntity{
			ID:           val.ID,
			CategorySlug: val.CategorySlug,
			ParentID:     val.ParentID,
			Name:         val.Name,
			Image:        val.Image,
			Description:  val.Description,
			RegulerPrice: val.RegulerPrice,
			SalePrice:    val.SalePrice,
			Unit:         val.Unit,
			Weight:       val.Weight,
			Stock:        val.Stock,
			Variant:      val.Variant,
			Status:       val.Status,
			CategoryName: val.Category.Name,
			Child:        childEntities,
			CreatedAt:    val.CreatedAt,
		})
	}

	return &entity.ProductEntity{
		ID:           modelProduct.ID,
		CategorySlug: modelProduct.CategorySlug,
		ParentID:     modelProduct.ParentID,
		Name:         modelProduct.Name,
		Image:        modelProduct.Image,
		Description:  modelProduct.Description,
		RegulerPrice: modelProduct.RegulerPrice,
		SalePrice:    modelProduct.SalePrice,
		Unit:         modelProduct.Unit,
		Weight:       modelProduct.Weight,
		Stock:        modelProduct.Stock,
		Variant:      modelProduct.Variant,
		Status:       modelProduct.Status,
		CategoryName: modelProduct.Category.Name,
		Child:        childEntities,
		CreatedAt:    modelProduct.CreatedAt,
	}, nil
}

// GetAll implements ProductRepositoryInterface.
func (p *ProductRepository) GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	modelProducts := []model.Product{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit
	defaultStatus := "ACTIVE"
	if query.Status != "" {
		defaultStatus = query.Status
	}
	sqlMain := p.db.Preload("Category").
		Where("parent_id IS NULL AND status = ?", defaultStatus).
		Where("name ILIKE ? OR description ILIKE ? OR category_slug ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")
	if query.CategorySlug != "" {
		sqlMain = sqlMain.Where("category_slug = ?", query.CategorySlug)
	}

	if query.StartPrice > 0 {
		sqlMain = sqlMain.Where("sale_price >= ?", query.StartPrice)
	}

	if query.EndPrice > 0 {
		sqlMain = sqlMain.Where("sale_price <= ?", query.EndPrice)
	}

	if err := sqlMain.Model(&modelProducts).Count(&countData).Error; err != nil {
		log.Errorf("[ProductRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))
	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelProducts).Error; err != nil {
		log.Errorf("[ProductRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelProducts) == 0 {
		log.Errorf("[ProductRepository-3] GetAll: %v", "Data not found")
		return nil, 0, 0, errors.New("404")
	}

	respProducts := []entity.ProductEntity{}
	for _, val := range modelProducts {
		respProducts = append(respProducts, entity.ProductEntity{
			ID:           val.ID,
			CategorySlug: val.CategorySlug,
			ParentID:     val.ParentID,
			Name:         val.Name,
			Image:        val.Image,
			Description:  val.Description,
			RegulerPrice: val.RegulerPrice,
			SalePrice:    val.SalePrice,
			Unit:         val.Unit,
			Weight:       val.Weight,
			Stock:        val.Stock,
			Variant:      val.Variant,
			Status:       val.Status,
			CategoryName: val.Category.Name,
			CreatedAt:    val.CreatedAt,
		})
	}

	return respProducts, countData, int64(totalPage), nil
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &ProductRepository{db: db}
}
