package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/config"
	httpclient "github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/http_client"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/message"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/error"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type PaymentServiceInterface interface {
	ProcessPayment(ctx context.Context, payment entity.PaymentEntity, accessToken, role string) (*entity.PaymentEntity, error)
	UpdateStatusByOrderCode(ctx context.Context, orderCode, status string) error
	GetAll(ctx context.Context, req entity.PaymentQueryStringRequest, accessToken string, role string) ([]entity.PaymentEntity, int64, int64, error)
	GetDetail(ctx context.Context, paymentID uuid.UUID, accessToken string, role string) (*entity.PaymentEntity, error)
	VerifyMidtransSignature(notification map[string]interface{}) (bool, error)
}

type PaymentService struct {
	repoPayment       repository.PaymentRepositoryInterface
	publisherRabbitMQ message.PublishRabbitMQInterface
	cfg               *config.Config
	midtrans          httpclient.MidtransClientInterface
	httpClient        httpclient.HttpClientToService
}

// VerifyMidtransSignature implements PaymentServiceInterface.
func (p *PaymentService) VerifyMidtransSignature(notification map[string]interface{}) (bool, error) {

	orderID, ok := notification["order_id"].(string)
	if !ok {
		return false, errs.ErrOrderIDNotFound
	}

	statusCode, ok := notification["status_code"].(string)
	if !ok {
		return false, errs.ErrStatusCodeNotFound
	}

	grossAmount, ok := notification["gross_amount"].(string)
	if !ok {
		return false, errs.ErrGrossAmountNotFound
	}

	signatureKey, ok := notification["signature_key"].(string)
	if !ok {
		return false, errs.ErrSignatureKeyNotFound
	}

	serverKey := p.cfg.Midtrans.ServerKey

	input := orderID + statusCode + grossAmount + serverKey

	hash := sha512.Sum512([]byte(input))
	expectedSignature := hex.EncodeToString(hash[:])

	log.Infof("[PaymentService-VerifySignature] OrderID=%s, StatusCode=%s, GrossAmount=%s", orderID, statusCode, grossAmount)
	log.Infof("[PaymentService-VerifySignature] Input for hash: %s", orderID+statusCode+grossAmount+"***")
	log.Infof("[PaymentService-VerifySignature] Expected signature: %s", expectedSignature)
	log.Infof("[PaymentService-VerifySignature] Received signature: %s", signatureKey)

	if expectedSignature != signatureKey {
		log.Errorf("[PaymentService-VerifySignature] Signature mismatch!")
		return false, nil
	}

	return true, nil

}

