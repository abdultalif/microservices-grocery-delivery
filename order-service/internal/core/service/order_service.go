package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	httpclient "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/http_client"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/utils"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type OrderServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID uuid.UUID) (*entity.OrderEntity, error)
	Create(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error)
	UpdateStatus(ctx context.Context, req entity.OrderEntity) error
	DeleteOrderByID(ctx context.Context, orderID uuid.UUID) error

	GetAllCustomer(ctx context.Context, query entity.QueryStringEntity, tokenCustomer string) ([]entity.OrderEntity, int64, int64, error)
	GetOrderByOrderCode(ctx context.Context, orderCode string) (*entity.OrderEntity, error)
	GetDetailCustomer(ctx context.Context, orderID uuid.UUID, buyerID int64, accessToken string) (*entity.OrderEntity, error)

	GetInternalToken() (string, error)
	GetPublicOrderIDByOrderCode(ctx context.Context, orderCode string) (uuid.UUID, error)
	UpdateStatusByOrderCode(ctx context.Context, orderCode, status, remarks string) error

	resolveProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductResponseEntity, error)
	resolveBuyer(ctx context.Context, buyerID int64) (*entity.CustomerResponseEntity, error)
}

type OrderService struct {
	orderRepository repository.OrderRepositoryInterface
	elasticRepo     repository.ElasticRepositoryInterface
	localDataRepo   repository.LocalDataRepositoryInterface

	cfg               *config.Config
	httpClient        httpclient.HttpClient
	publisherRabbitMQ message.PublishRabbitMQInterface
	grpcClient        *GRPCClient
}

// UpdateStatusByOrderCode implements OrderServiceInterface.
func (o *OrderService) UpdateStatusByOrderCode(ctx context.Context, orderCode string, status string, remarks string) error {
	order, err := o.orderRepository.GetOrderByOrderCode(ctx, orderCode)
	if err != nil {
		log.Errorf("[OrderService] UpdateStatusByOrderCode-1: %v", err)
		return err
	}

	validTransitions := map[string][]string{
		"pending":   {"confirmed", "cancelled"},
		"confirmed": {"shipped", "cancelled"},
		"shipped":   {"delivered"},
	}

	currentStatus := strings.ToLower(order.Status)
	newStatus := strings.ToLower(status)

	if currentStatus == newStatus {
		log.Infof("[OrderService] Order %s already in status %s, skipping update", orderCode, status)
		return nil
	}

	if allowedStatuses, exists := validTransitions[currentStatus]; exists {
		isValid := false
		for _, allowed := range allowedStatuses {
			if allowed == newStatus {
				isValid = true
				break
			}
		}

		if !isValid {
			log.Errorf("[OrderService] Invalid status transition from %s to %s", currentStatus, newStatus)
			return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
		}
	}

	orderEntity := entity.OrderEntity{
		ID:      order.ID,
		Status:  status,
		Remarks: remarks,
	}

	err = o.UpdateStatus(ctx, orderEntity)
	if err != nil {
		log.Errorf("[OrderService] UpdateStatusByOrderCode-2: %v", err)
		return err
	}

	log.Infof("[OrderService] Successfully updated order %s to status %s", orderCode, status)
	return nil
}

// GetPublicOrderIDByOrderCode implements OrderServiceInterface.
func (o *OrderService) GetPublicOrderIDByOrderCode(ctx context.Context, orderCode string) (uuid.UUID, error) {

	result, err := o.orderRepository.GetOrderByOrderCode(ctx, orderCode)
	if err != nil {
		log.Errorf("[OrderService-1] GetPublicOrderIDByOrderCode: %v", err)
		return uuid.Nil, err
	}

	return result.ID, nil

}

// GetDetailCustomer implements OrderServiceInterface.
func (o *OrderService) GetDetailCustomer(ctx context.Context, orderID uuid.UUID, buyerID int64, accessToken string) (*entity.OrderEntity, error) {

	result, err := o.orderRepository.GetByIDCustomer(ctx, orderID, buyerID)
	if err != nil {
		log.Errorf("[OrderService-1] GetByID: %v", err)
		return nil, err
	}

	userResponse, err := o.httpClientUserService(result.BuyerID, accessToken, true)
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
		// yang dibawah versi komunikasi antar service melalui REST,
		productResponse, err := o.httpClientProductService(val.ProductID, accessToken, true)
		if err != nil {
			log.Errorf("[OrderService-3] GetByID: %v", err)
			return nil, err
		}

		result.OrderItems[key].ProductImage = productResponse.ProductImage
		if productResponse.Child != nil {
			result.OrderItems[key].ProductImage = productResponse.Child[0].Image
		}
		result.OrderItems[key].ProductName = productResponse.ProductName
		result.OrderItems[key].Price = int64(productResponse.SalePrice)
		result.OrderItems[key].ProductWeight = int64(productResponse.Weight)
		result.OrderItems[key].ProductUnit = productResponse.Unit
	}

	return result, nil

}

