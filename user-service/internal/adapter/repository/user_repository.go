package repository

import (
	"context"
	"errors"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
	UploadPhoto(ctx context.Context, userID int64, photoURL string) error
}

type UserRepository struct {
	db *gorm.DB
}

// UploadPhoto implements UserRepositoryInterface.
func (u *UserRepository) UploadPhoto(ctx context.Context, userID int64, photoURL string) error {
	if err := u.db.Model(&model.User{}).Where("id = ?", userID).Update("photo", photoURL).Error; err != nil {
		log.Errorf("[UserRepository-1] UpdatePhoto: %v", err)
		return err
	}
	return nil
}

// UpdateDataUser implements UserRepositoryInterface.
func (u *UserRepository) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	modelUser := model.User{
		Name:    req.Name,
		Email:   req.Email,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Address: req.Address,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	if err := u.db.Where("id = ? AND is_verified = true", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] UpdateDataUser: %v", err)
			return err
		}
		log.Errorf("[UserRepository-2] UpdateDataUser: %v", err)
		return err
	}

	if err := u.db.Save(&modelUser).Error; err != nil {
		log.Errorf("[UserRepository-3] UpdateDataUser: %v", err)
		return err
	}

	return nil

}

// GetUserByID implements UserRepositoryInterface.
func (u *UserRepository) GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.Where("id = ? AND is_verified = true", userID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrUserNotFound
			log.Errorf("[UserRepository-1] GetUserByID: %v", err)
			return nil, err
		}
		log.Errorf("[UserRepository-2] GetUserByID: %v", err)
		return nil, err
	}

	return &entity.UserEntity{
		ID:       modelUser.ID,
		Name:     modelUser.Name,
		Email:    modelUser.Email,
		Password: modelUser.Password,
		RoleName: modelUser.Roles[0].Name,
		Lat:      modelUser.Lat,
		Lng:      modelUser.Lng,
		Address:  modelUser.Address,
		Phone:    modelUser.Phone,
		Photo:    modelUser.Photo,
	}, nil
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}
