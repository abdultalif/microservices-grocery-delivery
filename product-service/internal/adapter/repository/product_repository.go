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
	GetAllHome(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	SearchProduct(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error)
}

type ProductRepository struct {
	db *gorm.DB
	esClient *elasticsearch.Client
}

// GetByID implements ProductRepositoryInterface.
func (p *ProductRepository) GetByID(ctx context.Context, productID uuid.UUID) (*entity.ProductEntity, error) {
	modelProduct := model.Product{}
	if err := p.db.WithContext(ctx).
		Preload("Category").
		Preload("Childs.Category").
		First(&modelProduct, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrProductNotFound
		}
		log.Errorf("[ProductRepository-1] GetByID: %v", err)
		return nil, err
	}

	childEntities := []entity.ProductEntity{}
	for _, child := range modelProduct.Childs {
		childEntities = append(childEntities, entity.ProductEntity{
			ID:           child.ID,
			CategorySlug: child.CategorySlug,
			ParentID:     child.ParentID,
			Name:         child.Name,
			Image:        child.Image,
			Description:  child.Description,
			RegulerPrice: child.RegulerPrice,
			SalePrice:    child.SalePrice,
			Unit:         child.Unit,
			Weight:       child.Weight,
			Stock:        child.Stock,
			Variant:      child.Variant,
			Status:       child.Status,
			CategoryName: child.Category.Name,
			CreatedAt:    child.CreatedAt,
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
		CreatedAt:    modelProduct.CreatedAt,
		Child:        childEntities,
	}, nil
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
		childProducts := []entity.ProductEntity{}
		for _, child := range val.Childs {
			childProducts = append(childProducts, entity.ProductEntity{
				ID:           child.ID,
				CategorySlug: child.CategorySlug,
				ParentID:     child.ParentID,
				Name:         child.Name,
				Image:        child.Image,
				Description:  child.Description,
				RegulerPrice: child.RegulerPrice,
				SalePrice:    child.SalePrice,
				Unit:         child.Unit,
				Weight:       child.Weight,
				Stock:        child.Stock,
				Variant:      child.Variant,
				Status:       child.Status,
				CategoryName: child.Category.Name,
				CreatedAt:    child.CreatedAt,
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
