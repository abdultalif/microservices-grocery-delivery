package request

import "github.com/google/uuid"

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Icon        string `json:"icon" validate:"required"`
	Description string `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Status      string   `json:"status" validate:"required"`
}

type UpdateCategoryRequest struct {
	Name        *string
	Icon        *string
	Description *string
	ParentID    *uuid.UUID
	Status      *string
}