// GetDetail implements PaymentServiceInterface.
func (p *PaymentService) GetDetail(ctx context.Context, paymentID uuid.UUID, accessToken string, role string) (*entity.PaymentEntity, error) {
	result, err := p.repoPayment.GetDetail(ctx, paymentID)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-1: %v", err)
		return nil, err
	}

	isAdmin := false
	if role == "Super Admin" {
		isAdmin = true
	}

	orderDetail, err := p.httpClientOrderService(result.OrderID, accessToken, isAdmin)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-3: %v", err)
		return nil, err
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
func (p *PaymentService) GetAll(ctx context.Context, req entity.PaymentQueryStringRequest, accessToken string, role string) ([]entity.PaymentEntity, int64, int64, error) {

	results, totalData, totalPage, err := p.repoPayment.GetAll(ctx, req)
	if err != nil {
		log.Errorf("[PaymentService] GetAll: %v", err)
		return nil, 0, 0, err
	}

	isAdmin := false
	if role == "Super Admin" {
		isAdmin = true
	}

	for key, val := range results {
		orderDetail, err := p.httpClientOrderService(val.OrderID, accessToken, isAdmin)
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

	if status == "success" {

		if err := p.httpClientUpdateOrderStatus(orderCode, "Confirmed"); err != nil {
			log.Errorf("[PaymentService] Failed to update order status: %v", err)
		}

	}
	payment := entity.PaymentEntity{
		OrderID:       orderDetail,
		PaymentStatus: status,
	}

	if err := p.publisherRabbitMQ.PublishPaymentSuccess(payment); err != nil {
		log.Errorf("[PaymentService] Failed to publish to RabbitMQ: %v", err)
	}

	return nil
}

// ProcessPayment implements PaymentServiceInterface.
func (p PaymentService) ProcessPayment(ctx context.Context, payment entity.PaymentEntity, accessToken, role string) (*entity.PaymentEntity, error) {

	// err := p.repoPayment.GetByOrderID(ctx, payment.OrderID)
	// if err == nil {
	// 	log.Infof("[PaymentService] ProcessPayment-1: Payment already exists")
	// 	return nil, errs.ErrPaymentExist
	// }

	existingPayment, err := p.repoPayment.GetByOrderIDForUpdate(ctx, payment.OrderID)
	if err == nil && existingPayment != nil {
		log.Infof("[PaymentService] ProcessPayment: Payment already exists for order %s", payment.OrderID)
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

	if payment.PaymentMethod == "midtrans" {
		isAdmin := false
		if role == "Super Admin" {
			isAdmin = true
		}
		userResponse, err := p.httpClientUserService(payment.UserID, accessToken, isAdmin)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-5: %v", err)
			return nil, err
		}

		log.Infof("[PaymentService] User details: Name=%s, Email=%s", userResponse.Name, userResponse.Email)

		orderDetail, err := p.httpClientOrderService(payment.OrderID, accessToken, isAdmin)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-6: %v", err)
			return nil, err
		}

		log.Infof("[PaymentService] Order details: OrderCode=%s, Amount=%d", orderDetail.OrderCode, int64(payment.GrossAmount))

		transactionID, redirectURL, err := p.midtrans.CreateTransaction(int64(payment.GrossAmount), orderDetail.OrderCode, userResponse.Name, userResponse.Email)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-7: %v", err)
			return nil, err
		}

		log.Infof("[PaymentService] Midtrans transaction created: Token=%s, OrderCode=%s", transactionID, orderDetail.OrderCode)

		payment.PaymentStatus = "Pending"
		payment.PaymentGatewayID = transactionID
		payment.PaymentURL = redirectURL

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

func (o *PaymentService) httpClientUserService(userID int64, accessToken string, isAdmin bool) (*entity.CustomerResponseEntity, error) {

	baseUrlUser := fmt.Sprintf("%s/%s", o.cfg.App.APIGatewayUrl, "user/profile")
	if isAdmin {
		baseUrlUser = fmt.Sprintf("%s/%s", o.cfg.App.APIGatewayUrl, "admin/customers/"+strconv.FormatInt(userID, 10))
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

func (p *PaymentService) httpClientOrderService(orderId uuid.UUID, accessToken string, isAdmin bool) (*entity.OrderDetailHttpResponse, error) {
	baseUrlOrder := fmt.Sprintf("%s/%s", p.cfg.App.APIGatewayUrl, "auth/orders/"+orderId.String())

	if isAdmin {
		baseUrlOrder = fmt.Sprintf("%s/%s", p.cfg.App.APIGatewayUrl, "admin/orders/"+orderId.String())
	}

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
	baseUrlOrder := fmt.Sprintf("%s/public/orders/%s/code", p.cfg.App.APIGatewayUrl, orderCode)

	header := map[string]string{
		"Accept": "application/json",
	}

	dataOrder, err := p.httpClient.CallURL("GET", baseUrlOrder, header, nil)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-1: %v", err)
		return uuid.Nil, err
	}
	defer dataOrder.Body.Close()

	body, err := io.ReadAll(dataOrder.Body)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-2: %v", err)
		return uuid.Nil, err
	}

	if dataOrder.StatusCode != 200 {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-3: Non-200 status")
		return uuid.Nil, errs.ErrNotFoundOrder
	}

	var rawResponse map[string]interface{}
	err = json.Unmarshal(body, &rawResponse)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-4: Failed to unmarshal: %v", err)
		return uuid.Nil, err
	}

	success, ok := rawResponse["success"].(bool)
	if !ok || !success {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-5: API returned unsuccessful response")
		return uuid.Nil, errs.ErrNotFoundOrder
	}

	data, ok := rawResponse["data"].(map[string]interface{})
	if !ok {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-6: Invalid data structure in response")
		return uuid.Nil, fmt.Errorf("invalid response structure")
	}

	orderIDStr, ok := data["order_id"].(string)
	if !ok || orderIDStr == "" {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-7: order_id not found or empty in data")
		return uuid.Nil, fmt.Errorf("order ID not found in response")
	}

	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		log.Errorf("[PaymentService] httpClientOrderByCodeService-8: Invalid UUID format: %v", err)
		return uuid.Nil, fmt.Errorf("invalid order ID format: %v", err)
	}

	return orderUUID, nil
}

func (p *PaymentService) httpClientUpdateOrderStatus(orderCode, newStatus string) error {
	url := fmt.Sprintf("%s/public/orders/%s/status", p.cfg.App.APIGatewayUrl, orderCode)

	payload := map[string]interface{}{
		"status":  newStatus,
		"remarks": "Payment confirmed",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call order service: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	log.Infof("[PaymentService] Update order status response: StatusCode=%d, Body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("order service returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
