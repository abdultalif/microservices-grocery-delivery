package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/domain/model"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	SearchProduct(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
}

type ProductRepository struct {
	db *gorm.DB
	esClient *elasticsearch.Client
}

// SearchProduct implements ProductRepositoryInterface.
func (p *ProductRepository) SearchProduct(ctx context.Context, query entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	var (
		mainQueries   []string
		filterQueries []string
		products      []entity.ProductEntity
	)

	from := (query.Page - 1) * query.Limit

	sortField := "id"
	if query.OrderBy != "" {
		sortField = query.OrderBy
	}

	sortOrder := "asc"
	if query.OrderType == "desc" {
		sortOrder = "desc"
	}

	sortQuery := fmt.Sprintf(`{ "%s": "%s" }`, sortField, sortOrder)

	if query.CategorySlug != "" {
		filterQueries = append(filterQueries, fmt.Sprintf(`{ "term": { "category_slug.keyword": "%s" } }`, query.CategorySlug))
	}

	if query.StartPrice > 0 && query.EndPrice > 0 {
		filterQueries = append(filterQueries, fmt.Sprintf(`{ "range": { "reguler_price": { "gte": %d, "lte": %d } } }`, query.StartPrice, query.EndPrice))
	}

	if query.Search != "" {
		mainQueries = append(mainQueries, fmt.Sprintf(`{ "multi_match": { "query": "%s", "fields": ["name", "description", "category_name"] } }`, query.Search))
	}

	mainQuery := fmt.Sprintf(`{
		"from": %d,
		"size": %d,
		"query": {
			"bool": {
				"must": [ %s ],
				"filter": [ %s ]
			}
		},
		"sort": [ %s ]
	}`, from, query.Limit, strings.Join(mainQueries, ","), strings.Join(filterQueries, ","), sortQuery)

	res, err := p.esClient.Search(
		p.esClient.Search.WithContext(ctx),
		p.esClient.Search.WithIndex("products"),
		p.esClient.Search.WithBody(strings.NewReader(mainQuery)),
		p.esClient.Search.WithPretty(),
	)
	if err != nil {
		log.Errorf("[SearchProduct] Elasticsearch search error: %v", err)
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Errorf("[SearchProduct] Failed to decode response: %v", err)
		return nil, 0, 0, err
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
