package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"order-service/config"
	httpclient "order-service/internal/adapter/http_client"
	"order-service/internal/adapter/repository"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/utils/conv"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type OrderServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID uuid.UUID) (*entity.OrderEntity, error)
	Create(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error)
}

type OrderService struct {
	orderRepository repository.OrderRepositoryInterface
	cfg             *config.Config
	httpClient      httpclient.HttpClient
	elasticRepo     repository.ElasticRepositoryInterface
}

// Create implements OrderServiceInterface.
func (o *OrderService) Create(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error) {

	token, err := o.getInternalToken()
	if err != nil {
		log.Errorf("[OrderService-1] CreateOrder: %v", err)
		return uuid.Nil, err
	}

	fmt.Printf("[OrderService-2] CreateOrder: Internal Token: %s\n", token)

	_, err = o.httpClientUserService(req.BuyerID, token)
	if err != nil {
		log.Errorf("[OrderService-UserValidation] BuyerID %d not found: %v", req.BuyerID, err)
		return uuid.Nil, err
	}

	var notFoundProducts []string
	for _, item := range req.OrderItems {
		_, err := o.httpClientProductService(item.ProductID, token)
		if err != nil {
			log.Errorf("[OrderService-ProductValidation] ProductID %s not found: %v", item.ProductID, err)
			notFoundProducts = append(notFoundProducts, item.ProductID.String())
		}
	}

	if len(notFoundProducts) > 0 {
		return uuid.Nil, fmt.Errorf("%w: %v", errs.ErrNotFoundProduct, notFoundProducts)
	}

	req.OrderCode = conv.GenerateOrderCode()
	shippingFee := 0
	if req.ShippingType == "Delivery" {
		shippingFee = 5000
	}
	req.ShippingFee = int64(shippingFee)
	req.Status = "Pending"
	orderID, err := o.orderRepository.CreateOrder(ctx, req)
	if err != nil {
		log.Errorf("[OrderService-1] CreateOrder: %v", err)
		return uuid.Nil, err
	}

	resultData, err := o.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-2] CreateOrder: %v", err)
	}

	if err = o.publisherRabbitMQ.PublishOrderToQueue(*resultData); err != nil {
		log.Errorf("[OrderService-3] CreateOrder: %v", err)
	}

	for _, orderItem := range req.OrderItems {
		o.publisherRabbitMQ.PublishUpdateStock(orderItem.ProductID, orderItem.Quantity)
	}

	return orderID, nil
}

// GetByID implements OrderServiceInterface.
func (o *OrderService) GetByID(ctx context.Context, orderID uuid.UUID) (*entity.OrderEntity, error) {
	result, err := o.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-1] GetByID: %v", err)
		return nil, err
	}

	token, err := o.getInternalToken()
	if err != nil {
		log.Errorf("[OrderService-1] CreateOrder: %v", err)
		return nil, err
	}

	userResponse, err := o.httpClientUserService(result.BuyerID, token)
	if err != nil {
		log.Errorf("[OrderService-2] GetByID: %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address
	result.BuyerLat = userResponse.Lat
	result.BuyerLng = userResponse.Lng

	for key, val := range result.OrderItems {
		productResponse, err := o.httpClientProductService(val.ProductID, token)
		if err != nil {
			log.Errorf("[OrderService-3] GetByID: %v", err)
			return nil, err
		}

		result.OrderItems[key].ProductImage = productResponse.ProductImage
		result.OrderItems[key].ProductName = productResponse.ProductName
		result.OrderItems[key].Price = int64(productResponse.SalePrice)
	}

	return result, nil
}

// GetAll implements OrderServiceInterface.
func (o *OrderService) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {

	results, count, total, err := o.elasticRepo.SearchOrderElastic(ctx, queryString)
	if err == nil {
		return results, count, total, nil
	} else {
		log.Errorf("[OrderService-1] GetAll: %v", err)
	}

	result, count, total, err := o.orderRepository.GetAll(ctx, query)
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	token, err := o.getInternalToken()
	if err != nil {
		log.Errorf("[OrderService-1] CreateOrder: %v", err)
		return nil, 0, 0, err
	}

	for key, val := range result {

		userResponse, err := o.httpClientUserService(val.BuyerID, token)
		if err != nil {
			log.Errorf("[OrderService-2] GetAll: %v", err)
			return nil, 0, 0, err
		}

		result[key].BuyerName = userResponse.Name

		for key2, res := range val.OrderItems {
			productResponse, err := o.httpClientProductService(res.ProductID, token)
			if err != nil {
				log.Errorf("[OrderService-3] GetAll: %v", err)
				return nil, 0, 0, err
			}
			val.OrderItems[key2].ProductImage = productResponse.ProductImage

		}
	}
	return result, count, total, nil
}

func (o *OrderService) httpClientUserService(userID int64, accessToken string) (*entity.CustomerResponseEntity, error) {
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

	var userResponse entity.UserHttpClientResponse
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientUserService: %v", err)
		return nil, err
	}

	log.Infof("[OrderService-UserResponse] Raw: %+v", userResponse)

	// Bedakan kasus berdasarkan code
	if !userResponse.Success {
		switch userResponse.Code {
		case http.StatusNotFound:
			return nil, errs.ErrNotFoundBuyer
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("user service auth error: %s", userResponse.Message)
		default:
			return nil, fmt.Errorf("user service error (code %d): %s", userResponse.Code, userResponse.Message)
		}
	}
	return &userResponse.Data, nil
}

func (o *OrderService) httpClientProductService(productID uuid.UUID, accessToken string) (*entity.ProductResponseEntity, error) {
	baseUrlProduct := fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "admin/products/"+uuid.UUID(productID).String())
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

	var productResponse entity.ProductHttpClientResponse
	err = json.Unmarshal(body, &productResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientProductService: %v", err)
		return nil, err
	}

	if !productResponse.Success {
		return nil, errs.ErrNotFoundProduct
	}

	return &productResponse.Data, nil

}

func (o *OrderService) getInternalToken() (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"client_id":     o.cfg.App.AuthClientID,
		"client_secret": o.cfg.App.AuthClientSecret,
	})
	if err != nil {
		log.Errorf("[OrderService-1] getInternalToken: failed to marshal body: %v", err)
		return "", err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	res, err := o.httpClient.CallURL(
		"POST",
		o.cfg.App.UserServiceUrl+"/auth/service-token",
		headers,
		reqBody,
	)
	if err != nil {
		log.Errorf("[OrderService-2] getInternalToken: request failed: %v", err)
		return "", err
	}
	defer res.Body.Close()

	// tangani jika bukan 200
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("[OrderService-3] getInternalToken: unexpected status %d, body: %s", res.StatusCode, string(body))
	}

	var tokenResp entity.InternalTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		log.Errorf("[OrderService-4] getInternalToken: decode failed: %v", err)
		return "", err
	}

	if !tokenResp.Success || tokenResp.Data.AccessToken == "" {
		return "", fmt.Errorf("[OrderService-5] getInternalToken: failed, msg: %s", tokenResp.Message)
	}

	return tokenResp.Data.AccessToken, nil
}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, cfg *config.Config, httpClient httpclient.HttpClient, publisherRabbitMQ message.PublishRabbitMQInterface, elasticRepo repository.ElasticRepositoryInterface) OrderServiceInterface {
	return &OrderService{
		orderRepository:   orderRepo,
		cfg:               cfg,
		httpClient:        httpClient,
		publisherRabbitMQ: publisherRabbitMQ,
		elasticRepo:       elasticRepo,
	}
}
