package handler

import (
	"errors"
	"fmt"
	"net/http"
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
}

type userHandler struct {
	userService service.UserServiceInterface
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

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-2] ChangePassword: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] ChangePassword: %v", err)

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
		log.Infof("[UserHandler-4] ChangePassword: %s", "new password and confirm password does not match")
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
		log.Errorf("[UserHandler-5] ChangePassword: %v", err)
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
		log.Errorf("[UserHandler-2] UpdateDataUser: %v", err)
		res.Message = err.Error()
		res.Code = http.StatusBadRequest
		res.Data = nil
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] UpdateDataUser: %v", err)

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

	err := u.userService.UpdateDataUser(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-4] UpdateDataUser: %v", err)
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
		log.Errorf("[UserHandler-2] GetProfileUser: %v", err)
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

func NewUserHandler(userService service.UserServiceInterface) UserHandlerInterface {
	return &userHandler{userService: userService}
}
