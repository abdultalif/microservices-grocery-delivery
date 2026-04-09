package entity

import (
	"github.com/google/uuid"
)

type CategoryEntity struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name"`
	Icon        string     `json:"icon"`
	Status      string     `json:"status"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Products    []ProductEntity
}

type UpdateCategoryEntity struct {
	Name        *string
	Icon        *string
	Description *string
	ParentID    *uuid.UUID
	Status      *string
}

type QueryStringEntity struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderType string
}
