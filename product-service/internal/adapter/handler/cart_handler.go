package handler

import (
	"net/http"
	"product-service/internal/adapter/handler/request"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/service"
	v "product-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type CartHandlerInterface interface {
	AddToCart(c echo.Context) error
	GetCart(c echo.Context) error
}

type CartHandler struct {
	CartService    service.CartServiceInterface
	ProductService service.ProductServiceInterface
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

	if request.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, response.APIResponseError(http.StatusBadRequest, "quantity must be greater than 0"))
	}

	reqEntity := entity.CartItem{
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	}

	cart, err := ch.CartService.AddToCart(ctx, user.UserID, reqEntity)
	if err != nil {
		log.Errorf("[CartHandler-5] AddToCart: %v", err)
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.APIResponseSuccess(http.StatusCreated, "success", cart))

}

func NewCartHandler(cartService service.CartServiceInterface, productService service.ProductServiceInterface) CartHandlerInterface {
	return &CartHandler{
		CartService:    cartService,
		ProductService: productService,
	}
}
