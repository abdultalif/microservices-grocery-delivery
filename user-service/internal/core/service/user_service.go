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
	"user-service/utils"
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
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
	ChangePassword(ctx context.Context, req entity.UserEntity, currentPassword string) error
	UploadPhoto(ctx context.Context, userID int64, photoURL string) error

	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) error
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int64) error
}

type UserService struct {
	repo       repository.UserRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
	repoToken  repository.VerificationTokenRepositoryInterface
}

// UploadPhoto implements UserServiceInterface.
func (u *UserService) UploadPhoto(ctx context.Context, userID int64, photoURL string) error {
	return u.repo.UploadPhoto(ctx, userID, photoURL)
}

// CreateCustomer implements UserServiceInterface.
func (u *UserService) CreateCustomer(ctx context.Context, req entity.UserEntity) error {

	existingUser, err := u.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-1] CreateCustomer: %v", err)
		return err
	}
	if existingUser != nil {
		return ErrUserExist
	}

	passwordNoEncrypt := req.Password
	password, err := conv.HashPassword(passwordNoEncrypt)
	if err != nil {
		log.Fatalf("[UserService-2] CreateCustomer: %v", err)
		return err
	}
	req.Password = password

	userID, err := u.repo.CreateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("[UserService-3] CreateCustomer: %v", err)
		return err
	}

	messageparam := fmt.Sprintf("You have been registered in Sayur Project. Please login use: \n Email: %s\nPassword: %s", req.Email, passwordNoEncrypt)
	go message.PublishMessage(userID,
		req.Email,
		messageparam,
		utils.NOTIF_EMAIL_CREATE_CUSTOMER,
		"Account Exists")

	return nil
}

// DeleteCustomer implements UserServiceInterface.
func (u *UserService) DeleteCustomer(ctx context.Context, customerID int64) error {
	return u.repo.DeleteCustomer(ctx, customerID)
}

// GetCustomerAll implements UserServiceInterface.
func (u *UserService) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error) {
	return u.repo.GetCustomerAll(ctx, query)
}

// GetCustomerByID implements UserServiceInterface.
func (u *UserService) GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error) {
	return u.repo.GetCustomerByID(ctx, customerID)
}

// UpdateCustomer implements UserServiceInterface.
func (u *UserService) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	passwordNoencrypt := ""
	if req.Password != "" {
		passwordNoencrypt = req.Password
		password, err := conv.HashPassword(req.Password)
		if err != nil {
			log.Fatalf("[UserService-1] UpdateCustomer: %v", err)
			return err
		}

		req.Password = password
	}

	err := u.repo.UpdateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("[UserService-2] UpdateCustomer: %v", err)
		return err
	}

	if passwordNoencrypt != "" {
		messageparam := fmt.Sprintf("You're account has been updated. Please login use: \n Email: %s\nPassword: %s", req.Email, passwordNoencrypt)
		go message.PublishMessage(req.ID,
			req.Email,
			messageparam,
			utils.NOTIF_EMAIL_UPDATE_CUSTOMER,
			"Updated Data")
	}

	return nil
}

// ChangePassword implements UserServiceInterface.
func (u *UserService) ChangePassword(ctx context.Context, req entity.UserEntity, currentPassword string) error {

	userData, err := u.repo.GetUserByID(ctx, req.ID)
	if err != nil {
		log.Errorf("[UserService-1] ChangePassword - user not found: %v", err)
		return errors.New("user not found")
	}

	if !conv.CheckPasswordHash(currentPassword, userData.Password) {
		log.Errorf("[UserService-2] ChangePassword - wrong current password")
		return errors.New("401")
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService-3] ChangePassword: %v", err)
		return err
	}

	req.Password = password

	err = u.repo.UpdatePasswordByID(ctx, req)
	if err != nil {
		log.Errorf("[UserService-4] ChangePassword: %v", err)
		return err
	}
	return nil
}

// UpdateDataUser implements UserServiceInterface.
func (u *UserService) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {

	err := u.repo.UpdateDataUser(ctx, req)
	if err != nil {
		log.Errorf("[UserService-1] UpdateDataUser: %v", err)
		return err
	}
	return nil

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
	err = message.PublishMessage(req.ID, req.Email, messageParam, utils.NOTIF_EMAIL_FORGOT_PASSWORD, "forgot_password")
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
		log.Errorf("[UserService-1] CreateUserAccount: %v", err)
		return err
	}

	req.Password = password

	existingUser, err := u.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-2] CreateUserAccount: %v", err)
		return err
	}
	if existingUser != nil {
		return ErrUserExist
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

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) UserServiceInterface {
	return &UserService{
		repo:       repo,
		cfg:        cfg,
		jwtService: jwtService,
		repoToken:  repoToken,
	}
}
