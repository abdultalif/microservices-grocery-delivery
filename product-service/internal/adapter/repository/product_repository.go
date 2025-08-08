package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/domain/model"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)


type ProductRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
  	UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error
	GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error)
	Create(ctx context.Context, req entity.ProductEntity) (uuid.UUID, error)
	Update(ctx context.Context, req entity.ProductEntity) error
	CheckCategoryExists(ctx context.Context, slug string) error
	Delete(ctx context.Context, productID uuid.UUID) error

	GetAllHome(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	SearchProduct(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
}

type ProductRepository struct {
	db *gorm.DB
	esClient *elasticsearch.Client
}


// UploadPhoto implements ProductRepositoryInterface.
func (p *ProductRepository) UploadPhoto(ctx context.Context, productID uuid.UUID, photoURL string) error {
	if err := p.db.Model(&model.Product{}).Where("id = ?", productID).Update("image", photoURL).Error; err != nil {
		log.Errorf("[UserRepository-1] UpdatePhoto: %v", err)
		return err
	}
	return nil
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

	if err := p.db.WithContext(ctx).Select("Childs").Delete(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-2] Delete: %v", err)
		return err
	}

	res, err := p.esClient.Delete(
		"products",
		uuid.UUID(productID).String(),
		p.esClient.Delete.WithRefresh("true"),
	)
	if err != nil {
		log.Errorf("[ProductRepository-3] Delete: %v", err)
		return err
	}

	defer res.Body.Close()
	log.Infof("[ProductRepository-4] Delete Product Elasticsearch: %d", productID)
	
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
func (p *ProductRepository) Create(ctx context.Context, req entity.ProductEntity) (uuid.UUID, error) {
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
		Stock:        req.Stock,
		Variant:      req.Variant,
		Status:       req.Status,
	}

	if err := p.db.Where("name = ?", req.Name).First(&model.Product{}).Error; err == nil {
		err = errs.ErrProductAlreadyExists
		return uuid.Nil, err 
	}

	if err := p.db.Create(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-1] Create: %v", err)
		return uuid.Nil, err
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
			return uuid.Nil, err
		}
	}

	return modelProduct.ID, nil
}

