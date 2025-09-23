package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"payment-service/config"
	httpclient "payment-service/internal/adapter/http_client"
	"payment-service/internal/adapter/message"
	"payment-service/internal/adapter/repository"
	"payment-service/internal/core/domain/entity"
	errs "payment-service/internal/core/domain/error"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type PaymentServiceInterface interface {
	ProcessPayment(ctx context.Context, payment entity.PaymentEntity) (*entity.PaymentEntity, error)
	UpdateStatusByOrderCode(ctx context.Context, orderCode, status string) error
	GetAll(ctx context.Context, req entity.PaymentQueryStringRequest, accessToken string) ([]entity.PaymentEntity, int64, int64, error)
	GetDetail(ctx context.Context, paymentID uuid.UUID, accessToken string, role string) (*entity.PaymentEntity, error)
}

type PaymentService struct {
	repoPayment       repository.PaymentRepositoryInterface
	publisherRabbitMQ message.PublishRabbitMQInterface
	cfg               *config.Config
	midtrans          httpclient.MidtransClientInterface
	httpClient        httpclient.HttpClientToService
}

// GetDetail implements PaymentServiceInterface.
func (p *PaymentService) GetDetail(ctx context.Context, paymentID uuid.UUID, accessToken string, role string) (*entity.PaymentEntity, error) {
	result, err := p.repoPayment.GetDetail(ctx, paymentID)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-1: %v", err)
		return nil, err
	}

	orderDetail, err := p.httpClientOrderService(result.OrderID, accessToken)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-3: %v", err)
		return nil, err
	}

	isAdmin := false
	if role == "Super Admin" {
		isAdmin = true
	}

	userDetail, err := p.httpClientUserService(result.UserID, accessToken, isAdmin)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-4: %v", err)
		return nil, err
	}

	result.CustomerName = userDetail.Name
	result.CustomerEmail = userDetail.Email
	result.CustomerAddress = userDetail.Address

	result.OrderCode = orderDetail.OrderCode
	result.OrderShippingType = orderDetail.ShippingType
	result.OrderAt = orderDetail.OrderDatetime
	result.OrderRemarks = orderDetail.Remarks

	return result, nil
}

// GetAll implements PaymentServiceInterface.
func (p *PaymentService) GetAll(ctx context.Context, req entity.PaymentQueryStringRequest, accessToken string) ([]entity.PaymentEntity, int64, int64, error) {

	results, totalData, totalPage, err := p.repoPayment.GetAll(ctx, req)
	if err != nil {
		log.Errorf("[PaymentService] GetAll: %v", err)
		return nil, 0, 0, err
	}

	for key, val := range results {
		orderDetail, err := p.httpClientOrderService(val.OrderID, accessToken)
		if err != nil {
			log.Errorf("[PaymentService] GetAll-3: %v", err)
			return nil, 0, 0, err
		}
		results[key].OrderCode = orderDetail.OrderCode
		results[key].OrderShippingType = orderDetail.ShippingType
	}

	return results, totalData, totalPage, nil

}

// UpdateStatusByOrderCode implements PaymentServiceInterface.
func (p PaymentService) UpdateStatusByOrderCode(ctx context.Context, orderCode, status string) error {

	orderDetail, err := p.httpClientPublicOrderIDByCodeService(orderCode)
	if err != nil {
		log.Errorf("[PaymentService] UpdateStatusByOrderCode-1: %v", err)
		return err
	}

	if err := p.repoPayment.UpdateStatusByOrderCode(ctx, orderDetail, status); err != nil {
		log.Errorf("[PaymentService] UpdateStatusByOrderCode-2: %v", err)
		return err
	}

	return nil

}

// ProcessPayment implements PaymentServiceInterface.
func (p PaymentService) ProcessPayment(ctx context.Context, payment entity.PaymentEntity) (*entity.PaymentEntity, error) {

	err := p.repoPayment.GetByOrderID(ctx, payment.OrderID)
	if err == nil {
		log.Infof("[PaymentService] ProcessPayment-1: Payment already exists")
		return nil, errs.ErrPaymentExist
	}

	if payment.PaymentMethod == "cod" {
		payment.PaymentStatus = "success"

		if err := p.repoPayment.CreatePayment(ctx, payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-2: %v", err)
			return nil, err
		}

		if err := p.publisherRabbitMQ.PublishPaymentSuccess(payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-3: %v", err)
		}

		return &payment, nil
	}

	accessToken, err := p.getInternalToken()
	if err != nil {
		log.Errorf("[PaymentService-1] CreateOrder: %v", err)
		return nil, err
	}

	if payment.PaymentMethod == "midtrans" {
		userResponse, err := p.httpClientUserService(payment.UserID, accessToken, true)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-5: %v", err)
			return nil, err
		}

		orderDetail, err := p.httpClientOrderService(payment.OrderID, accessToken)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-6: %v", err)
			return nil, err
		}

		transactionID, err := p.midtrans.CreateTransaction(orderDetail.OrderCode, int64(payment.GrossAmount), userResponse.Name, userResponse.Email)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-7: %v", err)
			return nil, err
		}
		payment.PaymentStatus = "Pending"
		payment.PaymentGatewayID = transactionID

		if err := p.repoPayment.CreatePayment(ctx, payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-8: %v", err)
			return nil, err
		}

		if err := p.publisherRabbitMQ.PublishPaymentSuccess(payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-9: %v", err)
		}

		return &payment, nil
	}

	return nil, errs.ErrInvalidMethod

}

