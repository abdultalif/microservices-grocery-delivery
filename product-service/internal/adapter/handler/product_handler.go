package handler

import (
	"errors"
	"net/http"
	"product-service/config"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/service"
	"product-service/utils/conv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type ProductHandlerInterface interface {
	GetAllHome(c echo.Context) error
	GetAllShop(c echo.Context) error
}

type productHandler struct {
	service service.ProductServiceInterface
}

// GetAllShop implements ProductHandlerInterface.
func (p *productHandler) GetAllShop(c echo.Context) error {
	var (
		res       = response.DefaultResponseWithPaginations{}
		ctx       = c.Request().Context()
		respLists = []response.ProductHomeListResponse{}
	)

	orderyBy := "created_at"
	if c.QueryParam("orderBy") != "" {
		orderyBy = c.QueryParam("orderBy")
	}

	orderType := "desc"
	if c.QueryParam("orderType") != "" {
		orderyBy = c.QueryParam("orderType")
	}

	var page int64 = 1
	if c.QueryParam("page") != "" {
		page, _ = conv.StringToInt64(c.QueryParam("page"))
	}

	var perPage int64 = 5
	if c.QueryParam("perPage") != "" {
		perPage, _ = conv.StringToInt64(c.QueryParam("perPage"))
	}

	var startPrice int64 = 0
	var endPrice int64 = 0
	if c.QueryParam("price") != "" {
		price := strings.Split(c.QueryParam("price"), " - ")
		startPrice, _ = conv.StringToInt64(price[0])
		endPrice, _ = conv.StringToInt64(price[1])
	}

	
	reqEntity := entity.QueryStringProduct{
		OrderBy: orderyBy,
		OrderType: orderType,
		Page: int(page),
		Limit: int(perPage),
		StartPrice: startPrice,
		EndPrice: endPrice,
	}
	
	if c.QueryParam("search") != "" {
		reqEntity.Search =  c.QueryParam("search")
	}

	results, totalData, totalPage, err := p.service.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-1] GetAllShop: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			res.Code = http.StatusNotFound
			res.Success = false
			res.Message = "Data not found"
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
	}

	for _, result := range results {
		respLists = append(respLists, response.ProductHomeListResponse{
			ID: result.ID,
			ProductName: result.Name,
			ProductImage: result.Image,
			SalePrice: int64(result.SalePrice),
			RegulerPrice: int64(result.RegulerPrice),
			CategoryName: result.CategoryName,
		})
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "success"
	res.Pagination = &response.Pagination{
		Page: page,
		TotalCount: totalData,
		TotalPage: totalPage,
		PerPage: perPage,
	}
	res.Data = respLists
	return c.JSON(http.StatusOK, res)
}

func (p *productHandler) GetAllHome(c echo.Context) error {
	var (
		res       = response.ResponseDefault{}
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
	homeProduct.GET("/shop", product.GetAllShop)

	return product
}
