package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"user-service/config"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/utils"
	"user-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type AuthServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	ForgotPassword(ctx context.Context, req entity.UserEntity) error
	UpdatePassword(ctx context.Context, req entity.UserEntity) error
	ValidateForgotPasswordToken(ctx context.Context, token string) error
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error)
}

type AuthService struct {
	repo       repository.AuthRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
	repoToken  repository.VerificationTokenRepositoryInterface
}


// SignIn implements AuthServiceInterface.
func (u *AuthService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-1] SignIn: %v", err)
		return nil, "", errs.ErrUserNotFound
	}

	if checkPass := conv.CheckPasswordHash(req.Password, user.Password); !checkPass {
		err = errs.ErrInvalidPassword
		log.Errorf("[UserService-2] SignIn: %v", err)
		return nil, "", err
	}

	token, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Errorf("[UserService-3] SignIn: %v", err)
		return nil, "", err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
		"role":       user.RoleName,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return nil, "", err
	}

	redisConn := config.NewConfig().NewRedisClient()
	err = redisConn.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Errorf("[UserService-4] SignIn: %v", err)
		return nil, "", err
	}

	// Set TTL selama 24 jam
	err = redisConn.Expire(ctx, token, 24*time.Hour).Err()
	if err != nil {
		log.Errorf("[UserService-5] SignIn: %v", err)
		return nil, "", err
	}

	return user, token, nil

}

// ForgotPassword implements AuthServiceInterface.
func (u *AuthService) ForgotPassword(ctx context.Context, req entity.UserEntity) error {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-1] ForgotPassword: %v", err)
		return err
	}

	token := uuid.New().String()
	reqEntity := entity.VerificationTokenEnity{
		UserID:    user.ID,
		Token:     token,
		TokenType: "forgot_password",
		ExpiresAt: time.Now().Add(time.Minute * 30),
	}

	err = u.repoToken.CreateVerification(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserService-2] ForgotPassword: %v", err)
		return err
	}

	urlForgotPassword := fmt.Sprintf("%s/forgot-password?token=%s", u.cfg.App.UrlFrontend, token)
	messageParam := fmt.Sprintf("Please click link bellow for reset password: %v", urlForgotPassword)
	err = message.PublishMessage(req.ID, req.Email, messageParam, utils.NOTIF_EMAIL_FORGOT_PASSWORD, "forgot_password")
	if err != nil {
		log.Errorf("[UserService-3] ForgotPassword: %v", err)
		return err
	}
	return nil
}

// ValidateForgotPasswordToken implements AuthServiceInterface.
func (u *AuthService) ValidateForgotPasswordToken(ctx context.Context, token string) error {
	data, err := u.repoToken.GetDataWithoutDelete(ctx, token)
	if err != nil {
		log.Errorf("[UserService-1] ValidateForgotPasswordToken: %v", err)
		return err
	}

	if data.TokenType != "forgot_password" {
		err = errs.ErrInvalidToken
		log.Errorf("[UserService-2] ValidateForgotPasswordToken: Invalid token type")
		return err
	}

	return nil
}


// UpdatePassword implements AuthServiceInterface.
func (u *AuthService) UpdatePassword(ctx context.Context, req entity.UserEntity) error {
	token, err := u.repoToken.GetDataByToken(ctx, req.Token, "forgot_password")
	if err != nil {
		log.Errorf("[UserService-1] UpdatePassword: %v", err)
		return err
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService-2] UpdatePassword: %v", err)
		return err
	}
	req.Password = password
	req.ID = token.UserID

	err = u.repo.UpdatePasswordByID(ctx, req)
	if err != nil {
		log.Errorf("[UserService-3] UpdatePassword: %v", err)
		return err
	}

	return nil
}


// CreateUserAccount implements AuthServiceInterface.
func (u *AuthService) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService-1] CreateUserAccount: %v", err)
		return err
	}

	req.Password = password

	existingUser, err := u.repo.FindUserByEmail(ctx, req.Email)
	if err != nil  && !errors.Is(err, errs.ErrUserNotFound) {
		log.Errorf("[UserService-2] CreateUserAccount: %v", err)
		return err
	}
	if existingUser != nil {
		return errs.ErrUserExist
	}

	token := uuid.New().String()
	req.Token = token

	err = u.repo.CreateUserAccount(ctx, req)
	if err != nil {
		log.Errorf("[UserService-3] CreateUserAccount: %v", err)
		return err
	}

	urlVerify := fmt.Sprintf("%s/verify?token=%s", u.cfg.App.UrlFrontend, req.Token)
	messageParams := fmt.Sprintf("Please verify your account. Token: %s", urlVerify)
	err = message.PublishMessage(req.ID, req.Email, messageParams, utils.NOTIF_EMAIL_VERIFICATION, "Verification")
	if err != nil {
		log.Errorf("[UserService-4] CreateUserAccount: %v", err)
		return err
	}

	return nil

}

// VerifyToken implements AuthServiceInterface.
func (u *AuthService) VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error) {
	verifyToken, err := u.repoToken.GetDataByToken(ctx, token, "email_verification")
	if err != nil {
		log.Errorf("[UserService-1] VerifyToken: %v", err)
		return nil, err
	}

	user, err := u.repo.UpdateUserVerified(ctx, verifyToken.UserID)
	if err != nil {
		log.Errorf("[UserService-2] VerifyToken: %v", err)
		return nil, err
	}

	accessToken, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Errorf("[UserService-3] VerifyToken: %v", err)
		return nil, err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.RoleName,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return nil, err
	}

	redisConn := config.NewConfig().NewRedisClient()
	err = redisConn.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Errorf("[UserService-4] VerifyToken: %v", err)
		return nil, err
	}

	// Set TTL selama 24 jam
	err = redisConn.Expire(ctx, token, 24*time.Hour).Err()
	if err != nil {
		log.Errorf("[UserService-5] VerifyToken: %v", err)
		return nil, err
	}

	user.Token = accessToken

	return user, nil
}



func NewAuthService(repo repository.AuthRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) AuthServiceInterface {
	return &AuthService{
		repo:       repo,
		cfg:        cfg,
		jwtService: jwtService,
		repoToken:  repoToken,
	}
}