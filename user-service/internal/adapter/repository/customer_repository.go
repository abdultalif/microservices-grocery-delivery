package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type CustomerRepositoryInterface interface {
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) (int64, error)
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int64) error
	UpdateLocationCustomer(ctx context.Context, req entity.UserEntity) error
}

type CustomerRepository struct {
	db *gorm.DB
}

// UpdateLocationCustomer implements CustomerRepositoryInterface.
func (u *CustomerRepository) UpdateLocationCustomer(ctx context.Context, req entity.UserEntity) error {

	modelUser := model.User{}
	if err := u.db.Where("id =?", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrUserNotFound
			log.Infof("[CustomerRepository-1] UpdateLocationCustomer: User not found")
			return err
		}
		log.Errorf("[CustomerRepository-2] UpdateLocationCustomer: %v", err)
		return err
	}

	modelUser.Lat = req.Lat
	modelUser.Lng = req.Lng

	if err := u.db.Where("id = ?", req.ID).Updates(&modelUser).Error; err != nil {
		log.Errorf("[CustomerRepository-3] UpdateLocationCustomer: %v", err)
		return err
	}

	return nil

}

// CreateCustomer implements CustomerRepositoryInterface.
func (u *CustomerRepository) CreateCustomer(ctx context.Context, req entity.UserEntity) (int64, error) {
	modelRole := model.Role{}

	if err := u.db.Where("id =?", req.RoleID).First(&modelRole).Error; err != nil {
		log.Fatalf("[CustomerRepository-1] CreateCustomer: %v", err)
		return 0, err
	}

	modelUser := model.User{
		Name:       req.Name,
		Email:      req.Email,
		Password:   req.Password,
		Address:    req.Address,
		Lat:        req.Lat,
		Lng:        req.Lng,
		Phone:      req.Phone,
		Photo:      req.Photo,
		Roles:      []model.Role{modelRole},
		IsVerified: true,
	}

	if err := u.db.Create(&modelUser).Error; err != nil {
		log.Errorf("[CustomerRepository-2] CreateCustomer: %v", err)
		return 0, err
	}

	return modelUser.ID, nil
}

// DeleteCustomer implements CustomerRepositoryInterface.
func (u *CustomerRepository) DeleteCustomer(ctx context.Context, customerID int64) error {
	modelUser := model.User{}
	if err := u.db.Where("id =?", customerID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrUserNotFound
			log.Infof("[CustomerRepository-1] DeleteCustomer: User not found")
			return err
		}
		log.Errorf("[CustomerRepository-2] DeleteCustomer: %v", err)
		return err
	}

	if err := u.db.Delete(&modelUser).Error; err != nil {
		log.Errorf("[CustomerRepository-3] DeleteCustomer: %v", err)
		return err
	}

	return nil
}

// GetCustomerAll implements CustomerRepositoryInterface.
func (u *CustomerRepository) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error) {
	modelUsers := []model.User{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := u.db.Preload("Roles", "name = ?", "Customer").Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")

	if err := sqlMain.Model(&modelUsers).Count(&countData).Error; err != nil {
		log.Errorf("[CustomerRepository-1] GetCustomerAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))

	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelUsers).Error; err != nil {
		log.Errorf("[CustomerRepository-3] GetCustomerAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelUsers) < 1 {
		err := errs.ErrUserNotFound
		log.Infof("[CustomerRepository-4] GetCustomerAll: No Customer found")
		return nil, 0, 0, err
	}

	respEntities := []entity.UserEntity{}
	for _, val := range modelUsers {
		roleName := ""
		for _, role := range val.Roles {
			roleName = role.Name
		}
		respEntities = append(respEntities, entity.UserEntity{
			ID:       val.ID,
			Name:     val.Name,
			Email:    val.Email,
			RoleName: roleName,
			Phone:    val.Phone,
			Photo:    val.Photo,
		})
	}
	return respEntities, countData, int64(totalPage), nil
}

// GetCustomerByID implements CustomerRepositoryInterface.
func (u *CustomerRepository) GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.Where("id = ?", customerID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errs.ErrUserNotFound
			log.Infof("[CustomerRepository-1] GetCustomerByID: User not found")
			return nil, err
		}
		log.Errorf("[CustomerRepository-2] GetCustomerByID: %v", err)
		return nil, err
	}

	roleID := 0
	for _, role := range modelUser.Roles {
		roleID = int(role.ID)
	}

	return &entity.UserEntity{
		ID:       customerID,
		Name:     modelUser.Name,
		Email:    modelUser.Email,
		RoleID:   int64(roleID),
		RoleName: modelUser.Roles[0].Name,
		Address:  modelUser.Address,
		Lat:      modelUser.Lat,
		Lng:      modelUser.Lng,
		Phone:    modelUser.Phone,
		Photo:    modelUser.Photo,
	}, nil
}

// UpdateCustomer implements CustomerRepositoryInterface.
func (u *CustomerRepository) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	modelRole := model.Role{}

	if err := u.db.Where("id =?", req.RoleID).First(&modelRole).Error; err != nil {
		log.Fatalf("[CustomerRepository-1] UpdateCustomer: %v", err)
		return err
	}

	modelUser := model.User{}
	if err := u.db.Where("id =?", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrUserNotFound
			log.Infof("[CustomerRepository-2] UpdateCustomer: User not found")
			return err
		}
		log.Errorf("[CustomerRepository-3] UpdateCustomer: %v", err)
		return err
	}

	modelUser.Name = req.Name
	modelUser.Email = req.Email
	modelUser.Phone = req.Phone
	modelUser.Roles = []model.Role{modelRole}
	if req.Address != "" {
		modelUser.Address = req.Address
	}

	if req.Lat != "" {
		modelUser.Lat = req.Lat
	}

	if req.Lng != "" {
		modelUser.Lng = req.Lng
	}
	if req.Photo != "" {
		modelUser.Lat = req.Lat
	}

	if req.Password != "" {
		modelUser.Password = req.Password
	}

	if err := u.db.Save(&modelUser).Error; err != nil {
		log.Errorf("[CustomerRepository-4] UpdateCustomer: %v", err)
		return err
	}

	return nil
}

func NewCustomerRepository(db *gorm.DB) CustomerRepositoryInterface {
	return &CustomerRepository{db: db}
}
