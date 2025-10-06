package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/google/uuid"

	"github.com/elastic/go-elasticsearch/v7"
)

type ElasticRepositoryInterface interface {
	SearchOrderElastic(ctx context.Context, queryString entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
}

type elasticRepository struct {
	esClient *elasticsearch.Client
}

func NewElasticRepository(es *elasticsearch.Client) ElasticRepositoryInterface {
	return &elasticRepository{esClient: es}
}

// SearchOrderElasticByBuyerId implements ElasticRepositoryInterface.
func (e *elasticRepository) SearchOrderElasticByBuyerId(ctx context.Context, query entity.QueryStringEntity, buyerId int64) ([]entity.OrderEntity, int64, int64, error) {
	from := (query.Page - 1) * query.Limit

	statusFilter := ""
	if query.Status != "" {
		statusFilter = fmt.Sprintf(`{ "match": { "status": "%s" } },`, query.Status)
	}

	searchFilter := `{"match_all": {}}`
	if query.Search != "" {
		searchFilter = fmt.Sprintf(`{ "multi_match": { "query": "%s", "fields": ["order_code", "status", "buyer_name"] } }`, query.Search)
	}

	idFilter := ""
	if buyerId != 0 {
		idFilter = fmt.Sprintf(`{ "term": { "buyer_id": %d } },`, buyerId)
	}
	// Query Elasticsearch dengan filtering dan pagination
	mainQuery := fmt.Sprintf(`{
		"from": %d,
		"size": %d,
		"query": {
			"bool": {
				"must": [
					%s
					%s
					%s
				]
			}
		},
		"sort": [
			{ "id": "desc" }
		]
	}`, from, query.Limit, idFilter, statusFilter, searchFilter)

	// Kirim query ke Elasticsearch
	res, err := e.esClient.Search(
		e.esClient.Search.WithContext(ctx),
		e.esClient.Search.WithIndex("orders"),
		e.esClient.Search.WithBody(strings.NewReader(mainQuery)),
		e.esClient.Search.WithPretty(),
	)

	if err != nil {
		log.Printf("Error searching Elasticsearch: %s", err)
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error decoding response: %s", err)
			return nil, 0, 0, err
		}

		errType := e["error"].(map[string]interface{})["type"]
		if errType == "index_not_found_exception" {
			log.Printf("Index Not Found: %s", err)
			return nil, 0, 0, errors.New("index not found")
		}

		return nil, 0, 0, errors.New(e["error"].(string))
	}

	// Decode response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %s", err)
		return nil, 0, 0, err
	}

	// Ambil total data
	totalData := 0
	if hitsTotal, found := result["hits"].(map[string]interface{})["total"].(map[string]interface{}); found {
		totalData = int(hitsTotal["value"].(float64))
	}

	// Hitung total halaman
	totalPage := 0
	if query.Limit > 0 {
		totalPage = int(math.Ceil(float64(totalData) / float64(query.Limit)))
	}

	// Parsing hasil pencarian ke struct domain.Product
	orders := []entity.OrderEntity{}
	hits, found := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if found {
		for _, hit := range hits {
			source := hit.(map[string]interface{})["_source"]
			data, _ := json.Marshal(source)
			var order entity.OrderEntity
			json.Unmarshal(data, &order)
			orders = append(orders, order)
		}
	}

	return orders, int64(totalData), int64(totalPage), nil
}

