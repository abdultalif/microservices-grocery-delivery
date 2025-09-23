package handler

import (
	"errors"
	"net/http"
	"payment-service/internal/adapter/handler/request"
	"payment-service/internal/adapter/handler/response"
	"payment-service/internal/core/domain/entity"
	errs "payment-service/internal/core/domain/error"
	"payment-service/internal/core/service"
	"payment-service/utils/conv"
	v "payment-service/utils/validator"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type PaymentHandlerInterface interface {
	Create(c echo.Context) error
	MidtransWebHook(c echo.Context) error
	GetAllCustomer(c echo.Context) error
	GetAllAdmin(c echo.Context) error
	GetDetail(c echo.Context) error
}

type PaymentHandler struct {
	paymentService service.PaymentServiceInterface
}

// GetDetail implements PaymentHandlerInterface.
func (p *PaymentHandler) GetDetail(c echo.Context) error {

	var (
		ctx = c.Request().Context()
		res = response.PaymentDetailResponse{}
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[PaymentHandler-1] GetDetail: Invalid user data")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "Invalid user data"))
	}

	paymentID, err := uuid.Parse(c.Param("paymentID"))
	if err != nil {
		log.Errorf("[PaymentHandler-2] GetDetail: %v", err)
		return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, "Payment ID is not valid UUID"))
	}

	result, err := p.paymentService.GetDetail(ctx, paymentID, user.Token, user.Role)
	if err != nil {
		log.Errorf("[PaymentHandler-4] GetDetail: %v", err)
		if errors.Is(err, errs.ErrNotFoundPayment) {
			return c.JSON(http.StatusNotFound, response.APIResponseError(http.StatusNotFound, err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	res.ID = result.ID
	res.OrderCode = result.OrderCode
	res.PaymentMethod = result.PaymentMethod
	res.PaymentStatus = result.PaymentStatus
	res.GrossAmount = result.GrossAmount
	res.ShippingType = result.OrderShippingType
	res.PaymentAt = result.PaymentAt
	res.OrderAt = result.OrderAt
	res.OrderRemarks = result.OrderRemarks
	res.CustomerName = result.CustomerName
	res.CustomerAddress = result.CustomerAddress

	return c.JSON(http.StatusOK, response.APIResponseSuccess(http.StatusOK, "Get payment detail success", res))

}

// GetAllAdmin implements PaymentHandlerInterface.
func (p *PaymentHandler) GetAllAdmin(c echo.Context) error {
	var (
		ctx   = c.Request().Context()
		resps = []response.PaymentListResponse{}
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[PaymentHandler-1] GetAll: Invalid user data")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "Invalid user data"))
	}

	userID := 0
	search := c.QueryParam("search")
	var page int64 = 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var perPage int64 = 10
	if perPageStr := c.QueryParam("perPage"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	status := ""
	if statusStr := c.QueryParam("status"); statusStr != "" {
		status = statusStr
	}

	orderBy := "created_at"
	if orderByStr := c.QueryParam("orderBy"); orderByStr != "" {
		orderBy = orderByStr
	}

	orderType := "desc"
	if orderTypeStr := c.QueryParam("orderType"); orderTypeStr != "" {
		orderType = orderTypeStr
	}

	reqEntity := entity.PaymentQueryStringRequest{
		Search:    search,
		Status:    status,
		Page:      page,
		Limit:     perPage,
		OrderBy:   orderBy,
		OrderType: orderType,
		UserID:    int64(userID),
	}

	results, count, total, err := p.paymentService.GetAll(ctx, reqEntity, user.Token)
	if err != nil {
		if errors.Is(err, errs.ErrNotFoundPayment) {
			return c.JSON(http.StatusOK, response.APIResponseError(http.StatusOK, err.Error()))
		}
		log.Errorf("[PaymentHandler-3] GetAll: %v", err)
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error))
	}

	for _, val := range results {
		resps = append(resps, response.PaymentListResponse{
			ID:            val.ID,
			OrderCode:     val.OrderCode,
			PaymentStatus: val.PaymentStatus,
			PaymentMethod: val.PaymentMethod,
			GrossAmount:   val.GrossAmount,
			ShippingType:  val.OrderShippingType,
		})
	}

	return c.JSON(http.StatusOK, response.DefaultResponseWithPaginations{
		Success: true,
		Code:    http.StatusOK,
		Message: "Get all payment success",
		Pagination: &response.Pagination{
			Page:       page,
			TotalCount: count,
			PerPage:    perPage,
			TotalPage:  total,
		},
		Data: resps,
	})
}

// GetAllCustomer implements PaymentHandlerInterface.
func (p *PaymentHandler) GetAllCustomer(c echo.Context) error {
	var (
		ctx = c.Request().Context()
		res = []response.PaymentListResponse{}
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[PaymentHandler-1] GetAllCustomer: Invalid user data")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "Invalid user data"))
	}

	search := c.QueryParam("search")
	var page int64 = 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var perPage int64 = 10
	if perPageStr := c.QueryParam("perPage"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	status := ""
	if statusStr := c.QueryParam("status"); statusStr != "" {
		status = statusStr
	}

	orderBy := "created_at"
	if orderByStr := c.QueryParam("orderBy"); orderByStr != "" {
		orderBy = orderByStr
	}

	orderType := "desc"
	if orderTypeStr := c.QueryParam("orderType"); orderTypeStr != "" {
		orderType = orderTypeStr
	}

	reqEntity := entity.PaymentQueryStringRequest{
		Search:    search,
		Status:    status,
		Page:      page,
		Limit:     perPage,
		OrderBy:   orderBy,
		OrderType: orderType,
		UserID:    user.UserID,
	}

	results, count, total, err := p.paymentService.GetAll(ctx, reqEntity, user.Token)
	if err != nil {
		if errors.Is(err, errs.ErrNotFoundPayment) {
			return c.JSON(http.StatusOK, response.APIResponseError(http.StatusOK, err.Error()))
		}
		log.Errorf("[PaymentHandler-3] GetAll: %v", err)
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	for _, val := range results {
		res = append(res, response.PaymentListResponse{
			ID:            val.ID,
			OrderCode:     val.OrderCode,
			PaymentStatus: val.PaymentStatus,
			PaymentMethod: val.PaymentMethod,
			GrossAmount:   val.GrossAmount,
			ShippingType:  val.OrderShippingType,
		})
	}

	return c.JSON(http.StatusOK, response.DefaultResponseWithPaginations{
		Success: true,
		Code:    http.StatusOK,
		Message: "Get all payment success",
		Pagination: &response.Pagination{
			Page:       page,
			TotalCount: count,
			PerPage:    perPage,
			TotalPage:  total,
		},
		Data: res,
	})

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
