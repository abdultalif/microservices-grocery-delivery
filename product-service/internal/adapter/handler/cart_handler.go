package handler

import (
	"errors"
	"net/http"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler/request"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler/response"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/service"
	v "github.com/abdultalif/microservices-grocery-delivery/product-service/utils/validator"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type CartHandlerInterface interface {
	AddToCart(c echo.Context) error
	RemoveFromCart(c echo.Context) error
	GetCart(c echo.Context) error
	RemoveAllCart(c echo.Context) error
}

type CartHandler struct {
	CartService    service.CartServiceInterface
	ProductService service.ProductServiceInterface
}

// RemoveAllCart implements CartHandlerInterface.
func (ch *CartHandler) RemoveAllCart(c echo.Context) error {

	var ctx = c.Request().Context()

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] RemoveAllCart: user data not found in context")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "unauthorized"))
	}

	err := ch.CartService.RemoveAllCart(ctx, user.UserID)
	if err != nil {
		log.Errorf("[CartHandler-2] RemoveAllCart: %v", err)
		if errors.Is(err, errs.ErrCartNotFound) {
			return c.JSON(http.StatusNotFound, response.APIResponseError(http.StatusNotFound, "cart not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.APIResponseSuccess(http.StatusOK, "success", nil))
}

// RemoveFromCart implements CartHandlerInterface.
func (ch *CartHandler) RemoveFromCart(c echo.Context) error {

	var ctx = c.Request().Context()

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] Create: user data not found in context")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "unauthorized"))
	}

	productID, err := uuid.Parse(c.QueryParam("product_id"))
	if err != nil {
		log.Errorf("[CartHandler-1] RemoveFromCart: %v", err)
		return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, "invalid product ID"))
	}

	err = ch.CartService.RemoveFromCart(ctx, user.UserID, productID)
	if err != nil {
		log.Errorf("[CartHandler-2] RemoveFromCart: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, response.APIResponseError(http.StatusNotFound, "product not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.APIResponseSuccess(http.StatusOK, "success", nil))

}

// GetCart implements CartHandlerInterface.
func (ch *CartHandler) GetCart(c echo.Context) error {

	var (
		ctx = c.Request().Context()
		res = []response.CartResponse{}
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] Create: user data not found in context")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "unauthorized"))
	}

	cart, err := ch.CartService.GetCartByUserID(ctx, user.UserID)
	if err != nil {
		log.Errorf("[CartHandler-3] GetCart: %v", err)
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}
	for _, item := range cart {
		product, err := ch.ProductService.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Errorf("[CartHandler-4] GetCart: %v", err)
			return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
		}

		res = append(res, response.CartResponse{
			ID:            item.ProductID,
			ProductName:   product.Name,
			ProductImage:  product.Image,
			ProductStatus: product.Status,
			SalePrice:     int64(product.SalePrice),
			Quantity:      item.Quantity,
			Unit:          product.Unit,
			Weight:        int64(product.Weight),
		})
	}

	return c.JSON(http.StatusOK, response.APIResponseSuccess(http.StatusOK, "success", res))

}

// AddToCart implements CartHandlerInterface.
func (ch *CartHandler) AddToCart(c echo.Context) error {

	var (
		ctx     = c.Request().Context()
		request = request.CartRequest{}
	)

	if err := c.Bind(&request); err != nil {
		log.Errorf("[CartHandler-1] AddToCart: %v", err)
		return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, err.Error()))
	}

	if err := c.Validate(request); err != nil {
		log.Errorf("[CategoryHandler-3] Create: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity, response.APIResponseError(http.StatusUnprocessableEntity, ve.Errors))
		}

		return c.JSON(http.StatusUnprocessableEntity, response.APIResponseError(http.StatusUnprocessableEntity, err.Error()))
	}

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CategoryHandler-1] Create: user data not found in context")
		return c.JSON(http.StatusUnauthorized, response.APIResponseError(http.StatusUnauthorized, "unauthorized"))
	}

	reqEntity := entity.CartItem{
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	}

	cart, err := ch.CartService.AddToCart(ctx, user.UserID, reqEntity)
	if err != nil {
		log.Errorf("[CartHandler-5] AddToCart: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, response.APIResponseError(http.StatusNotFound, "product not found"))
		} else if errors.Is(err, errs.ErrInvalidQuantity) {
			return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, "invalid quantity"))
		} else {
			return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
		}
	}

	return c.JSON(http.StatusCreated, response.APIResponseSuccess(http.StatusCreated, "success", cart))

}

func NewCartHandler(cartService service.CartServiceInterface, productService service.ProductServiceInterface) CartHandlerInterface {
	return &CartHandler{
		CartService:    cartService,
		ProductService: productService,
	}
}
