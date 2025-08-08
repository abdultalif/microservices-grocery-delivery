package handler

import (
	"errors"
	"net/http"
	"product-service/config"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/internal/core/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type CategoryHandlerInterface interface {
	GetAllHome(c echo.Context) error
	GetAllShop(c echo.Context) error
}

type categoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// GetAllHome implements CategoryHandlerInterface.
func (ch *categoryHandler) GetAllShop(c echo.Context) error {
	var (
		res           = response.ResponseDefault{}
		ctx            = c.Request().Context()
		respCategories = []response.CategoryListShopResponse{}
	)

	results, err := ch.categoryService.GetAllPublished(ctx)
	if err != nil {
		log.Errorf("[CategoryHandler-1] GetAllShop: %v", err)
		if errors.Is(err, errs.ErrCategoryNotFound) {
			res.Code = http.StatusNotFound
			res.Data = nil
			res.Message = "Data not found"
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	respCategories = RekursifCategory(results, nil, 0)

	res.Success = true
	res.Code = http.StatusOK
	res.Message = "success"
	res.Data = respCategories
	return c.JSON(http.StatusOK, res)
}

// GetAllHome implements CategoryHandlerInterface.
func (ch *categoryHandler) GetAllHome(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
		resCategories = []response.CategoryListHomeResponse{}
	)

	results, err := ch.categoryService.GetAllPublished(ctx)
	if err != nil {
		log.Errorf("[CategoryHandler-1] GetAllHome: %v", err)
		if errors.Is(err, errs.ErrCategoryNotFound) {
			res.Message = "Data not found"
			res.Success = false
			res.Code = http.StatusNotFound
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	for _, result := range results {
		if result.ParentID == nil {
			resCategories = append(resCategories, response.CategoryListHomeResponse{
				Name: result.Name,
				Icon: result.Icon,
				Slug: result.Slug,
			})
		}
	}


	res.Message = "success"
	res.Code = http.StatusOK
	res.Success = false
	res.Data = resCategories
	return c.JSON(http.StatusOK, res)
		
	
}

func RekursifCategory(results []entity.CategoryEntity, parentID *uuid.UUID, level int) []response.CategoryListShopResponse {
	var resps []response.CategoryListShopResponse

	for _, category := range results {
		if (category.ParentID == nil && parentID == nil) ||
			(category.ParentID != nil && parentID != nil && *category.ParentID == *parentID) {

			childCategories := RekursifCategory(results, &category.ID, level+1)
			resps = append(resps, response.CategoryListShopResponse{
				Name:  category.Name,
				Slug:  category.Slug,
				Child: childCategories,
			})
		}
	}

	return resps
}


func NewCategoryHandler(e *echo.Group, categoryService service.CategoryServiceInterface, cfg *config.Config) CategoryHandlerInterface {
	category := &categoryHandler{categoryService: categoryService}

	categoryApp := e.Group("/categories")
	categoryApp.GET("/home", category.GetAllHome)
	categoryApp.GET("/shop", category.GetAllShop)

	return category
}
