package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/internal/core/service"
	"user-service/utils/conv"
	v "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type CustomerHandlerInterface interface {
	GetCustomerAll(c echo.Context) error
	GetCustomerByID(c echo.Context) error
	CreateCustomer(c echo.Context) error
	UpdateCustomer(c echo.Context) error
	DeleteCustomer(c echo.Context) error
	UpdateLocationCustomer(c echo.Context) error
}

type CustomerHandler struct {
	customerService service.CustomerServiceInterface
	userService     service.UserServiceInterface
}

// UpdateLocationCustomer implements CustomerHandlerInterface.
func (u *CustomerHandler) UpdateLocationCustomer(c echo.Context) error {
	var (
		ctx = c.Request().Context()
		req = request.CustomerLocationRequest{}
	)

	idParam := c.Param("id")
	customerID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid customer ID", nil))
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CustomerHandler-1] UpdateLocationCustomer: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body format", nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[CustomerHandler-2] UpdateLocationCustomer: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	request := entity.UserEntity{
		ID:  customerID,
		Lat: req.Lat,
		Lng: req.Lng,
	}

	err = u.customerService.UpdateLocationCustomer(ctx, request)
	if err != nil {
		log.Errorf("[CustomerHandler-4] UpdateLocationCustomer: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, err.Error(), nil))
		}
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Success", nil))

}

