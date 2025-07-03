package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/utils/conv"
	v "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type UserHandlerInterface interface {
	SignIn(ctx echo.Context) error
	ForgotPassword(ctx echo.Context) error
	UpdatePassword(c echo.Context) error
	ValidateForgotPasswordToken(ctx echo.Context) error
	CreateUserAccount(c echo.Context) error
	VerifyAccount(c echo.Context) error
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
		req         = request.ChangePasswordRequest{}
		res         = response.ResponseDefault{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] ChangePassword: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] ChangePassword: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
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
		ID:    jwtUserData.UserID,
	}

	err = u.userService.ChangePassword(ctx, reqEntity, req.CurrentPassword)
	if err != nil {
		log.Errorf("[UserHandler-6] ChangePassword: %v", err)
		if err.Error() == "401" {
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
		req         = request.UpdateDataUserRequest{}
		res         = response.ResponseDefault{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] UpdateDataUser: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] UpdateDataUser: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	userID := jwtUserData.UserID

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
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetProfileUser: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] GetProfileUser: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	userID := jwtUserData.UserID

	dataUser, err := u.userService.GetProfileUser(ctx, userID)
	if err != nil {
		log.Errorf("[UserHandler-3] GetProfileUser: %v", err)
		if err.Error() == "404" {
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

// ValidateForgotPasswordToken implements UserHandlerInterface.
func (u *userHandler) ValidateForgotPasswordToken(c echo.Context) error {
	var res = response.ResponseDefault{}
	ctx := c.Request().Context()

	token := c.QueryParam("token")
	if token == "" {
		res.Code = http.StatusBadRequest
		res.Message = "Token is required"
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	err := u.userService.ValidateForgotPasswordToken(ctx, token)
	log.Infof("[UserHandler-1] ValidateForgotPasswordToken: %s", err)
	if err != nil {
		if err.Error() == "401" {
			res.Code = http.StatusUnauthorized
			res.Message = "Token is invalid or expired"
			res.Success = false
			res.Data = nil
			return c.JSON(http.StatusUnauthorized, res)
		}
		res.Code = http.StatusInternalServerError
		res.Message = "Internal server error"
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Code = http.StatusOK
	res.Message = "Token is valid"
	res.Success = true
	res.Data = nil
	return c.JSON(http.StatusOK, res)
}

// UpdatePassword implements UserHandlerInterface.
func (u *userHandler) UpdatePassword(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		req = request.UpdatePasswordRequest{}
		ctx = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	log.Infof("[UserHandler-1] UpdatePassword: %s", "Token is required")
	if tokenString == "" {
		res.Code = http.StatusBadRequest
		res.Message = "Token is required"
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Bind(&req); err != nil {
		log.Infof("[UserHandler-2] UpdatePassword: %v", err)
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Success = false
		return c.JSON(http.StatusBadRequest, res)
	}
	// cara lihat request kita
	log.Infof("[UserHandler-5] Hasil request: %+v", req)

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] ForgotPassword: %v", err)

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
		log.Infof("[UserHandler-4] UpdatePassword: %s", "new password and confirm password does not match")
		res.Message = "new password and confirm password does not match"
		res.Data = nil
		res.Code = http.StatusUnprocessableEntity
		res.Success = false
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	err = u.userService.UpdatePassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdatePassword: %v", err)
		if err.Error() == "404" {
			res.Message = "User not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}

		if err.Error() == "401" {
			res.Message = "Token expired or invalid"
			res.Data = nil
			res.Code = http.StatusUnauthorized
			res.Success = false
			return c.JSON(http.StatusUnauthorized, res)
		}
		res.Message = err.Error()
		res.Data = nil
		res.Code = http.StatusInternalServerError
		res.Success = false
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Data = nil
	res.Message = "Password updated successfully"
	res.Code = http.StatusOK
	res.Success = true

	return c.JSON(http.StatusOK, res)
}

// ForgotPassword implements UserHandlerInterface.
func (u *userHandler) ForgotPassword(c echo.Context) error {
	var (
		req = request.ForgotPasswordRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] ForgotPassword: %v", err)
		res.Message = "Invalid request body format"
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] ForgotPassword: %v", err)

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

	reqEntity := entity.UserEntity{
		Email: req.Email,
	}

	err = u.userService.ForgotPassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-3] ForgotPassword: %v", err)
		if err.Error() == "404" {
			res.Code = http.StatusNotFound
			res.Message = "User not found"
			res.Success = false
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "Forgot password successfully, please check your email!"
	res.Data = nil

	return c.JSON(http.StatusOK, res)

}

// VerifyAccount implements UserHandlerInterface.
func (u *userHandler) VerifyAccount(c echo.Context) error {
	var (
		res              = response.ResponseDefault{}
		resVerifyAccount = response.SignInResponse{}
		ctx              = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Errorf("[UserHandler-1] VefifyAccount: %s", "Missing or invalid token")
		res.Code = http.StatusUnauthorized
		res.Data = nil
		res.Message = "Missing or invalid token"
		res.Success = false
		return c.JSON(http.StatusUnauthorized, res)
	}

	user, err := u.userService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Errorf("[UserHandler-2] VefifyAccount: %s", err)
		if err.Error() == "404" {
			res.Code = http.StatusNotFound
			res.Data = nil
			res.Message = "User not found"
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		if err.Error() == "401" {
			res.Code = http.StatusUnauthorized
			res.Data = nil
			res.Message = "Token expired or invalid"
			res.Success = false
			return c.JSON(http.StatusUnauthorized, res)
		}
		res.Code = http.StatusInternalServerError
		res.Data = nil
		res.Message = err.Error()
		res.Success = false
		return c.JSON(http.StatusInternalServerError, res)
	}

	resVerifyAccount.ID = user.ID
	resVerifyAccount.Name = user.Name
	resVerifyAccount.Email = user.Email
	resVerifyAccount.Role = user.RoleName
	resVerifyAccount.Lat = user.Lat
	resVerifyAccount.Lng = user.Lng
	resVerifyAccount.Phone = user.Phone
	resVerifyAccount.AccessToken = user.Token

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "Success verify account"
	res.Data = resVerifyAccount

	return c.JSON(http.StatusOK, res)

}

// CreateUserAccount implements UserHandlerInterface.
func (u *userHandler) CreateUserAccount(c echo.Context) error {
	var (
		req = request.SignUpRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] CreateUserAccount: %v", err)
		res.Message = "Invalid request body format"
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] CreateUserAccount: %v", err)

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

	if req.Password != req.ConfirmPassword {
		err = errors.New("password and confirm password must be same")
		log.Errorf("[UserHandler-3] CreateUserAccount: %v", err)
		res.Message = "Password and confirm password must be same"
		res.Success = false
		res.Code = http.StatusUnprocessableEntity
		res.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err = u.userService.CreateUserAccount(ctx, reqEntity)

	if errors.Is(err, service.ErrUserExist) {
		log.Errorf("[UserHandler-4] CreateUserAccount: %v", err)
		res.Message = "Email already exists"
		res.Success = false
		res.Code = http.StatusConflict
		res.Data = nil
		return c.JSON(http.StatusConflict, res)
	}
	if err != nil {
		log.Errorf("[UserHandler-5] CreateUserAccount: %v", err)
		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "User created successfully"
	res.Success = true
	res.Code = http.StatusCreated
	res.Data = nil
	return c.JSON(http.StatusCreated, res)

}

var err error

// SignIn implements UserHandlerInterface.
func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req       = request.SignInRequest{}
		res       = response.ResponseDefault{}
		resSignIn = response.SignInResponse{}
		ctx       = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		res.Message = "Invalid request body format"
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] SignIn: %v", err)

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

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.userService.SignIn(ctx, reqEntity)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrInvalidPassword):
			log.Errorf("[UserHandler-3] SignIn: %v", err)
			res.Code = http.StatusUnauthorized
			res.Message = "Email or password is incorrect"
			res.Success = false
			res.Data = nil
			return c.JSON(http.StatusUnauthorized, res)
		default:
			log.Errorf("[UserHandler-4] SignIn: %v", err)
			res.Message = err.Error()
			res.Success = false
			res.Code = http.StatusInternalServerError
			res.Data = nil
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	resSignIn.ID = user.ID
	resSignIn.Name = user.Name
	resSignIn.Email = user.Email
	resSignIn.Role = user.RoleName
	resSignIn.Lat = user.Lat
	resSignIn.Lng = user.Lng
	resSignIn.Phone = user.Phone
	resSignIn.AccessToken = token

	res.Code = http.StatusOK
	res.Success = true
	res.Message = "Success Login"
	res.Data = resSignIn

	return c.JSON(http.StatusOK, res)

}

func NewUserHandler(g *echo.Group, userService service.UserServiceInterface, cfg *config.Config, jwtService service.JwtServiceInterface) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}

	g.POST("/auth/login", userHandler.SignIn)
	g.GET("/auth/validate-forgot-token", userHandler.ValidateForgotPasswordToken)
	g.POST("/auth/forgot-password", userHandler.ForgotPassword)
	g.PATCH("/auth/reset-password", userHandler.UpdatePassword)
	g.POST("/auth", userHandler.CreateUserAccount)
	g.GET("/auth/verify-account", userHandler.VerifyAccount)

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)
	g.Use(mid.CheckToken())
	// adminGroup := g.Group("/admin", mid.CheckToken())
	g.GET("/profile", userHandler.GetProfileUser)
	g.PATCH("/update-profile", userHandler.UpdateDataUser)
	g.PATCH("/change-password", userHandler.ChangePassword)
	return userHandler
}
