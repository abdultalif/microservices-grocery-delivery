package handler

import (
	"errors"
	"net/http"
	"order-service/config"
	"order-service/internal/adapter/handler/response"
	"order-service/internal/adapter/middleware"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/internal/core/service"
	"order-service/utils/conv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OrderHandlerInterface interface {
	GetAll(e echo.Context) error
}

type OrderHandler struct {
	orderService service.OrderServiceInterface
}

// GetAll implements OrderHandlerInterface.
func (o *OrderHandler) GetAll(e echo.Context) error {
	var (
		ctx        = e.Request().Context()
		resOrders = []response.OrderAdminList{}
	)

	user, ok := e.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] GetAll: user data not found in context")
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

	return orderHandler
}