// SearchOrderElastic implements ElasticRepositoryInterface.
// SearchOrderElastic implements ElasticRepositoryInterface.
func (e *elasticRepository) SearchOrderElastic(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {
	from := (query.Page - 1) * query.Limit

	statusFilter := ""
	if query.Status != "" {
		statusFilter = fmt.Sprintf(`{ "match": { "status": "%s" } },`, query.Status)
	}

	searchFilter := `{"match_all": {}}`
	if query.Search != "" {
		searchFilter = fmt.Sprintf(`{ "multi_match": { "query": "%s", "fields": ["order_code", "status", "buyer_name"] } }`, query.Search)
	}

	// PERBAIKAN: Ganti sort field dari "id" ke field yang tersedia
	mainQuery := fmt.Sprintf(`{
		"from": %d,
		"size": %d,
		"query": {
			"bool": {
				"must": [
					%s
					%s
				]
			}
		},
		"sort": [
			{ "created_at": "desc" }  // Ganti "id" dengan "created_at"
		]
	}`, from, query.Limit, statusFilter, searchFilter)

	// Kirim query ke Elasticsearch
	res, err := e.esClient.Search(
		e.esClient.Search.WithContext(ctx),
		e.esClient.Search.WithIndex("orders"),
		e.esClient.Search.WithBody(strings.NewReader(mainQuery)),
		e.esClient.Search.WithPretty(),
	)

	if err != nil {
		log.Printf("Error searching Elasticsearch: %s", err)
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error decoding response: %s", err)
			return nil, 0, 0, err
		}

		// Handle error response dengan aman
		var errorMsg string
		if errorObj, exists := e["error"]; exists {
			switch v := errorObj.(type) {
			case string:
				errorMsg = v
			case map[string]interface{}:
				// Convert map to string
				errorBytes, _ := json.Marshal(v)
				errorMsg = string(errorBytes)
			default:
				errorMsg = fmt.Sprintf("%v", v)
			}
		} else {
			errorMsg = "unknown error"
		}

		log.Printf("Elasticsearch Error: %s", errorMsg)
		return nil, 0, 0, errors.New(errorMsg)
	}

	// Decode response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %s", err)
		return nil, 0, 0, err
	}

	// Ambil total data
	totalData := 0
	if hitsTotal, found := result["hits"].(map[string]interface{})["total"].(map[string]interface{}); found {
		totalData = int(hitsTotal["value"].(float64))
	}

	// Hitung total halaman
	totalPage := 0
	if query.Limit > 0 {
		totalPage = int(math.Ceil(float64(totalData) / float64(query.Limit)))
	}

	// Parsing hasil pencarian ke struct domain.Product
	orders := []entity.OrderEntity{}
	hits, found := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if found {
		for _, hit := range hits {
			hitMap, ok := hit.(map[string]interface{})
			if !ok {
				log.Printf("Invalid hit type: %T", hit)
				continue
			}

			source, ok := hitMap["_source"].(map[string]interface{})
			if !ok {
				log.Printf("Invalid _source type: %T", hitMap["_source"])
				continue
			}

			// Gunakan method parsing yang aman
			order, err := e.parseOrderFromSource(source)
			if err != nil {
				log.Printf("Error parsing order from source: %v", err)
				continue
			}

			orders = append(orders, order)
		}
	}

	return orders, int64(totalData), int64(totalPage), nil
}

func (e *elasticRepository) parseOrderFromSource(source map[string]interface{}) (entity.OrderEntity, error) {
	var order entity.OrderEntity

	// Parse ID dengan error handling
	if id, ok := source["id"]; ok {
		switch v := id.(type) {
		case string:
			parsedID, err := uuid.Parse(v)
			if err != nil {
				return order, fmt.Errorf("error parsing UUID: %v", err)
			}
			order.ID = parsedID
		default:
			log.Printf("Unexpected type for id: %T", id)
		}
	}

	// Parse string fields
	if orderCode, ok := source["order_code"].(string); ok {
		order.OrderCode = orderCode
	}
	if status, ok := source["status"].(string); ok {
		order.Status = status
	}
	if buyerName, ok := source["buyer_name"].(string); ok {
		order.BuyerName = buyerName
	}

	// Parse numeric fields
	if totalAmount, ok := source["total_amount"].(float64); ok {
		order.TotalAmount = int64(totalAmount)
	}

	// Parse buyer_id
	if buyerID, ok := source["buyer_id"].(float64); ok {
		order.BuyerID = int64(buyerID)
	}

	// Parse order_items dengan safe type assertion
	if orderItems, ok := source["order_items"].([]interface{}); ok {
		for _, item := range orderItems {
			if itemMap, ok := item.(map[string]interface{}); ok {
				orderItem := entity.OrderItemEntity{}

				// Parse product_id
				if productID, ok := itemMap["product_id"].(string); ok {
					parsedProductID, err := uuid.Parse(productID)
					if err == nil {
						orderItem.ProductID = parsedProductID
					}
				}

				// Parse product_image
				if productImage, ok := itemMap["product_image"].(string); ok {
					orderItem.ProductImage = productImage
				}

				// Parse quantity
				if quantity, ok := itemMap["quantity"].(float64); ok {
					orderItem.Quantity = int64(quantity)
				}

				order.OrderItems = append(order.OrderItems, orderItem)
			}
		}
	}

	return order, nil
}
