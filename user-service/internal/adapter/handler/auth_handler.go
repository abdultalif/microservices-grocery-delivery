package handler

import (
	"errors"
	"net/http"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/internal/core/service"
	v "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type AuthHandlerInterface interface {
	SignIn(ctx echo.Context) error
	ForgotPassword(ctx echo.Context) error
	UpdatePassword(c echo.Context) error
	ValidateForgotPasswordToken(ctx echo.Context) error
	CreateUserAccount(c echo.Context) error
	VerifyAccount(c echo.Context) error
}

type authHandler struct {
	authService service.AuthServiceInterface
}

// SignIn implements AuthHandlerInterface.
func (u *authHandler) SignIn(c echo.Context) error {
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

	user, token, err := u.authService.SignIn(ctx, reqEntity)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUserNotFound), errors.Is(err, errs.ErrInvalidPassword):
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



// CreateUserAccount implements AuthHandlerInterface.
func (u *authHandler) CreateUserAccount(c echo.Context) error {
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

	err = u.authService.CreateUserAccount(ctx, reqEntity)

	if errors.Is(err, errs.ErrUserExist) {
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

// VerifyAccount implements AuthHandlerInterface.
func (u *authHandler) VerifyAccount(c echo.Context) error {
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

	user, err := u.authService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Errorf("[UserHandler-2] VefifyAccount: %s", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Code = http.StatusNotFound
			res.Data = nil
			res.Message = "User not found"
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		}
		if errors.Is(err, errs.ErrTokenExpired) || errors.Is(err, errs.ErrInvalidToken) {
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


// ForgotPassword implements AuthHandlerInterface.
func (u *authHandler) ForgotPassword(c echo.Context) error {
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

	err = u.authService.ForgotPassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-3] ForgotPassword: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
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


// ValidateForgotPasswordToken implements AuthHandlerInterface.
func (u *authHandler) ValidateForgotPasswordToken(c echo.Context) error {
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

	err := u.authService.ValidateForgotPasswordToken(ctx, token)
	log.Infof("[UserHandler-1] ValidateForgotPasswordToken: %s", err)
	if err != nil {
		if errors.Is(err, errs.ErrTokenExpired) || errors.Is(err, errs.ErrInvalidToken) {
			res.Code = http.StatusUnauthorized
			res.Message = "Token is invalid or expired"
			res.Success = false
			res.Data = nil
			return c.JSON(http.StatusUnauthorized, res)
		} else {
			res.Code = http.StatusInternalServerError
			res.Message = err.Error()
			res.Success = false
			res.Data = nil
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	res.Code = http.StatusOK
	res.Message = "Token is valid"
	res.Success = true
	res.Data = nil
	return c.JSON(http.StatusOK, res)
}


// UpdatePassword implements AuthHandlerInterface.
func (u *authHandler) UpdatePassword(c echo.Context) error {
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

	err = u.authService.UpdatePassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdatePassword: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Message = "User not found"
			res.Data = nil
			res.Code = http.StatusNotFound
			res.Success = false
			return c.JSON(http.StatusNotFound, res)
		} else if errors.Is(err, errs.ErrTokenExpired) || errors.Is(err, errs.ErrInvalidToken) {
			res.Message = "Token expired or invalid"
			res.Data = nil
			res.Code = http.StatusUnauthorized
			res.Success = false
			return c.JSON(http.StatusUnauthorized, res)
		}else {
			res.Message = err.Error()
			res.Data = nil
			res.Code = http.StatusInternalServerError
			res.Success = false
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	res.Data = nil
	res.Message = "Password updated successfully"
	res.Code = http.StatusOK
	res.Success = true

	return c.JSON(http.StatusOK, res)
}


func NewAuthHandler(g *echo.Group, authService service.AuthServiceInterface) AuthHandlerInterface {
	authHandler := &authHandler{authService: authService}

	g.POST("/auth/login", authHandler.SignIn)
	g.POST("/auth/", authHandler.CreateUserAccount)
	g.GET("/auth/verify-account", authHandler.VerifyAccount)
	g.POST("/auth/forgot-password", authHandler.ForgotPassword)
	g.GET("/auth/validate-forgot-token", authHandler.ValidateForgotPasswordToken)
	g.PATCH("/auth/reset-password", authHandler.UpdatePassword)
	
	return authHandler
}