func (o *OrderService) DeleteOrderByID(ctx context.Context, orderID uuid.UUID) error {

	err := o.orderRepository.DeleteOrderByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-1] DeleteByID: %v", err)
		return err
	}

	err = o.publisherRabbitMQ.PublishDeleteOrderFromQueue(orderID)
	if err != nil {
		log.Errorf("[OrderService-2] DeleteByID: %v", err)
		return err
	}

	return nil

}

// GetOrderByOrderCode implements OrderServiceInterface.
func (o *OrderService) GetOrderByOrderCode(ctx context.Context, orderCode string) (*entity.OrderEntity, error) {
	result, err := o.orderRepository.GetOrderByOrderCode(ctx, orderCode)
	if err != nil {
		log.Errorf("[OrderService-1] GetOrderByOrderCode: %v", err)
		return nil, err
	}

	token, err := o.GetInternalToken()
	if err != nil {
		log.Errorf("[OrderService-1] GetOrderByOrderCode: %v", err)
		return nil, err
	}

	userResponse, err := o.httpClientUserService(result.BuyerID, token, false)
	if err != nil {
		log.Errorf("[OrderService-2] GetOrderByOrderCode: %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address
	result.BuyerLat = userResponse.Lat
	result.BuyerLng = userResponse.Lng

	for key, val := range result.OrderItems {
		// productResponse, err := o.GetProductFromSnapshoot(val.ProductID)
		// yang dibawah versi komunikasi antar service melalui REST,
		productResponse, err := o.httpClientProductService(val.ProductID, token, true)
		if err != nil {
			log.Errorf("[OrderService-3] GetOrderByOrderCode: %v", err)
			return nil, err
		}

		result.OrderItems[key].ProductImage = productResponse.ProductImage
		result.OrderItems[key].ProductName = productResponse.ProductName
		result.OrderItems[key].Price = int64(productResponse.SalePrice)
	}

	return result, nil
}

// GetAllCustomer implements OrderServiceInterface.
func (o *OrderService) GetAllCustomer(ctx context.Context, query entity.QueryStringEntity, tokenCustomer string) ([]entity.OrderEntity, int64, int64, error) {
	result, count, total, err := o.orderRepository.GetAll(ctx, query)
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	log.Infof("[OrderService-1] GetAllCustomer: Token Customer: %s\n", tokenCustomer)

	for key, val := range result {

		userResponse, err := o.httpClientUserService(val.BuyerID, tokenCustomer, true)
		if err != nil {
			log.Errorf("[OrderService-2] GetAll: %v", err)
			return nil, 0, 0, err
		}

		result[key].BuyerName = userResponse.Name

		for key2, res := range val.OrderItems {
			// yang dibawah versi komunikasi antar service melalui REST,
			productResponse, err := o.httpClientProductService(res.ProductID, tokenCustomer, true)
			if err != nil {
				log.Errorf("[OrderService-3] GetAll: %v", err)
				return nil, 0, 0, err
			}
			val.OrderItems[key2].ProductImage = productResponse.ProductImage
			val.OrderItems[key2].ProductName = productResponse.ProductName
			val.OrderItems[key2].ProductWeight = int64(productResponse.Weight)
			val.OrderItems[key2].ProductUnit = productResponse.Unit

		}
	}
	return result, count, total, nil
}

// UpdateStatus implements OrderServiceInterface.
func (o *OrderService) UpdateStatus(ctx context.Context, req entity.OrderEntity) error {

	currentOrder, err := o.orderRepository.GetByID(ctx, req.ID)
	if err != nil {
		log.Errorf("[OrderService] UpdateStatus-GetOrder: %v", err)
		return err
	}

	if strings.EqualFold(currentOrder.Status, req.Status) {
		log.Infof("[OrderService] Order %s already in status %s, skipping notification", req.ID, req.Status)
		return nil
	}

	accessToken, err := o.GetInternalToken()
	if err != nil {
		log.Errorf("[OrderService-1] UpdateStatus: %v", err)
		return err
	}

	buyerID, statusOrder, orderCode, err := o.orderRepository.UpdateStatus(ctx, req)
	if err != nil {
		log.Errorf("[OrderService-1] UpdateStatus: %v", err)
		return err
	}

	userResponse, err := o.httpClientUserService(buyerID, accessToken, false)
	if err != nil {
		log.Errorf("[OrderService-3] UpdateStatus: %v", err)
		return err
	}

	message := fmt.Sprintf("Hello,\n\nYour order with ID %s has been updated to status: %s.\n\nThank you for shopping with us!", orderCode, statusOrder)
	err = o.publisherRabbitMQ.PublishSendEmailUpdateStatus(userResponse.Email, message, o.cfg.Publisher.EmailUpdateStatus, buyerID)
	if err != nil {
		log.Errorf("[OrderService-4] UpdateStatus: %v", err)
		return err
	}
	go o.publisherRabbitMQ.PublishSendEmailUpdateStatus(userResponse.Email, message, o.cfg.Publisher.EmailUpdateStatus, buyerID)
	go o.publisherRabbitMQ.PublishSendPushNotifUpdateStatus(message, utils.PUSH_NOTIF, buyerID)
	go o.publisherRabbitMQ.PublishUpdateStatus(o.cfg.Publisher.PublisherUpdateStatus, req.ID, req.Status)

	return nil

}

// Create implements OrderServiceInterface.
func (o *OrderService) Create(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error) {

	// ini versi komunikasi antar service dengan rest api
	// token, err := o.GetInternalToken()
	// if err != nil {
	// 	log.Errorf("[OrderService-1] CreateOrder: %v", err)
	// 	return uuid.Nil, err
	// }

	// fmt.Printf("[OrderService-2] CreateOrder: Internal Token: %s\n", token)

	// _, err = o.httpClientUserService(req.BuyerID, token, false)

	buyer, err := o.resolveBuyer(ctx, req.BuyerID)
	if err != nil {
		log.Errorf("[OrderService-UserValidation] BuyerID %d not found: %v", req.BuyerID, err)
		return uuid.Nil, errs.ErrNotFoundBuyer
	}

	// Denormalisasi — snapshot buyer disimpan ke order
	req.BuyerName = buyer.Name
	req.BuyerEmail = buyer.Email
	req.BuyerPhone = buyer.Phone
	req.BuyerAddress = buyer.Address
	req.BuyerLat = buyer.Lat
	req.BuyerLng = buyer.Lng

	var notFoundProducts []string
	for _, item := range req.OrderItems {
		// _, err := o.httpClientProductService(item.ProductID, token, true)
		_, err := o.resolveProduct(ctx, item.ProductID)
		if err != nil {
			log.Errorf("[OrderService-ProductValidation] ProductID %s not found: %v", item.ProductID, err)
			notFoundProducts = append(notFoundProducts, item.ProductID.String())
			continue
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

	token, err := o.GetInternalToken()
	if err != nil {
		log.Errorf("[OrderService-1] CreateOrder: %v", err)
		return nil, err
	}

	userResponse, err := o.httpClientUserService(result.BuyerID, token, false)
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
		// productResponse, err := o.GetProductFromSnapshoot(val.ProductID)
		productResponse, err := o.httpClientProductService(val.ProductID, token, false)
		if err != nil {
			log.Errorf("[OrderService-3] GetByID: %v", err)
			return nil, err
		}

		result.OrderItems[key].ProductImage = productResponse.ProductImage
		result.OrderItems[key].ProductName = productResponse.ProductName
		result.OrderItems[key].Price = int64(productResponse.SalePrice)
		result.OrderItems[key].ProductWeight = int64(productResponse.Weight)
		result.OrderItems[key].ProductUnit = productResponse.Unit
		result.OrderItems[key].TotalPrice = int64(productResponse.SalePrice) * int64(val.Quantity)
		result.OrderItems[key].OrderCode = result.OrderCode
		result.OrderItems[key].OrderID = result.ID
	}

	return result, nil
}

// GetAll implements OrderServiceInterface.
// func (o *OrderService) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {

// 	// Sengaja buat versi grpc buat belajar
// 	return o.GetAllgRPC(ctx, query)
// }

// GetAll implements OrderServiceInterface. (versi komunikasi antar service melalui REST)
func (o *OrderService) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {

	result, count, total, err := o.elasticRepo.SearchOrderElastic(ctx, query)
	if err == nil {
		return result, count, total, nil
	} else {
		log.Errorf("[OrderService-1] GetAll: %v", err)
	}

	result, count, total, err = o.orderRepository.GetAll(ctx, query)
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	// token, err := o.GetInternalToken()
	// if err != nil {
	// 	log.Errorf("[OrderService-1] CreateOrder: %v", err)
	// 	return nil, 0, 0, err
	// }

	for key, val := range result {

		// userResponse, err := o.httlocalDataRepo.GetBuyer(ctx, val.BuyerID)
		userResponse, err := o.localDataRepo.GetBuyer(ctx, val.BuyerID)
		if err != nil {
			log.Errorf("[OrderService-2] GetAll: %v", err)
			return nil, 0, 0, err
		}

		result[key].BuyerName = userResponse.Name

		for key2, res := range val.OrderItems {
			productResponse, err := o.localDataRepo.GetProduct(ctx, res.ProductID)
			// productResponse, err := o.httpClientProductService(res.ProductID, token, false)
			if err != nil {
				log.Errorf("[OrderService-3] GetAll: %v", err)
				return nil, 0, 0, err
			}
			val.OrderItems[key2].ProductImage = productResponse.ProductImage

		}
	}
	return result, count, total, nil
}

func (o *OrderService) httpClientUserService(userID int64, accessToken string, isCustomer bool) (*entity.CustomerResponseEntity, error) {
	baseUrlUser := fmt.Sprintf("%s/%s", o.cfg.App.ApiGatewayServiceUrl, "admin/customers/"+strconv.FormatInt(userID, 10))
	// baseUrlUser := fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "admin/customers/"+strconv.FormatInt(userID, 10))

	if isCustomer {
		baseUrlUser = fmt.Sprintf("%s/%s", o.cfg.App.ApiGatewayServiceUrl, "user/profile")
		// baseUrlUser = fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "user/profile")
	}

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

func (o *OrderService) httpClientProductService(productID uuid.UUID, accessToken string, isCustomer bool) (*entity.ProductResponseEntity, error) {
	baseUrlProduct := fmt.Sprintf("%s/%s", o.cfg.App.ApiGatewayServiceUrl, "admin/products/"+uuid.UUID(productID).String())
	// baseUrlProduct := fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "admin/products/"+uuid.UUID(productID).String())

	if isCustomer {
		baseUrlProduct = fmt.Sprintf("%s/%s", o.cfg.App.ApiGatewayServiceUrl, "products/home"+"/"+uuid.UUID(productID).String())
		// baseUrlProduct = fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "products/home"+"/"+uuid.UUID(productID).String())
	}

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

func (o *OrderService) GetInternalToken() (string, error) {
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
		o.cfg.App.ApiGatewayServiceUrl+"/auth/service-token",
		// o.cfg.App.UserServiceUrl+"/auth/service-token",
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

func (o *OrderService) resolveBuyer(ctx context.Context, buyerID int64) (*entity.CustomerResponseEntity, error) {
	buyer, err := o.localDataRepo.GetBuyer(ctx, buyerID)
	if err == nil {
		log.Infof("[OrderService] Buyer %d resolved from local DB", buyerID)
		return buyer, nil
	}

	log.Warnf("[OrderService] Buyer %d not in local DB, fetching via HTTP...", buyerID)
	token, err := o.GetInternalToken()
	if err != nil {
		return nil, err
	}

	buyer, err = o.httpClientUserService(buyerID, token, false)
	if err != nil {
		return nil, err
	}

	// Simpan ke DB lokal untuk request berikutnya
	if saveErr := o.localDataRepo.UpsertBuyer(ctx, *buyer); saveErr != nil {
		log.Warnf("[OrderService] Failed to save buyer to local DB: %v", saveErr)
	}

	return buyer, nil
}

func (o *OrderService) resolveProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductResponseEntity, error) {
	product, err := o.localDataRepo.GetProduct(ctx, productID)
	if err == nil {
		log.Infof("[OrderService] Product %s resolved from local DB", productID)
		return product, nil
	}

	log.Warnf("[OrderService] Product %s not in local DB, fetching via HTTP...", productID)
	token, err := o.GetInternalToken()
	if err != nil {
		return nil, err
	}

	product, err = o.httpClientProductService(productID, token, true)
	if err != nil {
		return nil, err
	}

	if saveErr := o.localDataRepo.UpsertProduct(ctx, *product); saveErr != nil {
		log.Warnf("[OrderService] Failed to save product to local DB: %v", saveErr)
	}

	return product, nil
}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, cfg *config.Config, httpClient httpclient.HttpClient, publisherRabbitMQ message.PublishRabbitMQInterface, elasticRepo repository.ElasticRepositoryInterface, localDataRepo repository.LocalDataRepositoryInterface) OrderServiceInterface {

	grpcClient, err := NewGRPCClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create GRPC client: %v", err)
	}

	return &OrderService{
		orderRepository:   orderRepo,
		cfg:               cfg,
		httpClient:        httpClient,
		publisherRabbitMQ: publisherRabbitMQ,
		elasticRepo:       elasticRepo,
		grpcClient:        grpcClient,
		localDataRepo:     localDataRepo,
	}
}