// GetByID implements ProductRepositoryInterface.
func (p *ProductRepository) GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error) {
	modelProduct := model.Product{}
	if err := p.db.WithContext(ctx).
		Preload("Category").
		Preload("Childs").
		Preload("Childs.Category").
		First(&modelProduct, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrProductNotFound
		}
		log.Errorf("[ProductRepository-1] GetByID: %v", err)
		return nil, err
	}

	childEntities := modelProduct.Childs
	childChildEntities := make([]entity.ProductChildEntity, len(childEntities))
	for i, child := range childEntities {
		childChildEntities[i] = entity.ProductChildEntity{
			ID:           child.ID,
			Image:        child.Image,
			Weight:       child.Weight,
			Stock:        child.Stock,
			RegulerPrice: child.RegulerPrice,
			SalePrice:    child.SalePrice,
		}
	}

	// 👈 Revisi tambahan untuk jaga-jaga jika Category null
	categoryName := modelProduct.Category.Name
	if categoryName == "" {
		log.Warnf("[ProductRepository-2] GetByID: Category name kosong untuk slug %s", modelProduct.CategorySlug)
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
		CategoryName: categoryName,
		CreatedAt:    modelProduct.CreatedAt,
		Child:        childChildEntities,
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
		Preload("Childs").
		Model(&model.Product{}).
		Where("parent_id IS NULL").
		Where("status = ?", defaultStatus)
	
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		sqlMain = sqlMain.Where(`
			(name ILIKE @pattern OR description ILIKE @pattern OR category_slug ILIKE @pattern)
		`, map[string]interface{}{"pattern": searchPattern})
	}

	if query.CategorySlug != "" {
		sqlMain = sqlMain.Where("category_slug = ?", query.CategorySlug)
	}

	if query.StartPrice > 0 {
		sqlMain = sqlMain.Where("sale_price >= ?", query.StartPrice)
	}

	if query.EndPrice > 0 {
		sqlMain = sqlMain.Where("sale_price <= ?", query.EndPrice)
	}

	if err := sqlMain.Count(&countData).Error; err != nil {
		log.Errorf("[ProductRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))

	if err := sqlMain.Order(order).Limit(query.Limit).Offset(offset).Find(&modelProducts).Error; err != nil {
		log.Errorf("[ProductRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelProducts) == 0 {
		log.Warn("[ProductRepository-3] GetAll: Data not found")
		return nil, 0, 0, errs.ErrProductNotFound
	}

	respProducts := make([]entity.ProductEntity, len(modelProducts))
    for i, val := range modelProducts {
        childProductsEntity := make([]entity.ProductChildEntity, len(val.Childs))
        for j, child := range val.Childs {
            childProductsEntity[j] = entity.ProductChildEntity{
                ID:           child.ID,
                Image:        child.Image,
                Weight:       child.Weight,
                Stock:        child.Stock,
                RegulerPrice: child.RegulerPrice,
                SalePrice:    child.SalePrice,
            }
        }

        respProducts[i] = entity.ProductEntity{
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
            Child:        childProductsEntity,
        }
    }

	return respProducts, countData, int64(totalPage), nil
}


// SearchProduct implements ProductRepositoryInterface.
func (p *ProductRepository) SearchProduct(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	var (
		products      []entity.ProductEntity
	)

	from := (query.Page - 1) * query.Limit

	// Fix: Mapping field yang benar untuk sorting
	sortField := "id"
	switch query.OrderBy {
	case "name":
		sortField = "name.keyword" // Gunakan .keyword untuk sorting text field
	case "created_at":
		sortField = "created_at"
	case "reguler_price":
		sortField = "reguler_price"
	case "sale_price":
		sortField = "sale_price"
	default:
		sortField = "created_at" // Default ke created_at
	}

	sortOrder := "asc"
	if query.OrderType == "desc" {
		sortOrder = "desc"
	}

	// Build query dinamis
	var queryClause map[string]interface{}
	var mustClauses []map[string]interface{}
	var filterClauses []map[string]interface{}

	// Search clause
	if query.Search != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query.Search,
				"fields": []string{"name", "description", "category_name"},
				"fuzziness": "AUTO", // Tambahkan fuzziness untuk pencarian yang lebih fleksibel
			},
		})
	}

	// Category filter
	if query.CategorySlug != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"category_slug.keyword": query.CategorySlug,
			},
		})
	}

	// Price range filter
	if query.StartPrice > 0 && query.EndPrice > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"reguler_price": map[string]interface{}{
					"gte": query.StartPrice,
					"lte": query.EndPrice,
				},
			},
		})
	} else if query.StartPrice > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"reguler_price": map[string]interface{}{
					"gte": query.StartPrice,
				},
			},
		})
	} else if query.EndPrice > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"reguler_price": map[string]interface{}{
					"lte": query.EndPrice,
				},
			},
		})
	}

	// Build bool query
	if len(mustClauses) > 0 || len(filterClauses) > 0 {
		boolQuery := map[string]interface{}{}
		
		if len(mustClauses) > 0 {
			boolQuery["must"] = mustClauses
		}
		
		if len(filterClauses) > 0 {
			boolQuery["filter"] = filterClauses
		}

		queryClause = map[string]interface{}{
			"bool": boolQuery,
		}
	} else {
		// Jika tidak ada filter, gunakan match_all
		queryClause = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	// Build complete query
	esQuery := map[string]interface{}{
		"from": from,
		"size": query.Limit,
		"query": queryClause,
		"sort": []map[string]interface{}{
			{
				sortField: map[string]interface{}{
					"order": sortOrder,
				},
			},
		},
	}

	// Convert to JSON
	queryBytes, err := json.Marshal(esQuery)
	if err != nil {
		log.Errorf("[SearchProduct] Failed to marshal query: %v", err)
		return nil, 0, 0, err
	}

	log.Infof("[SearchProduct] Elasticsearch query: %s", string(queryBytes))

	res, err := p.esClient.Search(
		p.esClient.Search.WithContext(ctx),
		p.esClient.Search.WithIndex("products"),
		p.esClient.Search.WithBody(strings.NewReader(string(queryBytes))),
		p.esClient.Search.WithPretty(),
	)
	if err != nil {
		log.Errorf("[SearchProduct] Elasticsearch search error: %v", err)
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	// Read response body untuk debugging
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("[SearchProduct] Failed to read response body: %v", err)
		return nil, 0, 0, err
	}

	log.Infof("[SearchProduct] Elasticsearch response: %s", string(bodyBytes))

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Errorf("[SearchProduct] Failed to decode response: %v", err)
		return nil, 0, 0, err
	}

	// Check for errors in response
	if errorInfo, exists := result["error"]; exists {
		log.Errorf("[SearchProduct] Elasticsearch error in response: %v", errorInfo)
		return nil, 0, 0, fmt.Errorf("elasticsearch error: %v", errorInfo)
	}

	// Default values
	var totalData int64 = 0
	var totalPage int64 = 0

	// Ambil totalData dari hits.total.value
	if hitsMap, ok := result["hits"].(map[string]interface{}); ok {
		if totalMap, ok := hitsMap["total"].(map[string]interface{}); ok {
			if val, ok := totalMap["value"].(float64); ok {
				totalData = int64(val)
			}
		}
	}

	if query.Limit > 0 {
		totalPage = int64(math.Ceil(float64(totalData) / float64(query.Limit)))
	}

	// Ambil hits list
	if hitsMap, ok := result["hits"].(map[string]interface{}); ok {
		if rawHits, ok := hitsMap["hits"].([]interface{}); ok {
			for _, h := range rawHits {
				hit, ok := h.(map[string]interface{})
				if !ok {
					continue
				}
				source, ok := hit["_source"]
				if !ok {
					continue
				}
				data, err := json.Marshal(source)
				if err != nil {
					log.Warnf("[SearchProduct] Failed to marshal _source: %v", err)
					continue
				}
				var product entity.ProductEntity
				if err := json.Unmarshal(data, &product); err != nil {
					log.Warnf("[SearchProduct] Failed to unmarshal into ProductEntity: %v", err)
					continue
				}
				products = append(products, product)
			}
		}
	}

	return products, totalData, totalPage, nil
}

// GetAllHome implements ProductRepositoryInterface.
func (p *ProductRepository) GetAllHome(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	modelProducts := []model.Product{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit
	defaultStatus := "ACTIVE"
	if query.Status != "" {
		defaultStatus = query.Status
	}

	sqlMain := p.db.Preload("Category").
		Preload("Childs").
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
		return nil, 0, 0, errs.ErrProductNotFound
	}

	respProducts := []entity.ProductEntity{}
	for _, val := range modelProducts {
		childProducts := []entity.ProductChildEntity{}
		for _, child := range val.Childs {
			childProducts = append(childProducts, entity.ProductChildEntity{
				ID:           child.ID,
				Image:        child.Image,
				RegulerPrice: child.RegulerPrice,
				SalePrice:    child.SalePrice,
				Weight:       child.Weight,
				Stock:        child.Stock,
			})
		}

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
			Child:        childProducts,
			CreatedAt:    val.CreatedAt,
		})
	}

	return respProducts, countData, int64(totalPage), nil
}

func NewProductRepository(db *gorm.DB, es *elasticsearch.Client) ProductRepositoryInterface {
	return &ProductRepository{db: db, esClient: es}
}
