package handler

import (
	"net/http"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/service"
	"product-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type CategoryHandlerInterface interface {
	GetAll(ctx echo.Context) error
	GetByID(ctx echo.Context) error
}

type CategoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// GetAll implements CategoryHandlerInterface.
func (ct *CategoryHandler) GetAll(c echo.Context) error {
	var (
		res          = response.DefaultResponseWithPaginations{}
		ctx            = c.Request().Context()
		resCategories = []response.CategoryResponse{}
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
		log.Errorf("[CategoryHandler-1] Create: %v", err)
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
		ctx            = c.Request().Context()
		resCategories = response.CategoryDetailResponse{}
	)

	idStr := c.Param("categoryId")
	if idStr == "" {
		log.Errorf("[CategoryHandler-1] GetByIDAdmin: %v", "invalid id")
		res.Message = "ID is required"
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		log.Errorf("[CategoryHandler-2] GetByIDAdmin: %v", err.Error())
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	result, err := ct.categoryService.GetByID(ctx, categoryID)
	if err != nil {
		log.Errorf("[CategoryHandler-3] GetByIDAdmin: %v", err)
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

func NewCategoryHandler(g *echo.Group, categoryService service.CategoryServiceInterface) CategoryHandlerInterface {
	categoryHandler := &CategoryHandler{categoryService: categoryService}


	// mid := adapter.NewMiddlewareAdapter(cfg)
	// g.Use(mid.CheckToken())
	// adminGroup := g.Group("/admin", mid.CheckToken())
	adminGroup := g.Group("/admin")
	adminGroup.GET("/categories", categoryHandler.GetAll)
	adminGroup.GET("/categories/:categoryId", categoryHandler.GetByID)
	// adminGroup.GET("/current", func(c echo.Context) error {
	// return c.JSON(http.StatusOK, c.Get("user"))
	// })
	return categoryHandler
}
