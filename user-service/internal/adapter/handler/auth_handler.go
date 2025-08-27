package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"user-service/config"
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

	GenerateServiceToken(ctx echo.Context) error
}

type authHandler struct {
	authService service.AuthServiceInterface
	jwtService  service.JwtServiceInterface
	cfg         *config.Config
}

// GenerateServiceToken implements AuthHandlerInterface.
func (u *authHandler) GenerateServiceToken(c echo.Context) error {
	var (
		req = request.TokenRequest{}
		ctx = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[AuthHandler-1] GenerateServiceToken: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body format", nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[AuthHandler-2] GenerateServiceToken: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	if req.ClientID != u.cfg.App.AuthClientID || req.ClientSecret != u.cfg.App.AuthClientSecret {
		log.Errorf("[AuthHandler-3] GenerateServiceToken: %s", "Invalid client ID or secret")
		return c.JSON(http.StatusUnauthorized,
			response.ResponseAPI(false, http.StatusUnauthorized, "Invalid client ID or secret", nil))
	}

	token, err := u.jwtService.GenerateToken(0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	sessionData := map[string]interface{}{
		"user_id":    0,
		"name":       "internal-service",
		"email":      "internal@system.local",
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
		"role":       "Super Admin",
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		log.Errorf("[AuthHandler-4] GenerateServiceToken marshal: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to prepare session data", nil))
	}

	redisConn, err := config.NewConfig().NewRedisClient()
	if err != nil {
		log.Errorf("[AuthHandler-5] GenerateServiceToken redis connect: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to connect to redis", nil))
	}

	if err := redisConn.Set(ctx, token, jsonData, 1*time.Hour).Err(); err != nil {
		log.Errorf("[AuthHandler-6] GenerateServiceToken redis set: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to set session data in redis", nil))
	}

	// TTL untuk token
	if err := redisConn.Expire(ctx, token, 1*time.Hour).Err(); err != nil {
		log.Errorf("[AuthHandler-7] GenerateServiceToken redis expire: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to set token expiration in redis", nil))
	}

	data := response.TokenResponse{
		AccessToken: token,
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Service token generated successfully", data))
}

// SignIn implements AuthHandlerInterface.
func (u *authHandler) SignIn(c echo.Context) error {
	var (
		req       = request.SignInRequest{}
		resSignIn = response.SignInResponse{}
		ctx       = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[AuthHandler-1] SignIn: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body format", nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[AuthHandler-2] SignIn: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.authService.SignIn(ctx, reqEntity)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUserNotFound), errors.Is(err, errs.ErrInvalidPassword):
			log.Errorf("[AuthHandler-3] SignIn: %v", err)
			return c.JSON(http.StatusUnauthorized,
				response.ResponseAPI(false, http.StatusUnauthorized, "Email or password is incorrect", nil))
		default:
			log.Errorf("[AuthHandler-4] SignIn: %v", err)
			return c.JSON(http.StatusInternalServerError,
				response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
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

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Success Login", resSignIn))

}

// CreateUserAccount implements AuthHandlerInterface.
func (u *authHandler) CreateUserAccount(c echo.Context) error {
	var (
		req = request.SignUpRequest{}
		ctx = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[AuthHandler-1] CreateUserAccount: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body format", nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[AuthHandler-2] CreateUserAccount: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	if req.Password != req.ConfirmPassword {
		log.Errorf("[AuthHandler-3] CreateUserAccount: %s", "Password and confirm password must be same")
		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, "Password and confirm password must be same", nil))
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err := u.authService.CreateUserAccount(ctx, reqEntity)

	if errors.Is(err, errs.ErrUserExist) {
		log.Errorf("[AuthHandler-4] CreateUserAccount: %v", err)
		return c.JSON(http.StatusConflict,
			response.ResponseAPI(false, http.StatusConflict, "Email already exists", nil))
	}
	if err != nil {
		log.Errorf("[AuthHandler-5] CreateUserAccount: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	return c.JSON(http.StatusCreated,
		response.ResponseAPI(true, http.StatusCreated, "User created successfully", nil))

}

// VerifyAccount implements AuthHandlerInterface.
func (u *authHandler) VerifyAccount(c echo.Context) error {
	var (
		resVerifyAccount = response.SignInResponse{}
		ctx              = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Errorf("[AuthHandler-1] VerifyAccount: %s", "Missing or invalid token")
		return c.JSON(http.StatusUnauthorized,
			response.ResponseAPI(false, http.StatusUnauthorized, "Missing or invalid token", nil))
	}

	user, err := u.authService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Errorf("[AuthHandler-2] VerifyAccount: %s", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "User not found", nil))
		}
		if errors.Is(err, errs.ErrTokenExpired) || errors.Is(err, errs.ErrInvalidToken) {
			return c.JSON(http.StatusUnauthorized,
				response.ResponseAPI(false, http.StatusUnauthorized, "Token expired or invalid", nil))
		}

		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))

	}

	resVerifyAccount.ID = user.ID
	resVerifyAccount.Name = user.Name
	resVerifyAccount.Email = user.Email
	resVerifyAccount.Role = user.RoleName
	resVerifyAccount.Lat = user.Lat
	resVerifyAccount.Lng = user.Lng
	resVerifyAccount.Phone = user.Phone
	resVerifyAccount.AccessToken = user.Token

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Success verify account", resVerifyAccount))

}

