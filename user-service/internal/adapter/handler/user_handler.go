package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"user-service/config"
	"user-service/internal/adapter"
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

type UserHandlerInterface interface {
	GetProfileUser(c echo.Context) error
	UpdateDataUser(c echo.Context) error
	ChangePassword(c echo.Context) error

	GetCustomerAll(c echo.Context) error
	GetCustomerByID(c echo.Context) error
	CreateCustomer(c echo.Context) error
	UpdateCustomer(c echo.Context) error
	DeleteCustomer(c echo.Context) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

// CreateCustomer implements UserHandlerInterface.
func (u *userHandler) CreateCustomer(c echo.Context) error {
	var (
		res = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
		req  = request.CustomerRequest{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] GetProfileUser: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-2] CreateCustomer: %v", err)
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] CreateCustomer: %v", err)

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
		log.Infof("[UserHandler-4] CreateCustomer: %s", "password and confirm password does not match")
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

	err = u.userService.CreateCustomer(ctx, reqEntity)
	if errors.Is(err, errs.ErrUserExist) {
		log.Errorf("[UserHandler-4] CreateCustomer: %v", err)
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

// DeleteCustomer implements UserHandlerInterface.
func (u *userHandler) DeleteCustomer(c echo.Context) error {
	var (
		res = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] GetProfileUser: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[UserHandler-2] DeleteCustomer: %s", "missing or invalid customer ID")
		res.Message = "missing or invalid customer ID"
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Infof("[UserHandler-3] DeleteCustomer: %s", "invalid customer ID")
		res.Message = "invalid customer ID"
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	err = u.userService.DeleteCustomer(ctx, id)
	if err != nil {
		log.Infof("[UserHandler-4] DeleteCustomer: %v", err)
		if err.Error() == "404" {
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

// GetCustomerAll implements UserHandlerInterface.
func (u *userHandler) GetCustomerAll(c echo.Context) error {
	var (
		res     = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		respUser = []response.CustomerListResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] GetProfileUser: user data not found in context")
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

	results, countData, totalPages, err := u.userService.GetCustomerAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-2] GetCustomerAll: %v", err)
		if err.Error() == "404" {
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

// GetCustomerByID implements UserHandlerInterface.
func (u *userHandler) GetCustomerByID(c echo.Context) error {
	var (
		res     = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		resUser = response.CustomerResponse{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] GetProfileUser: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	idParam := c.Param("id")
	if idParam == "" {
		log.Errorf("[UserHandler-2] GetCustomerByID: %s", "id invalid")
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Message = "id invalid"
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Errorf("[UserHandler-3] GetCustomerByID: %v", err)
		res.Code = http.StatusBadRequest
		res.Success = false
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	result, err := u.userService.GetCustomerByID(ctx, id)
	if err != nil {
		log.Errorf("[UserHandler-4] GetCustomerByID: %v", err)
		if err.Error() == "404" {
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

// UpdateCustomer implements UserHandlerInterface.
func (u *userHandler) UpdateCustomer(c echo.Context) error {
	var (
		res = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
		req  = request.CustomerRequest{}
	)

	_, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] ChangePassword: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		res.Data = nil
		return c.JSON(http.StatusUnauthorized, res)
	}

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-2] UpdateCustomer: %v", err)
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] UpdateCustomer: %v", err)

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
		log.Infof("[UserHandler-4] UpdateCustomer: %s", "missing or invalid customer ID")
		res.Message = "missing or invalid customer ID"
		res.Data = nil
		res.Success = false
		res.Code = http.StatusBadRequest
		return c.JSON(http.StatusBadRequest, res)
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Infof("[UserHandler-5] UpdateCustomer: %s", "invalid customer ID")
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

	err = u.userService.UpdateDataUser(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-6] UpdateCustomer: %v", err)
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

// ChangePassword implements UserHandlerInterface.
func (u *userHandler) ChangePassword(c echo.Context) error {

	var (
		req = request.ChangePasswordRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] ChangePassword: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		return c.JSON(http.StatusUnauthorized, res)
	}

	err = c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-3] ChangePassword: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-4] ChangePassword: %v", err)

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

	if req.NewPassword != req.ConfirmPassword {
		log.Infof("[UserHandler-5] ChangePassword: %s", "new password and confirm password does not match")
		res.Message = "new password and confirm password does not match"
		res.Data = nil
		res.Code = http.StatusUnprocessableEntity
		res.Success = false
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		ID:       user.UserID,
	}

	err = u.userService.ChangePassword(ctx, reqEntity, req.CurrentPassword)
	if err != nil {
		log.Errorf("[UserHandler-6] ChangePassword: %v", err)
		if errors.Is(err, errs.ErrCurrentPasswordIncorrect) {
			res.Success = false
			res.Code = http.StatusUnauthorized
			res.Message = "Current password is incorrect"
			res.Data = nil
			return c.JSON(http.StatusUnauthorized, res)
		}

		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Success = true
	res.Code = http.StatusOK
	res.Message = "success"
	res.Data = nil
	return c.JSON(http.StatusOK, res)

}

// UpdateDataUser implements UserHandlerInterface.
func (u *userHandler) UpdateDataUser(c echo.Context) error {
	var (
		req = request.UpdateDataUserRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] UpdateDataUser: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		return c.JSON(http.StatusUnauthorized, res)
	}

	userID := user.UserID
	if err := c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-3] UpdateDataUser: %v", err)
		res.Message = err.Error()
		res.Code = http.StatusBadRequest
		res.Data = nil
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-4] UpdateDataUser: %v", err)

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

	latString := conv.LatLngToString(req.Lat)
	lngString := conv.LatLngToString(req.Lng)
	phoneString := fmt.Sprintf("%d", req.Phone)

	reqEntity := entity.UserEntity{
		ID:      userID,
		Name:    req.Name,
		Email:   req.Email,
		Lat:     latString,
		Lng:     lngString,
		Address: req.Address,
		Phone:   phoneString,
	}

	err = u.userService.UpdateDataUser(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdateDataUser: %v", err)
		if err.Error() == "404" {
			res.Message = "user not found"
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

	res.Message = "success"
	res.Success = true
	res.Code = http.StatusOK
	res.Data = nil
	return c.JSON(http.StatusOK, res)
}

// GetProfileUser implements UserHandlerInterface.
func (u *userHandler) GetProfileUser(c echo.Context) error {
	var (
		res         = response.ResponseDefault{}
		respProfile = response.ProfileResponse{}
		ctx         = c.Request().Context()
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] GetProfileUser: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		return c.JSON(http.StatusUnauthorized, res)
	}

	userID := user.UserID

	dataUser, err := u.userService.GetProfileUser(ctx, userID)
	if err != nil {
		log.Errorf("[UserHandler-3] GetProfileUser: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Message = "user not found"
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

	respProfile.Address = dataUser.Address
	respProfile.Name = dataUser.Name
	respProfile.Email = dataUser.Email
	respProfile.ID = dataUser.ID
	respProfile.Lat = dataUser.Lat
	respProfile.Lng = dataUser.Lng
	respProfile.Phone = dataUser.Phone
	respProfile.Photo = dataUser.Photo
	respProfile.RoleName = dataUser.RoleName

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "success"
	res.Data = respProfile

	return c.JSON(http.StatusOK, res)
}











var err error



func NewUserHandler(g *echo.Group, userService service.UserServiceInterface, cfg *config.Config, jwtService service.JwtServiceInterface) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}
	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)

	adminGroup := g.Group("/admin", mid.CheckToken(), mid.CheckRole("Super Admin"))
	adminGroup.GET("/customers", userHandler.GetCustomerAll)
	adminGroup.POST("/customers", userHandler.CreateCustomer)
	adminGroup.PATCH("/customers/:id", userHandler.UpdateCustomer)
	adminGroup.GET("/customers/:id", userHandler.GetCustomerByID)
	adminGroup.DELETE("/customers/:id", userHandler.DeleteCustomer)
	
	userGroup := g.Group("/user", mid.CheckToken(), mid.CheckRole("Customer", "Super Admin"))
	userGroup.GET("/profile", userHandler.GetProfileUser)
	userGroup.PATCH("/update-profile", userHandler.UpdateDataUser)
	userGroup.PATCH("/change-password", userHandler.ChangePassword)
	return userHandler
}
