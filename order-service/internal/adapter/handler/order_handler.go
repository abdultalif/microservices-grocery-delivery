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
	v "order-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OrderHandlerInterface interface {
	Create(e echo.Context) error
}

type OrderHandler struct {
	orderService service.OrderServiceInterface
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


func NewOrderHandler(g *echo.Group, orderService service.OrderServiceInterface, cfg *config.Config, JwtService service.JwtServiceInterface) OrderHandlerInterface {
	orderHandler := &OrderHandler{orderService: orderService}

	mid := middleware.NewmiddlewareAuth(cfg, JwtService)
	orderAauth := g.Group("/auth", mid.CheckToken(), mid.CheckRole("Customer"))
	orderAauth.POST("/orders", orderHandler.Create)

	return orderHandler
}
