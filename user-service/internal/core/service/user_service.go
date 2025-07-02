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
	"user-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	ForgotPassword(ctx context.Context, req entity.UserEntity) error
	UpdatePassword(ctx context.Context, req entity.UserEntity) error
	ValidateForgotPasswordToken(ctx context.Context, token string) error
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error)
	GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error)
}

type UserService struct {
	repo       repository.UserRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
	repoToken  repository.VerificationTokenRepositoryInterface
}

// GetProfileUser implements UserServiceInterface.
func (u *UserService) GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		log.Errorf("[UserService-1] GetProfileUser: %v", err)
		return nil, err
	}
	return user, nil
}

// ValidateForgotPasswordToken implements UserServiceInterface.
func (u *UserService) ValidateForgotPasswordToken(ctx context.Context, token string) error {
	data, err := u.repoToken.GetDataWithoutDelete(ctx, token)
	if err != nil {
		log.Errorf("[UserService-1] ValidateForgotPasswordToken: %v", err)
		return err
	}

	if data.TokenType != "forgot_password" {
		err = errors.New("401")
		log.Errorf("[UserService-2] ValidateForgotPasswordToken: Invalid token type")
		return err
	}

	return nil
}

// UpdatePassword implements UserServiceInterface.
func (u *UserService) UpdatePassword(ctx context.Context, req entity.UserEntity) error {
	token, err := u.repoToken.GetDataByToken(ctx, req.Token, "forgot_password")
	if err != nil {
		log.Errorf("[UserService-1] UpdatePassword: %v", err)
		return err
	}

	if token.TokenType != "forgot_password" {
		err = errors.New("401")
		log.Errorf("[UserService-2] UpdatePassword: %v", err)
		return err
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService-3] UpdatePassword: %v", err)
		return err
	}
	req.Password = password
	req.ID = token.UserID

	err = u.repo.UpdatePasswordByID(ctx, req)
	if err != nil {
		log.Errorf("[UserService-4] UpdatePassword: %v", err)
		return err
	}

	return nil
}

// VerifyToken implements UserServiceInterface.
func (u *UserService) VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error) {
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
		"token":      token,
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return nil, err
	}

	redisConn := config.NewConfig().NewRedisClient()
	err = redisConn.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Errorf("[UserService-4] SignIn: %v", err)
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

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserExist       = errors.New("Email already exists")
)

// ForgotPassword implements UserServiceInterface.
func (u *UserService) ForgotPassword(ctx context.Context, req entity.UserEntity) error {
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
	err = message.PublishMessage(req.Email, messageParam, "forgot_password")
	if err != nil {
		log.Errorf("[UserService-3] ForgotPassword: %v", err)
		return err
	}
	return nil
}

// CreateUserAccount implements UserServiceInterface.
func (u *UserService) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService1] CreateUserAccount: %v", err)
		return err
	}

	req.Password = password

	existingUser, err := u.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-0] CreateUserAccount: %v", err)
		return err
	}
	if existingUser != nil {
		return ErrUserExist
	}

	token := uuid.New().String()
	req.Token = token

	err = u.repo.CreateUserAccount(ctx, req)
	if err != nil {
		log.Errorf("[UserService-2] CreateUserAccount: %v", err)
		return err
	}

	urlVerify := fmt.Sprintf("http://localhost:%s/verify?token=%s", u.cfg.App.AppPort, req.Token)
	messageParams := fmt.Sprintf("Please verify your account. Token: %s", urlVerify)
	err = message.PublishMessage(req.Email, messageParams, "email_verification")
	if err != nil {
		log.Errorf("[UserService-3] CreateUserAccount: %v", err)
		return err
	}

	return nil

}

// SignIn implements UserServiceInterface.
func (u *UserService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-1] SignIn: %v", err)
		return nil, "", ErrUserNotFound
	}

	if checkPass := conv.CheckPasswordHash(req.Password, user.Password); !checkPass {
		err = ErrInvalidPassword
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
		"role_name":  user.RoleName,
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

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) UserServiceInterface {
	return &UserService{
		repo:       repo,
		cfg:        cfg,
		jwtService: jwtService,
		repoToken:  repoToken,
	}
}
