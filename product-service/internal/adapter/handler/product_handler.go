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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type ProductHandlerInterface interface {
	GetAllHome(c echo.Context) error
	GetAllShop(c echo.Context) error
	GetDetailHome(c echo.Context) error
}

type productHandler struct {
	service service.ProductServiceInterface
}

// GetDetailHome implements ProductHandlerInterface.
func (p *productHandler) GetDetailHome(c echo.Context) error {

	var (
		res       = response.ResponseDefault{}
		ctx       = c.Request().Context()
		resDetail = response.ProductHomeDetailResponse{}
	)

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		log.Errorf("[ProductHandler-1] Update: Product id must be uuid")
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "Product id must be uuid"
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	result, err := p.service.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductHandler-2] GetDetailHome: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			res.Success = false
			res.Code = http.StatusNotFound
			res.Message = "Product not found"
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	resDetail.ID = result.ID
	resDetail.ProductName = result.Name
	resDetail.CategoryName = result.CategoryName
	resDetail.Description = result.Description
	resDetail.Unit = result.Unit

	for _, child := range result.Child {
		resDetail.Child = append(resDetail.Child, response.ProductChildHomeResponse{
			ID:           child.ID,
			Weight:       child.Weight,
			Stock:        child.Stock,
			RegulerPrice: int64(child.RegulerPrice),
			SalePrice:    int64(child.SalePrice),
			Image:        child.Image,
		})
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "Success"
	res.Data = resDetail

	return c.JSON(http.StatusOK, res)
}

// GetAllShop implements ProductHandlerInterface.
func (p *productHandler) GetAllShop(c echo.Context) error {
	var (
		res       = response.DefaultResponseWithPaginations{}
		ctx       = c.Request().Context()
		respLists = []response.ProductHomeListResponse{}
	)

	orderBy := "created_at"
	if c.QueryParam("order_by") != "" {
		orderBy = c.QueryParam("order_by")
	}

	orderType := "desc"
	if c.QueryParam("order_type") != "" {
		orderType = c.QueryParam("order_type")
	}

	var page int64 = 1
	if c.QueryParam("page") != "" {
		page, _ = conv.StringToInt64(c.QueryParam("page"))
	}

	// Fix: Gunakan per_page, bukan limit dari query param
	var perPage int64 = 10 // Default value yang lebih masuk akal
	if c.QueryParam("per_page") != "" {
		perPage, _ = conv.StringToInt64(c.QueryParam("per_page"))
	}
	// Jika ada limit di query param, gunakan itu juga
	if c.QueryParam("limit") != "" {
		perPage, _ = conv.StringToInt64(c.QueryParam("limit"))
	}

	var startPrice int64 = 0
	var endPrice int64 = 0

	if c.QueryParam("start_price") != "" {
		startPrice, _ = conv.StringToInt64(c.QueryParam("start_price"))
	}

	if c.QueryParam("end_price") != "" {
		endPrice, _ = conv.StringToInt64(c.QueryParam("end_price"))
	}

	reqEntity := entity.QueryStringProduct{
		OrderBy:      orderBy,
		OrderType:    orderType,
		Page:         int(page),
		Limit:        int(perPage),
		StartPrice:   startPrice,
		EndPrice:     endPrice,
		CategorySlug: c.QueryParam("category_slug"),
	}

	if c.QueryParam("search") != "" {
		reqEntity.Search = c.QueryParam("search")
	}

	results, totalData, totalPage, err := p.service.SearchProducts(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-1] GetAllShop: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			res.Code = http.StatusNotFound
			res.Success = false
			res.Message = "Data not found"
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}

		// Fix: Tambahkan handling untuk error lain
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Message = "Internal server error"
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
	res.Pagination = &response.Pagination{
		Page:       page,
		TotalCount: totalData,
		TotalPage:  totalPage,
		PerPage:    perPage,
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

	results, _, _, err := p.service.GetAllHome(ctx, reqEntity)
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
	homeProduct.GET("/shop", product.GetAllShop)
	homeProduct.GET("/home", product.GetAllHome)
	homeProduct.GET("/home/:productID", product.GetDetailHome)

	return product
}
