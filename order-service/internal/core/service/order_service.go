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
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type OrderServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID uuid.UUID, accessToken string) (*entity.OrderEntity, error)
}

type OrderService struct {
	orderRepository repository.OrderRepositoryInterface
	cfg             *config.Config
	httpClient      httpclient.HttpClient
}

// GetByID implements OrderServiceInterface.
func (o *OrderService) GetByID(ctx context.Context, orderID uuid.UUID, accessToken string) (*entity.OrderEntity, error) {
	result, err := o.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-1] GetByID: %v", err)
		return nil, err
	}

	userResponse, err := o.httpClientUserService(result.BuyerID, accessToken)
	if err != nil {
		log.Errorf("[OrderService-2] GetByID: %v", err)
		return nil, err
	}

	result.BuyerName = userResponse["name"].(string)
	result.BuyerEmail = userResponse["email"].(string)
	result.BuyerPhone = userResponse["phone"].(string)
	result.BuyerAddress = userResponse["address"].(string)
	result.BuyerLat = userResponse["lat"].(string)
	result.BuyerLng = userResponse["lng"].(string)

	for _, val := range result.OrderItems{
		productResponse, err := o.httpClientProductService(*val.ProductID, accessToken)
		if err != nil {
			log.Errorf("[OrderService-3] GetByID: %v", err)
			return nil, err
		}

		val.ProductImage = productResponse["product_image"].(string)
		val.ProductName = productResponse["product_name"].(string)
		val.Price = productResponse["sale_price"].(int64)
	}

	return result, nil
}

// GetAll implements OrderServiceInterface.
func (o *OrderService) GetAll(ctx context.Context, query entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error) {

	result, count, total, err := o.orderRepository.GetAll(ctx, query)
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	for _, val := range result {
		
		userResponse, err := o.httpClientUserService(val.BuyerID, accessToken)
		if err != nil {
			log.Errorf("[OrderService-2] GetAll: %v", err)
			return nil, 0, 0, err
		}

		val.BuyerName = userResponse["name"].(string)
	
		for _, res := range val.OrderItems {
			productResponse, err := o.httpClientProductService(*res.ProductID, accessToken)
			if err != nil {
				log.Errorf("[OrderService-3] GetAll: %v", err)
				return nil, 0, 0, err
			}
			res.ProductImage = productResponse["product_image"].(string)

		}
	}
	return result, count, total, nil
}

func (o *OrderService) httpClientUserService(userID int64, accessToken string) (map[string]interface{}, error) {
	baseUrlUser := fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "admin/customers/"+strconv.FormatInt(userID, 10))
	header := map[string]string{
			"Authorization": "Bearer " + accessToken,
			"Accept":        "application/json",
		}
	dataUser, err := o.httpClient.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[OrderService-1] httpClientUserService: %v", err)
		return nil, err
	}

	defer dataUser.Body.Close()

	body, err := io.ReadAll(dataUser.Body)
	if err != nil {
		log.Errorf("[OrderService-2] httpClientUserService: %v", err)
		return nil, err
	}

	var userResponse map[string]interface{}
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientUserService: %v", err)	
		return nil, err
	}

	return userResponse, nil
}

func (o *OrderService) httpClientProductService(productID uuid.UUID, accessToken string) (map[string]interface{}, error) {
	baseUrlProduct := fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "/admin/products/"+uuid.UUID(productID).String())
	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}
	dataProduct, err := o.httpClient.CallURL("GET", baseUrlProduct, header, nil)
	if err != nil {
		log.Errorf("[OrderService-1] httpClientProductService: %v", err)
		return nil, err
	}

	defer dataProduct.Body.Close()

	body, err := io.ReadAll(dataProduct.Body)
	if err != nil {
		log.Errorf("[OrderService-2] httpClientProductService: %v", err)
		return nil, err
	}

	var productResponse map[string]interface{}
	err = json.Unmarshal(body, &productResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientProductService: %v", err)
		return nil, err
	}

	return productResponse, nil

}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, cfg *config.Config, httpClient httpclient.HttpClient) OrderServiceInterface {
	return &OrderService{
		orderRepository: orderRepo,
		cfg:             cfg,
		httpClient:      httpClient,
	}
}
