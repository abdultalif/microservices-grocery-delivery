package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/message"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/utils"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/utils/conv"

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
	publisher  *message.UserEventPublisher
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
		log.Errorf("[CustomerService-2] CreateCustomer: %v", err)
		return err
	}
	req.Password = password

	userID, err := u.repo.CreateCustomer(ctx, req)
	if err != nil {
		log.Errorf("[CustomerService-3] CreateCustomer: %v", err)
		return err
	}

	messageparam := fmt.Sprintf("You have been registered in Sayur Project. Please login use: \n Email: %s\nPassword: %s", req.Email, passwordNoEncrypt)
	go message.PublishMessage(userID,
		req.Email,
		messageparam,
		utils.NOTIF_EMAIL_CREATE_CUSTOMER,
		"Account Exists")

	req.ID = userID
	go func() {
		if err := u.publisher.Publish("user.created", req); err != nil {
			log.Errorf("[CustomerService-1] PublishUserEvent error: %v", err)
		}
	}()

	return nil
}

// DeleteCustomer implements CustomerServiceInterface.
func (u *CustomerService) DeleteCustomer(ctx context.Context, customerID int64) error {
	err := u.repo.DeleteCustomer(ctx, customerID)
	if err != nil {
		log.Errorf("[CustomerService-1] DeleteCustomer: %v", err)
		return err
	}

	go func() {
		if err := u.publisher.Publish("user.deleted", entity.UserEntity{ID: customerID}); err != nil {
			log.Errorf("[CustomerService-2] PublishUserEvent error: %v", err)
		}
	}()

	return nil
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
			log.Errorf("[CustomerService-1] UpdateCustomer: %v", err)
			return err
		}

		req.Password = password
	}

	err := u.repo.UpdateCustomer(ctx, req)
	if err != nil {
		log.Errorf("[CustomerService-2] UpdateCustomer: %v", err)
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

	go func() {
		if err := u.publisher.Publish("user.updated", req); err != nil {
			log.Errorf("[CustomerService-3] PublishUserEvent error: %v", err)
		}
	}()

	return nil
}

func NewCustomerService(repo repository.CustomerRepositoryInterface, repoAuth repository.AuthRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface, publisher *message.UserEventPublisher) CustomerServiceInterface {
	return &CustomerService{
		repo:       repo,
		repoAuth:   repoAuth,
		cfg:        cfg,
		jwtService: jwtService,
		repoToken:  repoToken,
		publisher:  publisher,
	}
}
