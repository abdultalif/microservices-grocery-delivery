package service

import (
	"context"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
)

type RoleServiceInterface interface {
	GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error)
	Create(ctx context.Context, req entity.RoleEntity) error
	Update(ctx context.Context, req entity.RoleEntity) error
	Delete(ctx context.Context, id int64) error
}

type RoleService struct {
	repository repository.RoleRepositoryInterface
}

// Create implements RoleServiceInterface.
func (r *RoleService) Create(ctx context.Context, req entity.RoleEntity) error {
	return r.repository.Create(ctx, req)
}

// Delete implements RoleServiceInterface.
func (r *RoleService) Delete(ctx context.Context, id int64) error {
	return r.repository.Delete(ctx, id)
}

// GetAll implements RoleServiceInterface.
func (r *RoleService) GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	return r.repository.GetAll(ctx, search)
}

// GetByID implements RoleServiceInterface.
func (r *RoleService) GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	return r.repository.GetByID(ctx, id)
}

// Update implements RoleServiceInterface.
func (r *RoleService) Update(ctx context.Context, req entity.RoleEntity) error {
	return r.repository.Update(ctx, req)
}

func NewServiceRole(repository repository.RoleRepositoryInterface) RoleServiceInterface {
	return &RoleService{repository: repository}
}
