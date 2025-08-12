package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"order-service/config"
	httpclient "order-service/internal/adapter/http_client"
	"order-service/internal/adapter/repository"
	"order-service/internal/core/domain/entity"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type OrderServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
}

type OrderService struct {
	orderRepository repository.OrderRepositoryInterface
	cfg             *config.Config
	httpClient      httpclient.HttpClient
}

// GetAll implements OrderServiceInterface.
func (o *OrderService) GetAll(ctx context.Context, query entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error) {
	
	result, count, total, err := o.orderRepository.GetAll(ctx, query)
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0,err
	}
	
	for _, val := range result {
		for _, res := range val.OrderItems {
			baseUrl := fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "/admin/products/"+ uuid.UUID(*res.ProductID).String())
			header := map[string]string{
				"Authorization": "Bearer " + accessToken,
				"Accept": "application/json",
			}
			dataProduct, err := o.httpClient.CallURL("GET", baseUrl, header, nil)
			if err != nil {
				log.Errorf("[OrderService-2] GetAll: %v", err)
				return nil, 0, 0,err
			}
			
			defer dataProduct.Body.Close()
			
			body, err := io.ReadAll(dataProduct.Body)
			if err != nil {
				log.Errorf("[OrderService-3] GetAll: %v", err)
				return nil, 0, 0,err
			}
			
			var productResponse map[string]interface{}
			err = json.Unmarshal(body, &productResponse)
			if err != nil {
				log.Errorf("[OrderService-4] GetAll: %v", err)
				return nil, 0, 0,err
			}

			res.ProductImage = productResponse["product_image"].(string)

		}
	}

	return result, count, total, nil
}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, cfg *config.Config, httpClient httpclient.HttpClient) OrderServiceInterface {
	return &OrderService{
		orderRepository: orderRepo,
		cfg:             cfg,
		httpClient:      httpClient,
	}
}
