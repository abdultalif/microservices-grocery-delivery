package service

import (
	"context"
	"errors"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"
	errs "product-service/internal/core/domain/error"
	"product-service/utils/conv"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type CategoryServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetByID(ctx context.Context, categoryId uuid.UUID) (*entity.CategoryEntity, error)
	GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error)
	Create(ctx context.Context, req entity.CategoryEntity) error
	Update(ctx context.Context, id uuid.UUID, req entity.UpdateCategoryEntity) error
	Delete(ctx context.Context, categoryID uuid.UUID) error
}

type CategoryService struct {
	repository repository.CategoryRepositoryInterface
}

// Update implements CategoryServiceInterface.
func (c *CategoryService) Update(ctx context.Context, categoryID uuid.UUID, req entity.UpdateCategoryEntity) error {
	category, err := c.repository.GetByID(ctx, categoryID)
	if err != nil {
		log.Errorf("[CategoryService-1] Update: %v", err)
		return err
	}

	if req.Name != nil {
		category.Name = *req.Name
		category.Slug = conv.GenerateSlug(*req.Name)
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Status != nil {
		category.Status = *req.Status
	}
	if req.ParentID != nil {
		
		_, err = c.repository.GetByID(ctx, *req.ParentID)
		if err != nil {
			log.Errorf("[CategoryService-3] Update: parent not found")
			return errs.ErrCategoryBadRequest
		}
		category.ParentID = req.ParentID 
	}

	err = c.repository.Update(ctx, *category)
	if err != nil {
		log.Errorf("[CategoryService-4] Update: %v", err)
		return err
	}

	return nil
}

// Create implements CategoryServiceInterface.
func (c *CategoryService) Create(ctx context.Context, req entity.CategoryEntity) error {
	slug := conv.GenerateSlug(req.Name)
	result, err := c.repository.GetBySlug(ctx, slug)
	if err != nil {
		if !errors.Is(err, errs.ErrCategoryNotFound) {
			log.Errorf("[CategoryService-1] CreateCategory: %v", err)
			return err
		}
	}

	if result != nil {
		log.Errorf("[CategoryService-2] Create: Category with slug %s already exists", slug)
		return errs.ErrCategoryConflict
	}

	if req.ParentID != nil {
		_, err := c.repository.GetByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, errs.ErrCategoryNotFound) {
				err = errs.ErrParentCategoryNotFound
				log.Errorf("[CategoryService-3] Create: parent_id %s not found", req.ParentID)
				return err
			}
			log.Errorf("[CategoryService-4] Create: %v", err)
			return err
		}
	}

	req.Slug = slug
	err = c.repository.Create(ctx, req)
	if err != nil {
		log.Errorf("[CategoryService-5] Create: %v", err)
		return err
	}

	return nil
}

// Delete implements CategoryServiceInterface.
func (c *CategoryService) Delete(ctx context.Context, categoryID uuid.UUID) error {
	return c.repository.Delete(ctx, categoryID)
}

// GetBySlug implements CategoryServiceInterface.
func (c *CategoryService) GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error) {
	return c.repository.GetBySlug(ctx, slug)
}

// GetAll implements CategoryServiceInterface.
func (c *CategoryService) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error) {
	return c.repository.GetAll(ctx, query)
}

// GetByID implements CategoryServiceInterface.
func (c *CategoryService) GetByID(ctx context.Context, categoryId uuid.UUID) (*entity.CategoryEntity, error) {
	return c.repository.GetByID(ctx, categoryId)
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &CategoryService{
		repository: repo,
	}
}
