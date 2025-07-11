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

type QueryStringEntity struct {
	Search string
	Page int
	Limit int
	OrderBy string
	OrderType string
}