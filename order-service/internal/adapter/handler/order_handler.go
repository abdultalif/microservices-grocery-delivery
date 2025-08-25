package handler

import (
	"errors"
	"net/http"
	"order-service/config"
	"order-service/internal/adapter/handler/request"
	"order-service/internal/adapter/handler/response"
	"order-service/internal/adapter/middleware"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/internal/core/service"
	"order-service/utils/conv"
	v "order-service/utils/validator"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OrderHandlerInterface interface {
	GetAll(e echo.Context) error
	GetByID(e echo.Context) error
	Create(e echo.Context) error
	UpdateStatus(e echo.Context) error

	GetAllCustomer(e echo.Context) error
}

type OrderHandler struct {
	orderService service.OrderServiceInterface
}

// GetAllCustomer implements OrderHandlerInterface.
func (o *OrderHandler) GetAllCustomer(e echo.Context) error {
	var (
		ctx      = e.Request().Context()
		resOrder = []response.OrderCustomerList{}
	)

	tokenCustomer := e.Get("token").(string)
	if tokenCustomer == "" {
		log.Errorf("[OrderHandler-1] GetAllCustomer: %s", "Token is empty")
		return e.JSON(http.StatusUnauthorized, response.ResponseAPI(false, http.StatusUnauthorized, "Token is empty", nil))
	}

	user, ok := e.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CustomerHandler-1] CreateCustomer: user data not found in context")
		return e.JSON(http.StatusUnauthorized, response.ResponseAPI(false, http.StatusUnauthorized, "user data not found in context", nil))
	}

	userID := user.UserID

	search := e.QueryParam("search")
	var page int64 = 1
	if pageStr := e.QueryParam("page"); pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var perPage int64 = 10
	if perPageStr := e.QueryParam("perPage"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	status := ""
	if statusStr := e.QueryParam("status"); statusStr != "" {
		status = statusStr
	}

	reqEntity := entity.QueryStringEntity{
		Search:  search,
		Status:  status,
		Page:    page,
		Limit:   perPage,
		BuyerID: userID,
	}

	results, totalData, totalPage, err := o.orderService.GetAllCustomer(ctx, reqEntity, tokenCustomer)
	if err != nil {
		log.Errorf("[OrderHandler-3] GetAllCustomer: %v", err)
		if errors.Is(err, errs.ErrNotFoundOrder) {
			return e.JSON(http.StatusNotFound, response.ResponseAPI(false, http.StatusNotFound, "Order not found", nil))
		}
		return e.JSON(http.StatusInternalServerError, response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	for _, result := range results {

		if len(result.OrderItems) == 0 {
			log.Warnf("[OrderHandler] Order %s has no order items, skipping", result.OrderCode)
			continue
		}

		resOrder = append(resOrder, response.OrderCustomerList{
			ID:            result.ID,
			OrderCode:     result.OrderCode,
			Status:        result.Status,
			ProductName:   result.OrderItems[0].ProductName,
			TotalAmount:   result.TotalAmount,
			ProductImage:  result.OrderItems[0].ProductImage,
			Weight:        result.OrderItems[0].ProductWeight,
			Unit:          result.OrderItems[0].ProductUnit,
			Quantity:      result.OrderItems[0].Quantity,
			OrderDateTime: result.OrderDate,
		})
	}

	return e.JSON(
		http.StatusOK,
		response.ResponseAPIWithPagination(true, http.StatusOK, "Success", resOrder, page, totalData, totalPage, perPage),
	)
}

// UpdateStatus implements OrderHandlerInterface.
func (o *OrderHandler) UpdateStatus(e echo.Context) error {
	var (
		ctx = e.Request().Context()
		req = request.OrderUpdateStatusRequest{}
	)

	if err := e.Bind(&req); err != nil {
		log.Errorf("[OrderHandler-1] UpdateStatus: %v", err)
		return e.JSON(
			http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body", nil),
		)
	}

	if err := e.Validate(req); err != nil {
		log.Errorf("[OrderHandler-2] UpdateStatus: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return e.JSON(
				http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil),
			)
		}

		return e.JSON(
			http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil),
		)
	}

	orderID, err := uuid.Parse(e.Param("orderID"))
	if err != nil {
		log.Errorf("[OrderHandler-3] UpdateStatus: OrderID must be uuid")
		return e.JSON(
			http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "OrderID must be uuid", nil),
		)
	}

	reqEntity := entity.OrderEntity{
		Remarks: req.Remarks,
		Status:  req.Status,
		ID:      orderID,
	}

	err = o.orderService.UpdateStatus(ctx, reqEntity)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidStatus) {
			return e.JSON(http.StatusBadRequest, response.ResponseAPI(false, http.StatusBadRequest, "Invalid status transition", nil))
		} else if errors.Is(err, errs.ErrNotFoundOrder) {
			return e.JSON(http.StatusNotFound, response.ResponseAPI(false, http.StatusNotFound, "Order not found", nil))
		} else {
			return e.JSON(http.StatusInternalServerError, response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
		}
	}

	return e.JSON(http.StatusCreated, response.ResponseAPI(true, http.StatusCreated, "Success", nil))

}

