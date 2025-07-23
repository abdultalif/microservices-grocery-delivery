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
	GetAllAdmin(c echo.Context) error
	GetByID(c echo.Context) error
	Delete(c echo.Context) error
}

type productHandler struct {
	service service.ProductServiceInterface
}

// Delete implements ProductHandlerInterface.
func (p *productHandler) Delete(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		ctx        = c.Request().Context()
	)

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		log.Errorf("[ProductHandler-1] Delete: invalid product ID %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "invalid product ID"
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	err = p.service.Delete(ctx, productID)
	if err != nil {
		log.Errorf("[ProductHandler-2] Delete: %v", err)
		if errors.Is(err, errs.ErrProductHasChildren) {
			res.Message = "Product has children, cannot delete"
			res.Data = nil
			res.Code = http.StatusConflict
			res.Success = false
			return c.JSON(http.StatusConflict, res)
		} else if errors.Is(err, errs.ErrProductNotFound) {
			res.Message = "Product not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		} else {
			res.Message = err.Error()
			res.Data = nil
			res.Code = http.StatusInternalServerError
			res.Success = false
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Data = nil
	res.Message = "success"
	return c.JSON(http.StatusOK, res)


}

// GetByID implements ProductHandlerInterface.
func (p *productHandler) GetByID(c echo.Context) error {
	var (
		res        = response.ResponseDefault{}
		ctx        = c.Request().Context()
		resProduct = response.ProductDetailResponse{}
	)

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		log.Errorf("[ProductHandler-1] GetByID: invalid product ID %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "invalid product ID"
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	product, err := p.service.GetByID(ctx, productID)

	if err != nil {
		log.Errorf("[ProductHandler-2] GetByID: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			res.Message = "Product not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	resProduct = response.ProductDetailResponse{
		ID:            product.ID,
		CategorySlug:  product.CategorySlug,
		ParentID:      product.ParentID,
		ProductName:   product.Name,
		RegulerPrice:  int64(product.RegulerPrice),
		SalePrice:     int64(product.SalePrice),
		Unit:          product.Unit,
		Weight:        product.Weight,
		Stock:         product.Stock,
		ProductStatus: product.Status,
		CategoryName:  product.CategoryName,
		CreatedAt:     product.CreatedAt,
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Data = resProduct
	res.Message = "success"
	return c.JSON(http.StatusOK, res)
}

// GetAllAdmin implements ProductHandlerInterface.
func (p *productHandler) GetAllAdmin(c echo.Context) error {
	var (
		res         = response.DefaultResponseWithPaginations{}
		ctx         = c.Request().Context()
		resProducts = []response.ProductListResponse{}
	)

	search := c.QueryParam("search")
	orderBy := "created_at"
	if c.QueryParam("orderBy") != "" {
		orderBy = c.QueryParam("orderBy")
	}

	orderType := "desc"
	if c.QueryParam("orderType") != "" {
		orderType = c.QueryParam("orderType")
	}

	var page int64 = 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var perPage int64 = 10
	if perPageStr := c.QueryParam("limit"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	categorySlug := c.QueryParam("categorySlug")
	startPrice, err := conv.StringToInt64(c.QueryParam("startPrice"))
	if err != nil {
		startPrice = 0
	}

	endPrice, err := conv.StringToInt64(c.QueryParam("endPrice"))
	if err != nil {
		endPrice = 0
	}

	var status = ""
	if c.QueryParam("status") != "" {
		status = c.QueryParam("status")
	}

	reqEntity := entity.QueryStringProduct{
		Search:       search,
		OrderBy:      orderBy,
		OrderType:    orderType,
		Page:         int(page),
		Limit:        int(perPage),
		CategorySlug: categorySlug,
		StartPrice:   startPrice,
		EndPrice:     endPrice,
		Status:       status,
	}

	results, totalData, totalPage, err := p.service.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-1] GetAll: %v", err)
		if err.Error() == "404" {
			res.Message = "Data not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	for _, product := range results {
		resProducts = append(resProducts, response.ProductListResponse{
			ID:            product.ID,
			ProductName:   product.Name,
			ParentID:      product.ParentID,
			ProductImage:  product.Image,
			CategoryName:  product.CategoryName,
			ProductStatus: product.Status,
			SalePrice:     int64(product.SalePrice),
			CreatedAt:     product.CreatedAt,
		})
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Data = resProducts
	res.Message = "success"
	res.Pagination = &response.Pagination{
		Page:       page,
		TotalCount: totalData,
		TotalPage:  totalPage,
		PerPage:    perPage,
	}

	return c.JSON(http.StatusOK, res)

}

func NewProductHandler(g *echo.Group, productService service.ProductServiceInterface, cfg *config.Config, JwtService service.JwtServiceInterface) ProductHandlerInterface {
	product := &productHandler{service: productService}

	// mid := adapter.NewMiddlewareAdapter(cfg, JwtService)
	adminGroup := g.Group("/admin")
	adminGroup.GET("/products", product.GetAllAdmin)
	adminGroup.GET("/products/:productID", product.GetByID)
	adminGroup.DELETE("/products/:productID", product.Delete)

	return product
}
