package service

import (
	"context"
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
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
}

type UserService struct {
	repo       repository.UserRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserExist 	   = errors.New("Email already exists")
)

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
	err = message.PublishMessage(req.Email, messageParams, "email_verfication") 
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
		"token":      token,
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
	}

	redisConn := config.NewConfig().NewRedisClient()
	err = redisConn.HSet(ctx, token, sessionData).Err()
	if err != nil {
		log.Errorf("[UserService-4] SignIn: %v", err)
		return nil, "", err
	}

	return user, token, nil

}

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config, jwtService JwtServiceInterface) UserServiceInterface {
	return &UserService{
		repo:       repo,
		cfg:        cfg,
		jwtService: jwtService,
	}
}
