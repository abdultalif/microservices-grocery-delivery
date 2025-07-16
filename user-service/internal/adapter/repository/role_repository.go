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

type RoleRepositoryInterface interface {
	GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error)
	Create(ctx context.Context, req entity.RoleEntity) error
	Update(ctx context.Context, req entity.RoleEntity) error
	Delete(ctx context.Context, id int64) error
}

type RoleRepository struct {
	db *gorm.DB
}

// Create implements RoleRepositoryInterface.
func (r *RoleRepository) Create(ctx context.Context, req entity.RoleEntity) error {
	modelRole := model.Role{}
	err := r.db.Where("name = ?", req.Name).First(&modelRole).Error
	if err == nil {
		log.Errorf("[RoleRepository-1] Create: Role '%s' already exists", req.Name)
		return errs.ErrUserExist
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[RoleRepository-2] Create: %v", err)
		return err
	}

	newRole := model.Role{
		Name: req.Name,
	}
	if err := r.db.Create(&newRole).Error; err != nil {
		log.Errorf("[RoleRepository-3] Create: %v", err)
		return err
	}

	return nil
}


// Delete implements RoleRepositoryInterface.
func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	
	modelRole := model.Role{}

	if err := r.db.Where("id = ?", id).Preload("Users").First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrUserNotFound
			log.Errorf("[RoleRepository-1] Delete: Role not found")
			return err
		}
		log.Errorf("[RoleRepository-2] Delete: %v", err)
		return err
	}

	if len(modelRole.Users) > 0 {
		err := errs.ErrRoleHasUsers
		log.Errorf("[RoleRepository-3] Delete: Role has users")
		return err
	}

	if err := r.db.Delete(&modelRole).Error; err != nil {
		log.Errorf("[RoleRepository-4] Delete: %v", err)
		return err
	}

	return nil
}

// GetAll implements RoleRepositoryInterface.
func (r *RoleRepository) GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	modelRoles := []model.Role{}

	if err := r.db.Where("name LIKE ?", "%"+search+"%").Find(&modelRoles).Error; err != nil {
		log.Errorf("[RoleRepository-1] GetAll: %v", err)
		return nil, err
	}

	entityRole := []entity.RoleEntity{}
	for _, modelRole := range modelRoles {
		entityRole = append(entityRole, entity.RoleEntity{
			ID:   modelRole.ID,
			Name: modelRole.Name,
		})
	}

	return entityRole, nil
}

// GetByID implements RoleRepositoryInterface.
func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	
	modelRole := model.Role{}

	if err := r.db.Where("id = ?", id).First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrUserNotFound
			log.Errorf("[RoleRepository-1] GetByID: Role not found")
			return nil, err
		}
		log.Errorf("[RoleRepository-2] GetByID: %v", err)
		return nil, err
	}

	return &entity.RoleEntity{
		ID:   modelRole.ID,
		Name: modelRole.Name,
	}, nil
}

// Update implements RoleRepositoryInterface.
func (r *RoleRepository) Update(ctx context.Context, req entity.RoleEntity) error {
	
	modelRole := model.Role{}

	if err := r.db.Where("id = ?", req.ID).First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[RoleRepository-1] Update: Role not found")
			return err
		}
		log.Errorf("[RoleRepository-2] Update: %v", err)
		return err
	}

	modelRole.Name = req.Name

	if err := r.db.Save(&modelRole).Error; err != nil {
		log.Errorf("[RoleRepository-3] Update: %v", err)
		return err
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) RoleRepositoryInterface {
	return &RoleRepository{db: db}
}
