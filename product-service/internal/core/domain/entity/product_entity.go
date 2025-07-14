package entity

import (
	"github.com/google/uuid"
)

type ProductEntity struct {
	ID uuid.UUID
	ParentID *uuid.UUID
	CategorySlug string
	Name string
	Image string
	Description string
	RegulerPrice float64
	SalePrice float64
	Unit string
	Weight int
	Variant int
	Status string
	CategoryName string
}
