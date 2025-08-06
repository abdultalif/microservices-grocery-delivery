package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProductEntity struct {
	ID           uuid.UUID       `json:"id"`
	CategorySlug string          `json:"category_slug"`
	ParentID     *uuid.UUID      `json:"parent_id"`
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	Description  string          `json:"description"`
	RegulerPrice float64         `json:"reguler_price"`
	SalePrice    float64         `json:"sale_price"`
	Unit         string          `json:"unit"`
	Weight       int             `json:"weight"`
	Stock        int             `json:"stock"`
	Variant      string             `json:"variant"`
	Status       string          `json:"status"`
	CategoryName string          `json:"category_name"`
	Category     CategoryEntity      `json:"category"`
	Child []ProductChildEntity `json:"child"`
	CreatedAt    time.Time       `json:"created_at"`
}

type QueryStringProduct struct {
	Search       string
	Page         int
	Limit        int
	OrderBy      string
	OrderType    string
	CategorySlug string
	StartPrice   int64
	EndPrice     int64
	Status       string
}

type PublishOrderItemEntity struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type ProductChildEntity struct {
	ID           uuid.UUID
	Image        string
	Weight       int
	Stock        int
	RegulerPrice float64
	SalePrice    float64
}
