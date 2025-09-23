package handler

import (
	"errors"
	"net/http"
	"payment-service/internal/adapter/handler/request"
	"payment-service/internal/adapter/handler/response"
	"payment-service/internal/core/domain/entity"
	errs "payment-service/internal/core/domain/error"
	"payment-service/internal/core/service"
	v "payment-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type PaymentHandlerInterface interface {
	Create(c echo.Context) error
	MidtransWebHook(c echo.Context) error
}

type PaymentHandler struct {
	paymentService service.PaymentServiceInterface
}

// MidtransWebHook implements PaymentHandlerInterface.
func (p *PaymentHandler) MidtransWebHook(c echo.Context) error {

	var notificationPayload map[string]interface{}
	if err := c.Bind(&notificationPayload); err != nil {
		log.Errorf("[PaymentHandler-1] MidtranswebHookHandler: %v", err)
		return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, err.Error()))
	}

	transactionStatus := notificationPayload["transaction_status"].(string)
	orderID := notificationPayload["order_id"].(string)

	newStatus := ""
	switch transactionStatus {
	case "capture", "settlement":
		newStatus = "success"
	case "deny", "cancel", "expire":
		newStatus = "failed"
	case "pending":
		newStatus = "pending"
	default:
		newStatus = "unknown"
	}

	if err := p.paymentService.UpdateStatusByOrderCode(c.Request().Context(), orderID, newStatus); err != nil {
		log.Errorf("[PaymentHandler-3] MidtranswebHookHandler: %v", err)
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.APIResponseSuccess(http.StatusOK, "success", nil))

}

// Create implements PaymentHandlerInterface.
func (p *PaymentHandler) Create(c echo.Context) error {
	var (
		ctx = c.Request().Context()
		req = request.PaymentRequest{}
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[PaymentHandler-2] Create: %v", err)
		return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, err.Error()))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[CategoryHandler-3] Create: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(
				http.StatusUnprocessableEntity,
				response.APIResponseError(http.StatusUnprocessableEntity, ve.Errors),
			)
		}

		return c.JSON(
			http.StatusUnprocessableEntity,
			response.APIResponseError(http.StatusUnprocessableEntity, err.Error()),
		)
	}

	paymentEntity := entity.PaymentEntity{
		OrderID:       req.OrderID,
		PaymentMethod: req.PaymentMethod,
		GrossAmount:   float64(req.GrossAmount),
		UserID:        req.UserID,
		Remarks:       req.Remarks,
	}
	result, err := p.paymentService.ProcessPayment(ctx, paymentEntity)
	if err != nil {
		log.Errorf("[PaymentHandler-4] Create: %v", err)

		if errors.Is(err, errs.ErrPaymentExist) {
			return c.JSON(http.StatusConflict, response.APIResponseError(http.StatusConflict, err.Error()))
		} else if errors.Is(err, errs.ErrInvalidMethod) {
			return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, err.Error()))
		}

		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	responPayment := map[string]interface{}{
		"payment_token": result.PaymentGatewayID,
	}

	return c.JSON(http.StatusCreated, response.APIResponseSuccess(http.StatusCreated, "Create payment success", responPayment))

}

func NewPaymentHandler(paymentService service.PaymentServiceInterface) PaymentHandlerInterface {
	return &PaymentHandler{
		paymentService: paymentService,
	}

}
