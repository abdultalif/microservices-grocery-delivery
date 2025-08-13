package handler

import (
	"errors"
	"net/http"
	"order-service/config"

	// "order-service/internal/adapter/handler/request"
	"order-service/internal/adapter/handler/response"
	"order-service/internal/adapter/middleware"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/internal/core/service"
	"order-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OrderHandlerInterface interface {
	GetAll(e echo.Context) error
	GetByID(e echo.Context) error
	// Create(e echo.Context) error
}

type OrderHandler struct {
	orderService service.OrderServiceInterface
}

// Create implements OrderHandlerInterface.
// func (o *OrderHandler) Create(e echo.Context) error {
// 	var (
// 		ctx = c.Request().Context()
// 		req = request.CreateOrderRequest{}
// 	)

// }

// GetByID implements OrderHandlerInterface.
func (o *OrderHandler) GetByID(e echo.Context) error {
	var (
		ctx      = e.Request().Context()
		resOrder = response.OrderAdminDetail{}
	)

	user, ok := e.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[OrderHandler-1] GetByID: user data not found in context")
		return e.JSON(
			http.StatusUnauthorized,
			response.ResponseAPI(false, http.StatusUnauthorized, "Unauthorized", nil),
		)
	}

	orderID, err := uuid.Parse(e.Param("orderID"))
	if err != nil {
		log.Errorf("[OrderHandler-1] GetByID: OrderID must be uuid")
		return e.JSON(
			http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "OrderID must be uuid", nil),
		)
	}

	order, err := o.orderService.GetByID(ctx, orderID, user)
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
	resOrder.ShippingFee = order.ShipingFee
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

	user, ok := e.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[OrderHandler-1] GetAll: user data not found in context")
		return e.JSON(http.StatusUnauthorized, response.ResponseAPI(false, http.StatusUnauthorized, "Unauthorized", nil))
	}

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

	results, totalData, totalPage, err := o.orderService.GetAll(ctx, reqEntity, user)
	if err != nil {
		log.Errorf("[OrderHandler-1] GetAllAdmin: %v", err)
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

	return e.JSON(http.StatusOK, response.ResponseAPIWithPagination(true, http.StatusOK, "success", resOrders, page, totalData, totalPage, perPage))

}

func NewOrderHandler(g *echo.Group, orderService service.OrderServiceInterface, cfg *config.Config, JwtService service.JwtServiceInterface) OrderHandlerInterface {
	orderHandler := &OrderHandler{orderService: orderService}

	mid := middleware.NewMiddlewareAdapter(cfg, JwtService)
	orderAauth := g.Group("/auth", mid.CheckToken(), mid.CheckRole("Customer"))
	orderAauth.GET("/orders", orderHandler.GetAll)
	orderAauth.GET("/orders/:orderID", orderHandler.GetByID)
	// orderAauth.GET("/orders", orderHandler.Create)

	return orderHandler
}
