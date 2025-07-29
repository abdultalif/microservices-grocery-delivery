package handler

import (
	"errors"
	"net/http"
	"product-service/config"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type ProductHandlerInterface interface {
	GetAllHome(c echo.Context) error
}

type productHandler struct {
	service service.ProductServiceInterface
}

func (p *productHandler) GetAllHome(c echo.Context) error {
	var (
		res      = response.ResponseDefault{}
		ctx       = c.Request().Context()
		respLists = []response.ProductHomeListResponse{}
	)

	orderBy := "created_at"
	orderType := "desc"
	var page int64 = 1
	var perPage int64 = 5

	reqEntity := entity.QueryStringProduct{
		OrderBy:   orderBy,
		OrderType: orderType,
		Page:      int(page),
		Limit:     int(perPage),
	}

	results, _, _, err := p.service.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-1] GetAllHome: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			res.Code = http.StatusNotFound
			res.Success = false
			res.Message = "Data not found"
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}

		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	for _, result := range results {
		respLists = append(respLists, response.ProductHomeListResponse{
			ID:           result.ID,
			ProductName:  result.Name,
			ProductImage: result.Image,
			SalePrice:    int64(result.SalePrice),
			RegulerPrice: int64(result.RegulerPrice),
			CategoryName: result.CategoryName,
		})
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "success"
	res.Data = respLists
	return c.JSON(http.StatusOK, res)
}

func NewProductHandler(e *echo.Group, cfg *config.Config, productService service.ProductServiceInterface) ProductHandlerInterface {
	product := &productHandler{service: productService}

	homeProduct := e.Group("/products")
	homeProduct.GET("/home", product.GetAllHome)

	return product
}