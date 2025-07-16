package service

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/utils"
	"user-service/utils/conv"

	"github.com/labstack/gommon/log"
)

type UserServiceInterface interface {
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
	repoAuth	repository.AuthRepositoryInterface
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

	existingUser, err := u.repoAuth.FindUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-1] CreateCustomer: %v", err)
		return err
	}
	if existingUser != nil {
		return errs.ErrUserExist
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
		return errs.ErrUserNotFound
	}

	if !conv.CheckPasswordHash(currentPassword, userData.Password) {
		log.Errorf("[UserService-2] ChangePassword - wrong current password")
		return errs.ErrCurrentPasswordIncorrect
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService-3] ChangePassword: %v", err)
		return err
	}

	req.Password = password

	err = u.repoAuth.UpdatePasswordByID(ctx, req)
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


func NewUserService(repo repository.UserRepositoryInterface, repoAuth repository.AuthRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) UserServiceInterface {
	return &UserService{
		repo:       repo,
		repoAuth:   repoAuth,
		cfg:        cfg,
		jwtService: jwtService,
		repoToken:  repoToken,
	}
}
