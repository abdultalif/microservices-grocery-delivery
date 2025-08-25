package handler

import (
	"errors"
	"net/http"
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

	GetAllHome(c echo.Context) error
	GetAllShop(c echo.Context) error
	GetDetailHome(c echo.Context) error
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

	productChilds := []entity.ProductChildEntity{}
	if len(req.VariantDetail) > 1 {
		for i := 1; i < len(req.VariantDetail); i++ {
			productChilds = append(productChilds, entity.ProductChildEntity{
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

	children := []entity.ProductChildEntity{}
	if len(*req.VariantDetail) > 1 {
		for i := 1; i < len(*req.VariantDetail); i++ {
			child := (*req.VariantDetail)[i]
			children = append(children, entity.ProductChildEntity{
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

	resProduct = response.ProductDetailResponse{
		ID:                 product.ID,
		CategorySlug:       product.CategorySlug,
		ParentID:           product.ParentID,
		ProductName:        product.Name,
		ProductImage:       product.Image,
		ProductDescription: product.Description,
		RegulerPrice:       int64(product.RegulerPrice),
		SalePrice:          int64(product.SalePrice),
		Unit:               product.Unit,
		Weight:             product.Weight,
		Stock:              product.Stock,
		ProductStatus:      product.Status,
		CategoryName:       product.CategoryName,
		CreatedAt:          product.CreatedAt,
		Child:              childResponses,
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
		if errors.Is(err, errs.ErrProductNotFound) {
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
	resDetail.Image = result.Image
	resDetail.SalePrice = int64(result.SalePrice)
	resDetail.RegulerPrice = int64(result.RegulerPrice)
	resDetail.Stock = result.Stock
	resDetail.Weight = result.Weight

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

func NewProductHandler(productService service.ProductServiceInterface) ProductHandlerInterface {
	return &productHandler{service: productService}
}
