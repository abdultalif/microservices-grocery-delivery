package service

import (
	"context"
	"user-service/config"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/utils/conv"

	"github.com/labstack/gommon/log"
)

type UserServiceInterface interface {
	GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
	ChangePassword(ctx context.Context, req entity.UserEntity, currentPassword string) error
	UploadPhoto(ctx context.Context, userID int64, photoURL string) error
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
