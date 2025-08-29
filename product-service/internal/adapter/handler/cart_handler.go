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
}

type CartHandler struct {
	CartService    service.CartServiceInterface
	ProductService service.ProductServiceInterface
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

	err := ch.CartService.AddToCart(ctx, user.UserID, reqEntity)
	if err != nil {
		log.Errorf("[CartHandler-5] AddToCart: %v", err)
		return c.JSON(http.StatusInternalServerError, response.APIResponseError(http.StatusInternalServerError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.APIResponseSuccess(http.StatusCreated, "success", nil))

}

func NewCartHandler(cartService service.CartServiceInterface, productService service.ProductServiceInterface) CartHandlerInterface {
	return &CartHandler{
		CartService:    cartService,
		ProductService: productService,
	}
}
