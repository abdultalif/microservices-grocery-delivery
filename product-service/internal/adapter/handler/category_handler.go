package handler

import (
	"net/http"
	"product-service/config"
	"product-service/internal/adapter"
	"product-service/internal/adapter/handler/request"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/service"
	"product-service/utils/conv"
	v "product-service/utils/validator"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type CategoryHandlerInterface interface {
	GetAll(ctx echo.Context) error
	GetByID(ctx echo.Context) error
	GetBySlug(ctx echo.Context) error
	Create(ctx echo.Context) error
	Delete(ctx echo.Context) error
	Update(ctx echo.Context) error
}

type CategoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// Update implements CategoryHandlerInterface.
func (ct *CategoryHandler) Update(c echo.Context) error {
	var (
		req = request.UpdateCategoryRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] Update: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idStr := c.Param("categoryId")
	if idStr == "" {
		res.Message = "ID is required"
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		res.Message = err.Error()
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Bind(&req); err != nil {
		res.Message = "Invalid request body"
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	updateCategoryEntity := entity.UpdateCategoryEntity{
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		ParentID:    req.ParentID,
		Status:      req.Status,
	}


	err = ct.categoryService.Update(ctx, categoryID, updateCategoryEntity)
	if err != nil {
		log.Errorf("[CategoryHandler-2] Update: %v", err)
		if err.Error() == "404" {
			res.Message = "Data not found"
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		res.Message = err.Error()
		res.Code = http.StatusInternalServerError
		res.Success = false
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "success"
	res.Code = http.StatusOK
	res.Success = true
	return c.JSON(http.StatusOK, res)

}

// GetBySlug implements CategoryHandlerInterface.
func (ct *CategoryHandler) GetBySlug(c echo.Context) error {
	var (
		res           = response.ResponseDefault{}
		ctx           = c.Request().Context()
		resCategories = response.CategoryDetailResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] GetBySlug: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	slug := c.Param("slug")
	if slug == "" {
		res.Message = "Slug is required"
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	result, err := ct.categoryService.GetBySlug(ctx, slug)
	if err != nil {
		log.Errorf("[CategoryHandler-2] GetBySlug: %v", err)
		if err.Error() == "404" {
			res.Message = "Data not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusInternalServerError
		res.Success = false
		return c.JSON(http.StatusInternalServerError, res)
	}

	resCategories = response.CategoryDetailResponse{
		ID:          result.ID,
		Name:        result.Name,
		Icon:        result.Icon,
		Slug:        result.Slug,
		Status:      result.Status,
		Description: result.Description,
	}

	res.Data = resCategories
	res.Message = "success"
	res.Code = http.StatusOK
	res.Success = true
	return c.JSON(http.StatusOK, res)
}

// Create implements CategoryHandlerInterface.
func (ct *CategoryHandler) Create(c echo.Context) error {
	var (
		req = request.CreateCategoryRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] Create: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CategoryHandler-2] Create: %v", err)
		res.Message = "Invalid request body"
		res.Code = http.StatusBadRequest
		res.Success = false
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

	categoryEntity := entity.CategoryEntity{
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		ParentID:    &req.ParentID,
		Status:      req.Status,
	}

	if err := ct.categoryService.Create(ctx, categoryEntity); err != nil {
		log.Errorf("[CategoryHandler-5] Create: %v", err)
		if err.Error() == "409" {
			res.Message = "Category with slug " + categoryEntity.Slug + " already exists"
			res.Data = nil
			res.Code = http.StatusConflict
			res.Success = false
			return c.JSON(http.StatusConflict, res)
		} else if err.Error() == "400" {
			res.Message = "Parent category not found"
			res.Data = nil
			res.Code = http.StatusBadRequest
			res.Success = false
			return c.JSON(http.StatusBadRequest, res)
		} else if err.Error() == "404" {
			res.Message = err.Error()
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
		} else {
			res.Message = err.Error()
			res.Success = false
			res.Code = http.StatusInternalServerError
			res.Data = nil
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	res.Message = "success"
	res.Code = http.StatusCreated
	res.Success = true
	res.Data = nil
	return c.JSON(http.StatusCreated, res)
}

// Delete implements CategoryHandlerInterface.
func (ct *CategoryHandler) Delete(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] Delete: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idStr := c.Param("categoryId")
	if idStr == "" {
		log.Errorf("[CategoryHandler-2] Delete: %v", "invalid id")
		res.Message = "ID is required"
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		log.Errorf("[CategoryHandler-3] Delete: %v", err.Error())
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	err = ct.categoryService.Delete(ctx, categoryID)
	if err != nil {
		log.Errorf("[CategoryHandler-4] Delete: %v", err)
		if err.Error() == "404" {
			res.Message = "Data not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		} else if err.Error() == "304" {
			res.Message = "Category has products, cannot delete"
			res.Data = nil
			res.Code = http.StatusNotModified
			res.Success = false
			return c.JSON(http.StatusNotModified, res)
		} else if err.Error() == "409" {
			res.Message = "Category has children, cannot delete"
			res.Data = nil
			res.Code = http.StatusConflict
			res.Success = false
			return c.JSON(http.StatusConflict, res) 
		} else {
			res.Message = err.Error()
			res.Data = nil
			res.Code = http.StatusInternalServerError
			res.Success = false
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	res.Message = "success"
	res.Code = http.StatusOK
	res.Success = true
	return c.JSON(http.StatusOK, res)
}

// GetAll implements CategoryHandlerInterface.
func (ct *CategoryHandler) GetAll(c echo.Context) error {
	var (
		res           = response.DefaultResponseWithPaginations{}
		ctx           = c.Request().Context()
		resCategories = []response.CategoryResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] GetAll: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

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
	if perPageStr := c.QueryParam("perPage"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	reqEntity := entity.QueryStringEntity{
		Search:    search,
		OrderBy:   orderBy,
		OrderType: orderType,
		Page:      page,
		Limit:     perPage,
	}

	results, totalData, totalPage, err := ct.categoryService.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[CategoryHandler-2] GetAll: %v", err)
		if err.Error() == "404" {
			res.Message = "Data not found"
			res.Code = http.StatusNotFound
			res.Success = false
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Message = err.Error()
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	for _, result := range results {
		resCategories = append(resCategories, response.CategoryResponse{
			ID:           result.ID,
			Name:         result.Name,
			Icon:         result.Icon,
			Slug:         result.Slug,
			Status:       result.Status,
			TotalProduct: len(result.Products),
		})
	}

	pagination := response.Pagination{
		Page:       page,
		TotalCount: totalData,
		PerPage:    perPage,
		TotalPage:  totalPage,
	}
	res.Message = "success"
	res.Code = http.StatusOK
	res.Success = true
	res.Data = resCategories
	res.Pagination = &pagination
	return c.JSON(http.StatusOK, res)
}

// GetByID implements CategoryHandlerInterface.
func (ct *CategoryHandler) GetByID(c echo.Context) error {

	var (
		res           = response.ResponseDefault{}
		ctx           = c.Request().Context()
		resCategories = response.CategoryDetailResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] GetByIDAdmin: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idStr := c.Param("categoryId")
	if idStr == "" {
		log.Errorf("[CategoryHandler-2] GetByIDAdmin: %v", "invalid id")
		res.Message = "ID is required"
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		log.Errorf("[CategoryHandler-3] GetByIDAdmin: %v", err.Error())
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	result, err := ct.categoryService.GetByID(ctx, categoryID)
	if err != nil {
		log.Errorf("[CategoryHandler-4] GetByIDAdmin: %v", err)
		if err.Error() == "404" {
			res.Message = "Data not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	resCategories = response.CategoryDetailResponse{
		ID:          result.ID,
		Name:        result.Name,
		Slug:        result.Slug,
		Icon:        result.Icon,
		Status:      result.Status,
		Description: result.Description,
	}

	res.Message = "success"
	res.Code = http.StatusOK
	res.Success = true
	res.Data = resCategories
	return c.JSON(http.StatusOK, res)
}

func NewCategoryHandler(g *echo.Group, categoryService service.CategoryServiceInterface, cfg *config.Config, JwtService service.JwtServiceInterface) CategoryHandlerInterface {
	categoryHandler := &CategoryHandler{categoryService: categoryService}

	
	mid := adapter.NewMiddlewareAdapter(cfg, JwtService)
	adminGroup := g.Group("/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))
	adminGroup.PATCH("/categories/:categoryId", categoryHandler.Update)
	adminGroup.GET("/categories", categoryHandler.GetAll)
	adminGroup.GET("/categories/:slug/slug", categoryHandler.GetBySlug)
	adminGroup.POST("/categories", categoryHandler.Create)
	adminGroup.GET("/categories/:categoryId", categoryHandler.GetByID)
	adminGroup.DELETE("/categories/:categoryId", categoryHandler.Delete)

	return categoryHandler
}