// CreateCustomer implements CustomerHandlerInterface.
func (u *CustomerHandler) CreateCustomer(c echo.Context) error {
	var (
		res = response.DefaultResponseWithPaginations{}
		ctx = c.Request().Context()
		req = request.CustomerRequest{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CustomerHandler-1] CreateCustomer: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CustomerHandler-2] CreateCustomer: %v", err)
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[CustomerHandler-3] CreateCustomer: %v", err)

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

	if req.Password != req.PasswordConfirmation {
		log.Infof("[CustomerHandler-4] CreateCustomer: %s", "password and confirm password does not match")
		res.Message = "password and confirm password does not match"
		res.Data = nil
		res.Code = http.StatusUnprocessableEntity
		res.Success = false
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	latString := strconv.FormatFloat(req.Lat, 'g', -1, 64)
	lngString := strconv.FormatFloat(req.Lng, 'g', -1, 64)

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		Address:  req.Address,
		Lat:      latString,
		Lng:      lngString,
		Photo:    req.Photo,
		RoleID:   req.RoleID,
	}

	err := u.customerService.CreateCustomer(ctx, reqEntity)
	if errors.Is(err, errs.ErrUserExist) {
		log.Errorf("[CustomerHandler-4] CreateCustomer: %v", err)
		res.Message = "Email already exists"
		res.Success = false
		res.Code = http.StatusConflict
		res.Data = nil
		return c.JSON(http.StatusConflict, res)
	}

	if err != nil {
		res.Message = "failed to create customer"
		res.Data = nil
		res.Code = http.StatusInternalServerError
		res.Success = false
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "success"
	res.Success = true
	res.Code = http.StatusCreated
	res.Data = nil
	res.Pagination = nil
	return c.JSON(http.StatusCreated, res)
}

// DeleteCustomer implements CustomerHandlerInterface.
func (u *CustomerHandler) DeleteCustomer(c echo.Context) error {
	var (
		res = response.DefaultResponseWithPaginations{}
		ctx = c.Request().Context()
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CustomerHandler-1] DeleteCustomer: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[CustomerHandler-2] DeleteCustomer: %s", "missing or invalid customer ID")
		res.Message = "missing or invalid customer ID"
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Infof("[CustomerHandler-3] DeleteCustomer: %s", "invalid customer ID")
		res.Message = "invalid customer ID"
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	err = u.customerService.DeleteCustomer(ctx, id)
	if err != nil {
		log.Infof("[CustomerHandler-4] DeleteCustomer: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Message = "Customer not found"
			res.Success = false
			res.Code = http.StatusNotFound
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "Customer deleted successfully"
	res.Data = nil
	res.Success = true
	res.Code = http.StatusOK
	return c.JSON(http.StatusOK, res)
}

// GetCustomerAll implements CustomerHandlerInterface.
func (u *CustomerHandler) GetCustomerAll(c echo.Context) error {
	var (
		res      = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		respUser = []response.CustomerListResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CustomerHandler-1] GetCustomerAll: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	search := c.QueryParam("search")
	orderBy := "created_at"
	if c.QueryParam("order_by") != "" {
		orderBy = c.QueryParam("order_by")
	}

	orderType := c.QueryParam("order_type")
	if orderType != "asc" && orderType != "desc" {
		orderType = "desc"
	}

	pageStr := c.QueryParam("page")
	var page int64 = 1
	if pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	limitStr := c.QueryParam("limit")
	var limit int64 = 10
	if limitStr != "" {
		limit, _ = conv.StringToInt64(limitStr)
		if limit <= 0 {
			limit = 10
		}
	}

	reqEntity := entity.QueryStringCustomer{
		Search:    search,
		Page:      page,
		Limit:     limit,
		OrderBy:   orderBy,
		OrderType: orderType,
	}

	results, countData, totalPages, err := u.customerService.GetCustomerAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[CustomerHandler-2] GetCustomerAll: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Success = false
			res.Code = http.StatusNotFound
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

	for _, val := range results {
		respUser = append(respUser, response.CustomerListResponse{
			ID:    val.ID,
			Name:  val.Name,
			Email: val.Email,
			Photo: val.Photo,
			Phone: val.Phone,
		})
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "Data retrieved successfully"
	res.Data = respUser
	res.Pagination = &response.Pagination{
		Page:       page,
		TotalCount: countData,
		PerPage:    limit,
		TotalPage:  totalPages,
	}

	return c.JSON(http.StatusOK, res)
}

// GetCustomerByID implements CustomerHandlerInterface.
func (u *CustomerHandler) GetCustomerByID(c echo.Context) error {
	var (
		res     = response.DefaultResponseWithPaginations{}
		ctx     = c.Request().Context()
		resUser = response.CustomerResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CustomerHandler-1] GetCustomerByID: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idParam := c.Param("id")
	if idParam == "" {
		log.Errorf("[CustomerHandler-2] GetCustomerByID: %s", "id invalid")
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Message = "id invalid"
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Errorf("[CustomerHandler-3] GetCustomerByID: %v", err)
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	result, err := u.customerService.GetCustomerByID(ctx, id)
	if err != nil {
		log.Errorf("[CustomerHandler-4] GetCustomerByID: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Success = false
			res.Code = http.StatusNotFound
			res.Message = "Customer not found"
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	resUser.ID = result.ID
	resUser.Name = result.Name
	resUser.Email = result.Email
	resUser.Phone = result.Phone
	resUser.Address = result.Address
	resUser.Photo = result.Photo
	resUser.Lat = result.Lat
	resUser.Lng = result.Lng
	resUser.RoleID = result.RoleID
	resUser.RoleName = result.RoleName

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "success get customer by id"
	res.Data = resUser
	res.Pagination = nil

	return c.JSON(http.StatusOK, res)
}

// UpdateCustomer implements CustomerHandlerInterface.
func (u *CustomerHandler) UpdateCustomer(c echo.Context) error {
	var (
		res = response.DefaultResponseWithPaginations{}
		ctx = c.Request().Context()
		req = request.CustomerRequest{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[CustomerHandler-1] UpdateCustomer: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CustomerHandler-2] UpdateCustomer: %v", err)
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[CustomerHandler-3] UpdateCustomer: %v", err)

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

	latString := ""
	lngString := ""
	if req.Lat != 0 {
		latString = strconv.FormatFloat(req.Lat, 'g', -1, 64)
	}

	if req.Lng != 0 {
		lngString = strconv.FormatFloat(req.Lng, 'g', -1, 64)
	}
	phoneString := fmt.Sprintf("%v", req.Phone)

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[CustomerHandler-4] UpdateCustomer: %s", "missing or invalid customer ID")
		res.Message = "missing or invalid customer ID"
		res.Data = nil
		res.Success = false
		res.Code = http.StatusBadRequest
		return c.JSON(http.StatusBadRequest, res)
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Infof("[CustomerHandler-5] UpdateCustomer: %s", "invalid customer ID")
		res.Message = "invalid customer ID"
		res.Data = nil
		res.Success = false
		res.Code = http.StatusBadRequest
		return c.JSON(http.StatusBadRequest, res)
	}

	reqEntity := entity.UserEntity{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    phoneString,
		Address:  req.Address,
		Lat:      latString,
		Lng:      lngString,
		Photo:    req.Photo,
	}

	err = u.customerService.UpdateCustomer(ctx, reqEntity)
	if err != nil {
		log.Errorf("[CustomerHandler-6] UpdateCustomer: %v", err)
		if err.Error() == "404" {
			res.Message = "Customer not found"
			res.Data = nil
			res.Success = false
			res.Code = http.StatusNotFound
			return c.JSON(http.StatusNotFound, res)
		}
		res.Message = err.Error()
		res.Data = nil
		res.Success = false
		res.Code = http.StatusInternalServerError
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "Success"
	res.Success = true
	res.Code = http.StatusOK
	res.Data = nil

	return c.JSON(http.StatusOK, res)
}

func NewCustomerHandler(customerService service.CustomerServiceInterface, userService service.UserServiceInterface) CustomerHandlerInterface {
	return &CustomerHandler{
		customerService: customerService,
		userService:     userService,
	}
}
