package handler

import (
	"errors"
	"net/http"
	"product-service/config"
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler/request"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/service"
	"product-service/utils/conv"
	v "product-service/utils/validator"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type ProductHandlerInterface interface {
	GetAllAdmin(c echo.Context) error
	GetByID(c echo.Context) error
	Update(c echo.Context) error
	Create(c echo.Context) error
	Delete(c echo.Context) error
}

type productHandler struct {
	service service.ProductServiceInterface
}

// Create implements ProductHandlerInterface.
func (p *productHandler) Create(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		req = request.ProductRequest{}
		ctx = c.Request().Context()
	)

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[ProductHandler-1] Create: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "invalid request"
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[CategoryHandler-3] Create: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			res.Message = ve.Errors
			res.Success = false
			res.Code = http.StatusUnprocessableEntity
			res.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, res)
		}

		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusUnprocessableEntity
		res.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	reqEntity := entity.ProductEntity{
		CategorySlug: req.CategorySlug,
		ParentID:     nil,
		Name:         req.ProductName,
		Image:        req.VariantDetail[0].ProductImage,
		Description:  req.ProductDescription,
		RegulerPrice: float64(req.VariantDetail[0].RegulerPrice),
		SalePrice:    float64(req.VariantDetail[0].SalePrice),
		Unit:         req.Unit,
		Weight:       req.VariantDetail[0].Weight,
		Stock:        req.VariantDetail[0].Stock,
		Variant:      req.Variant,
		Status:       req.Status,
	}

	productChilds := []entity.ProductEntity{}
	if len(req.VariantDetail) > 1{
		for i := 1; i < len(req.VariantDetail); i++ {
			productChilds = append(productChilds, entity.ProductEntity{
				Image:        req.VariantDetail[i].ProductImage,
				RegulerPrice: float64(req.VariantDetail[i].RegulerPrice),
				SalePrice:    float64(req.VariantDetail[i].SalePrice),
				Weight:       req.VariantDetail[i].Weight,
				Stock:        req.VariantDetail[i].Stock,
			})
		}
	}

	reqEntity.Child = productChilds

	err = p.service.Create(ctx, reqEntity)
	if err != nil {
		if errors.Is(err, errs.ErrCategoryNotFound) {
			res.Message = "Category not found"
			res.Data = nil
			res.Code = http.StatusUnprocessableEntity
			res.Success = false
			return c.JSON(http.StatusUnprocessableEntity, res)
		} else if errors.Is(err, errs.ErrProductAlreadyExists) {
			res.Message = "Product already exist"
			res.Data = nil
			res.Code = http.StatusConflict
			res.Success = false
			return c.JSON(http.StatusConflict, res)
		}
		log.Errorf("[ProductHandler-2] Create: %v", err)
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Success = true
	res.Code = http.StatusCreated
	res.Message = "success"
	res.Data = nil
	return c.JSON(http.StatusCreated, res)

}

// Update implements ProductHandlerInterface.
func (p *productHandler) Update(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		req = request.UpdateProductRequest{}
		ctx = c.Request().Context()
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

	if err := c.Bind(&req); err != nil {
		res.Message = "Invalid request"
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if req.VariantDetail == nil || len(*req.VariantDetail) == 0 {
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "Variant detail is required and must contain at least one item"
		return c.JSON(http.StatusBadRequest, res)
	}

	mainVariant := (*req.VariantDetail)[0]
	productImage := mainVariant.ProductImage
	regulerPrice := float64(mainVariant.RegulerPrice)
	salePrice := float64(mainVariant.SalePrice)
	weight := mainVariant.Weight
	stock := mainVariant.Stock

	children := []entity.ProductEntity{}
	if len(*req.VariantDetail) > 1 {
		for i := 1; i < len(*req.VariantDetail); i++ {
			child := (*req.VariantDetail)[i]
			children = append(children, entity.ProductEntity{
				Image:        child.ProductImage,
				RegulerPrice: float64(child.RegulerPrice),
				SalePrice:    float64(child.SalePrice),
				Weight:       child.Weight,
				Stock:        child.Stock,
			})
		}
	}

	reqEntity := entity.ProductEntity{
		ID:           productID,
		CategorySlug: *req.CategorySlug,
		ParentID:     nil,
		Name:         *req.ProductName,
		Image:        productImage,
		Description:  *req.ProductDescription,
		RegulerPrice: regulerPrice,
		SalePrice:    salePrice,
		Unit:         *req.Unit,
		Weight:       weight,
		Stock:        stock,
		Variant:      *req.Variant,
		Status:       *req.Status,
		Child:        children,
	}

	err = p.service.Update(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-Update] %v", err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Message = "Product not found"
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}

		if errors.Is(err, errs.ErrCategoryNotFound) {
			res.Message = "Category not found"
			res.Code = http.StatusUnprocessableEntity
			res.Success = false
			return c.JSON(http.StatusUnprocessableEntity, res)
		}

		res.Message = err.Error()
		res.Code = http.StatusInternalServerError
		res.Success = false
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "Product updated successfully"
	res.Code = http.StatusOK
	res.Success = true
	return c.JSON(http.StatusOK, res)

}

// Delete implements ProductHandlerInterface.
func (p *productHandler) Delete(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		log.Errorf("[ProductHandler-1] Delete: Product id must be uuid")
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "Product id must be uuid"
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
		log.Errorf("[ProductHandler-1] GetByID: Product id must be uuid")
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = "Product id must be uuid"
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
		childResponses := []response.ProductChildResponse{}
		for _, child := range product.Child {
			childResponses = append(childResponses, response.ProductChildResponse{
				ID:           child.ID,
				Weight:       child.Weight,
				Stock:        child.Stock,
				RegulerPrice: int64(child.RegulerPrice),
				SalePrice:    int64(child.SalePrice),
			})
		}

		resProducts = append(resProducts, response.ProductListResponse{
			ID:            product.ID,
			ProductName:   product.Name,
			ParentID:      product.ParentID,
			ProductImage:  product.Image,
			CategoryName:  product.CategoryName,
			ProductStatus: product.Status,
			SalePrice:     int64(product.SalePrice),
			CreatedAt:     product.CreatedAt,
			Child:         childResponses,
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

	mid := adapter.NewMiddlewareAdapter(cfg, JwtService)
	adminGroup := g.Group("/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))
	adminGroup.GET("/products", product.GetAllAdmin)
	adminGroup.POST("/products", product.Create)
	adminGroup.PATCH("/products/:productID", product.Update)
	adminGroup.GET("/products/:productID", product.GetByID)
	adminGroup.DELETE("/products/:productID", product.Delete)

	return product
}
