package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID           uuid.UUID            `json:"id"`
	ParentID     *uuid.UUID           `json:"parent_id"`
	Name         string               `json:"name"`
	Image        string               `json:"image"`
	RegulerPrice float64              `json:"reguler_price"`
	SalePrice    float64              `json:"sale_price"`
	CategorySlug string               `json:"category_slug"`
	Unit         string               `json:"unit"`
	Weight       int                  `json:"weight"`
	Stock        int                  `json:"stock"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    *time.Time           `json:"updated_at"`
	Child        []ProductChildEntity `json:"child"`
}

type ProductChildEntity struct {
	ID           uuid.UUID
	Image        string
	Name         string
	Weight       int
	Stock        int
	RegulerPrice float64
	SalePrice    float64
	Unit         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
