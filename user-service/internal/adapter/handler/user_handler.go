package handler

import (
	"errors"
	"net/http"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	v "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type UserHandlerInterface interface {
	SignIn(c echo.Context) error
	CreateUserAccount(c echo.Context) error
	VerifyAccount(c echo.Context) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

// VerifyAccount implements UserHandlerInterface.
func (u *userHandler) VerifyAccount(c echo.Context) error {
	var (
		res = response.ResponseDefault{}
		resVerifyAccount = response.SignInResponse{}
		ctx = c.Request().Context()
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
		res.Code = 422
		res.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] CreateUserAccount: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			res.Message = ve.Errors
			res.Success = false
			res.Code = http.StatusBadRequest
			res.Data = nil
			return c.JSON(http.StatusBadRequest, res)
		}

		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if req.Password != req.ConfirmPassword {
		err = errors.New("password and confirm password must be same")
		log.Errorf("[UserHandler-3] CreateUserAccount: %v", err)
		res.Message = "Password and confirm password must be same"
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
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
		res.Code = 422
		res.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] SignIn: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			res.Message = ve.Errors
			res.Success = false
			res.Code = http.StatusBadRequest
			res.Data = nil
			return c.JSON(http.StatusBadRequest, res)
		}

		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
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

func NewUserHandler(g *echo.Group, userService service.UserServiceInterface, cfg *config.Config) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}

	g.POST("/auth/login", userHandler.SignIn)
	g.POST("/auth", userHandler.CreateUserAccount)
	g.GET("/auth/verify-account", userHandler.VerifyAccount)

	mid := adapter.NewMiddlewareAdapter(cfg)
	g.Use(mid.CheckToken())
	adminGroup := g.Group("/admin", mid.CheckToken())
	adminGroup.GET("/current", func(c echo.Context) error {
		return c.JSON(http.StatusOK, c.Get("user"))
	})
	return userHandler
}
