package entity

import (
	"github.com/google/uuid"
)

type CategoryEntity struct {
	ID uuid.UUID
	ParentID *uuid.UUID
	Name string
	Icon string
	Status string 
	Slug string 
	Description string
	Products []ProductEntity
}

type UpdateCategoryEntity struct {
	Name        *string
	Icon        *string
	Description *string
	ParentID    *uuid.UUID
	Status      *string
}


type QueryStringEntity struct {
	Search string
	Page int64
	Limit int64
	OrderBy string
	OrderType string
}