func NewPaymentService(repo repository.PaymentRepositoryInterface, cfg *config.Config, httpClientToService httpclient.HttpClientToService, midtrans httpclient.MidtransClientInterface, publisherRabbitMQ message.PublishRabbitMQInterface) PaymentServiceInterface {
	return &PaymentService{
		repoPayment:       repo,
		httpClient:        httpClientToService,
		midtrans:          midtrans,
		cfg:               cfg,
		publisherRabbitMQ: publisherRabbitMQ,
	}
}
func (o *PaymentService) getInternalToken() (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"client_id":     o.cfg.App.AuthClientID,
		"client_secret": o.cfg.App.AuthClientSecret,
	})
	if err != nil {
		log.Errorf("[PaymentService-1] getInternalToken: failed to marshal body: %v", err)
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
		log.Errorf("[PaymentService-2] getInternalToken: request failed: %v", err)
		return "", err
	}
	defer res.Body.Close()

	// tangani jika bukan 200
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("[PaymentService-3] getInternalToken: unexpected status %d, body: %s", res.StatusCode, string(body))
	}

	var tokenResp entity.InternalTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		log.Errorf("[PaymentService-4] getInternalToken: decode failed: %v", err)
		return "", err
	}

	if !tokenResp.Success || tokenResp.Data.AccessToken == "" {
		return "", fmt.Errorf("[PaymentService-5] getInternalToken: failed, msg: %s", tokenResp.Message)
	}

	return tokenResp.Data.AccessToken, nil
}

func (o *PaymentService) httpClientUserService(userID int64, accessToken string, isCustomer bool) (*entity.CustomerResponseEntity, error) {
	baseUrlUser := fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "admin/customers/"+strconv.FormatInt(userID, 10))

	if isCustomer {
		baseUrlUser = fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "user/profile")
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}
	dataUser, err := o.httpClient.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[PaymentService-1] httpClientUserService: %v", err)
		return nil, err
	}

	defer dataUser.Body.Close()

	body, err := io.ReadAll(dataUser.Body)
	if err != nil {
		log.Errorf("[PaymentService-2] httpClientUserService: %v", err)
		return nil, err
	}

	var userResponse entity.UserHttpClientResponse
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		log.Errorf("[PaymentService-3] httpClientUserService: %v", err)
		return nil, err
	}

	log.Infof("[PaymentService-UserResponse] Raw: %+v", userResponse)

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

func (p *PaymentService) httpClientOrderService(orderId uuid.UUID, accessToken string) (*entity.OrderDetailHttpResponse, error) {
	baseUrlOrder := fmt.Sprintf("%s/%s", p.cfg.App.OrderServiceUrl, "auth/orders/"+orderId.String())
	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}
	dataOrder, err := p.httpClient.CallURL("GET", baseUrlOrder, header, nil)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderService-1: %v", err)
		return nil, err
	}

	defer dataOrder.Body.Close()

	body, err := io.ReadAll(dataOrder.Body)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderService-2: %v", err)
		return nil, err
	}

	var orderDetail entity.OrderHttpClientResponse
	err = json.Unmarshal([]byte(body), &orderDetail)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderService-3: %v", err)
		return nil, err
	}

	if !orderDetail.Success {
		switch orderDetail.Code {
		case http.StatusNotFound:
			return nil, errs.ErrNotFoundOrder
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("order service auth error: %s", orderDetail.Message)
		default:
			return nil, fmt.Errorf("order service error (code %d): %s", orderDetail.Code, orderDetail.Message)
		}
	}

	return &orderDetail.Data, nil
}

func (p *PaymentService) httpClientPublicOrderIDByCodeService(orderCode string) (uuid.UUID, error) {
	baseUrlOrder := fmt.Sprintf("%s/%s", p.cfg.App.OrderServiceUrl, "public/orders/"+orderCode+"/code")
	header := map[string]string{
		"Accept": "application/json",
	}
	dataOrder, err := p.httpClient.CallURL("GET", baseUrlOrder, header, nil)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-1: %v", err)
		return uuid.Nil, err
	}

	defer dataOrder.Body.Close()

	if dataOrder.StatusCode != 200 {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-3: %v", err)
		return uuid.Nil, errs.ErrNotFoundOrder
	}

	body, err := io.ReadAll(dataOrder.Body)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-2: %v", err)
		return uuid.Nil, err
	}

	var orderDetail entity.GetOrderIDByCodeResponse
	err = json.Unmarshal([]byte(body), &orderDetail)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-4: %v", err)
		return uuid.Nil, err
	}

	if !orderDetail.Success {
		switch orderDetail.Code {
		case http.StatusNotFound:
			return uuid.Nil, errs.ErrNotFoundOrder
		default:
			return uuid.Nil, fmt.Errorf("order service error (code %d): %s", orderDetail.Code, orderDetail.Message)
		}
	}

	return orderDetail.Data.OrderID, nil
}