// Create implements OrderHandlerInterface.
func (o *OrderHandler) Create(e echo.Context) error {
	var (
		ctx = e.Request().Context()
		req = request.CreateOrderRequest{}
	)

	if err := e.Bind(&req); err != nil {
		log.Errorf("[CategoryHandler-2] Create: %v", err)
		return e.JSON(
			http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body", nil),
		)
	}

	if err := e.Validate(req); err != nil {
		log.Errorf("[CategoryHandler-3] Create: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return e.JSON(
				http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil),
			)
		}

		return e.JSON(
			http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil),
		)
	}

	reqEntity := entity.OrderEntity{
		BuyerID:      req.BuyerID,
		OrderDate:    req.OrderDate,
		TotalAmount:  req.TotalAmount,
		ShippingType: req.ShippingType,
		Remarks:      req.Remarks,
		OrderTime:    req.OrderTime,
	}

	orderDetails := []entity.OrderItemEntity{}
	for _, val := range req.OrderDetails {
		orderDetails = append(orderDetails, entity.OrderItemEntity{
			ProductID: val.ProductID,
			Quantity:  val.Quantity,
		})
	}

	reqEntity.OrderItems = orderDetails

	orderID, err := o.orderService.Create(ctx, reqEntity)
	if err != nil {
		log.Errorf("[OrderHandler-4] CreateOrder: %v", err)
		if errors.Is(err, errs.ErrNotFoundBuyer) {
			return e.JSON(
				http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "Buyer not found", nil),
			)
		} else if errors.Is(err, errs.ErrNotFoundProduct) {
			return e.JSON(
				http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, err.Error(), nil),
			)
		} else {
			return e.JSON(
				http.StatusInternalServerError,
				response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil),
			)
		}

	}

	return e.JSON(
		http.StatusCreated,
		response.ResponseAPI(true, http.StatusCreated, "success", map[string]interface{}{
			"order_id": orderID,
		}),
	)
}

// GetByID implements OrderHandlerInterface.
func (o *OrderHandler) GetByID(e echo.Context) error {
	var (
		ctx      = e.Request().Context()
		resOrder = response.OrderAdminDetail{}
	)

	orderID, err := uuid.Parse(e.Param("orderID"))
	if err != nil {
		log.Errorf("[OrderHandler-1] GetByID: OrderID must be uuid")
		return e.JSON(
			http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "OrderID must be uuid", nil),
		)
	}

	order, err := o.orderService.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderHandler-1] GetByID: %v", err)
		if errors.Is(err, errs.ErrNotFoundOrder) {
			return e.JSON(
				http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "Order not found", nil),
			)
		}
		return e.JSON(
			http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil),
		)
	}

	resOrder.ID = order.ID
	resOrder.OrderCode = order.OrderCode
	resOrder.Status = order.Status
	// resOrder.PaymentMethod = order.Pa
	resOrder.TotalAmount = order.TotalAmount
	resOrder.OrderDatetime = order.OrderDate
	resOrder.ShippingFee = order.ShippingFee
	resOrder.ShippingType = order.ShippingType
	resOrder.Remarks = order.Remarks
	resOrder.Customer = response.CustomerOrder{
		CustomerID:      int64(order.BuyerID),
		CustomerName:    order.BuyerName,
		CustomerEmail:   order.BuyerEmail,
		CustomerPhone:   order.BuyerPhone,
		CustomerAddress: order.BuyerAddress,
	}

	for _, item := range order.OrderItems {
		resOrder.OrderDetail = append(resOrder.OrderDetail, response.OrderDetail{
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			ProductPrice: item.Price,
			Quantity:     item.Quantity,
		})
	}

	return e.JSON(
		http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, nil, resOrder),
	)

}

// GetAll implements OrderHandlerInterface.
func (o *OrderHandler) GetAll(e echo.Context) error {
	var (
		ctx       = e.Request().Context()
		resOrders = []response.OrderAdminList{}
	)

	search := e.QueryParam("search")
	var page int64 = 1
	if pageStr := e.QueryParam("page"); pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var perPage int64 = 10
	if perPageStr := e.QueryParam("perPage"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	status := ""
	if statusStr := e.QueryParam("status"); statusStr != "" {
		status = statusStr
	}

	reqEntity := entity.QueryStringEntity{
		Search: search,
		Status: status,
		Page:   page,
		Limit:  perPage,
	}

	results, totalData, totalPage, err := o.orderService.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[OrderHandler-1] GetAll: %v", err)
		if errors.Is(err, errs.ErrNotFoundOrder) {
			return e.JSON(http.StatusNotFound, response.ResponseAPI(false, http.StatusNotFound, "Order not found", nil))
		}
		return e.JSON(http.StatusInternalServerError, response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	for _, result := range results {
		var productImage string
		for _, val := range result.OrderItems {
			productImage = val.ProductImage
		}

		resOrders = append(resOrders, response.OrderAdminList{
			ID:           result.ID,
			OrderCode:    result.OrderCode,
			Status:       result.Status,
			TotalAmount:  result.TotalAmount,
			ProductImage: productImage,
			CustomerName: result.BuyerName,
		})
	}

	return e.JSON(
		http.StatusOK,
		response.ResponseAPIWithPagination(
			true,
			http.StatusOK,
			"success",
			resOrders,
			page,
			totalData,
			totalPage,
			perPage,
		),
	)

}

func NewOrderHandler(g *echo.Group, orderService service.OrderServiceInterface, cfg *config.Config, JwtService service.JwtServiceInterface) OrderHandlerInterface {
	orderHandler := &OrderHandler{orderService: orderService}

	mid := middleware.NewmiddlewareAuth(cfg, JwtService)

	orderAauth := g.Group("/auth", mid.CheckToken(), mid.CheckRole("Customer"))
	orderAauth.POST("/orders", orderHandler.Create)
	orderAauth.GET("/orders", orderHandler.GetAllCustomer)

	orderAdmin := g.Group("/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))
	orderAdmin.GET("/orders", orderHandler.GetAll)
	orderAdmin.GET("/orders/:orderID", orderHandler.GetByID)
	orderAdmin.PUT("/orders/:orderID/status", orderHandler.UpdateStatus)

	return orderHandler
}