// ForgotPassword implements AuthHandlerInterface.
func (u *authHandler) ForgotPassword(c echo.Context) error {
	var (
		req = request.ForgotPasswordRequest{}
		ctx = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[AuthHandler-1] ForgotPassword: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body format", nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[AuthHandler-2] ForgotPassword: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	reqEntity := entity.UserEntity{
		Email: req.Email,
	}

	err := u.authService.ForgotPassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[AuthHandler-3] ForgotPassword: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "User not found", nil))
		}

		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Forgot password successfully, please check your email!", nil))

}

// ValidateForgotPasswordToken implements AuthHandlerInterface.
func (u *authHandler) ValidateForgotPasswordToken(c echo.Context) error {
	ctx := c.Request().Context()

	token := c.QueryParam("token")
	if token == "" {
		log.Infof("[AuthHandler-1] ValidateForgotPasswordToken: %s", "Token is required")
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Token is required", nil))
	}

	err := u.authService.ValidateForgotPasswordToken(ctx, token)
	log.Infof("[AuthHandler-2] ValidateForgotPasswordToken: %s", err)
	if err != nil {
		if errors.Is(err, errs.ErrTokenExpired) || errors.Is(err, errs.ErrInvalidToken) {
			return c.JSON(http.StatusUnauthorized,
				response.ResponseAPI(false, http.StatusUnauthorized, "Token is invalid or expired", nil))
		} else {
			return c.JSON(http.StatusInternalServerError,
				response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
		}
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Token is valid", nil))
}

// UpdatePassword implements AuthHandlerInterface.
func (u *authHandler) UpdatePassword(c echo.Context) error {
	var (
		req = request.UpdatePasswordRequest{}
		ctx = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	log.Infof("[AuthHandler-1] UpdatePassword: %s", "Token is required")
	if tokenString == "" {
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Token is required", nil))
	}

	if err := c.Bind(&req); err != nil {
		log.Infof("[AuthHandler-2] UpdatePassword: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, err.Error(), nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[AuthHandler-3] UpdatePassword: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Infof("[AuthHandler-4] UpdatePassword: %s", "new password and confirm password does not match")
		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, "new password and confirm password does not match", nil))
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	err := u.authService.UpdatePassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[AuthHandler-5] UpdatePassword: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "User not found", nil))
		} else if errors.Is(err, errs.ErrTokenExpired) || errors.Is(err, errs.ErrInvalidToken) {
			return c.JSON(http.StatusUnauthorized,
				response.ResponseAPI(false, http.StatusUnauthorized, "Token expired or invalid", nil))
		} else {
			return c.JSON(http.StatusInternalServerError,
				response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
		}
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Password updated successfully", nil))
}

func NewAuthHandler(
	authService service.AuthServiceInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
) AuthHandlerInterface {
	return &authHandler{
		authService: authService,
		cfg:         cfg,
		jwtService:  jwtService,
	}
}
