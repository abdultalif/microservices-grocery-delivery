package entity

import (
	"github.com/google/uuid"
)

type CategoryEntity struct {
	ID uuid.UUID
	ParentID *uuid.UUID
	Name string
	Icon string
	Status bool 
	Slug string 
	Description string
	Products []ProductEntity
}

type QueryStringEntity struct {
	Search string
	Page int64
	Limit int64
	OrderBy string
	OrderType string
}