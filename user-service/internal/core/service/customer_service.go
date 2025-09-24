package service

import (
	"context"
	"errors"
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

type CustomerServiceInterface interface {
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) error
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int64) error
	UpdateLocationCustomer(ctx context.Context, req entity.UserEntity) error
}

type CustomerService struct {
	repo       repository.CustomerRepositoryInterface
	repoAuth   repository.AuthRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
	repoToken  repository.VerificationTokenRepositoryInterface
}

// UpdateLocationCustomer implements CustomerServiceInterface.
func (u *CustomerService) UpdateLocationCustomer(ctx context.Context, req entity.UserEntity) error {

	return u.repo.UpdateLocationCustomer(ctx, req)

}

// CreateCustomer implements CustomerServiceInterface.
func (u *CustomerService) CreateCustomer(ctx context.Context, req entity.UserEntity) error {

	existingUser, err := u.repoAuth.FindUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		log.Errorf("[UserService-2] CreateUserAccount: %v", err)
		return err
	}
	if existingUser != nil {
		return errs.ErrUserExist
	}

	passwordNoEncrypt := req.Password
	password, err := conv.HashPassword(passwordNoEncrypt)
	if err != nil {
		log.Fatalf("[CustomerService-2] CreateCustomer: %v", err)
		return err
	}
	req.Password = password

	userID, err := u.repo.CreateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("[CustomerService-3] CreateCustomer: %v", err)
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

// DeleteCustomer implements CustomerServiceInterface.
func (u *CustomerService) DeleteCustomer(ctx context.Context, customerID int64) error {
	return u.repo.DeleteCustomer(ctx, customerID)
}

// GetCustomerAll implements CustomerServiceInterface.
func (u *CustomerService) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error) {
	return u.repo.GetCustomerAll(ctx, query)
}

// GetCustomerByID implements CustomerServiceInterface.
func (u *CustomerService) GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error) {
	return u.repo.GetCustomerByID(ctx, customerID)
}

// UpdateCustomer implements CustomerServiceInterface.
func (u *CustomerService) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	passwordNoencrypt := ""
	if req.Password != "" {
		passwordNoencrypt = req.Password
		password, err := conv.HashPassword(req.Password)
		if err != nil {
			log.Fatalf("[CustomerService-1] UpdateCustomer: %v", err)
			return err
		}

		req.Password = password
	}

	err := u.repo.UpdateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("[CustomerService-2] UpdateCustomer: %v", err)
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

func NewCustomerService(repo repository.CustomerRepositoryInterface, repoAuth repository.AuthRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) CustomerServiceInterface {
	return &CustomerService{
		repo:       repo,
		repoAuth:   repoAuth,
		cfg:        cfg,
		jwtService: jwtService,
		repoToken:  repoToken,
	}
}
