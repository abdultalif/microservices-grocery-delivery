package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProductHttpClientResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    ProductResponseEntity `json:"data"`
}

type ChildProductResponseEntity struct {
	ID           uuid.UUID `json:"id"`
	Weight       int       `json:"weight"`
	Stock        int       `json:"stock"`
	RegulerPrice float64   `json:"reguler_price"`
	SalePrice    float64   `json:"sale_price"`
	Unit         string    `json:"unit"`
	Image        string    `json:"image"`
}

type ProductResponseEntity struct {
	ID            uuid.UUID                    `json:"id"`
	ProductName   string                       `json:"product_name"`
	ParentID      uuid.UUID                    `json:"parent_id"`
	ProductImage  string                       `json:"product_image"`
	CategoryName  string                       `json:"category_name"`
	ProductStatus string                       `json:"product_status"`
	SalePrice     float64                      `json:"sale_price"`
	RegulerPrice  float64                      `json:"reguler_price"`
	CreatedAt     time.Time                    `json:"created_at"`
	Unit          string                       `json:"unit"`
	Weight        int                          `json:"weight"`
	Stock         int                          `json:"stock"`
	Child         []ChildProductResponseEntity `json:"child"`
}

type ProductCustomerResponse struct {
	ID           uuid.UUID                 `json:"id"`
	Name         string                    `json:"name"`
	Image        string                    `json:"image"`
	Stock        int                       `json:"stock"`
	SalePrice    int64                     `json:"sale_price"`
	RegulerPrice int64                     `json:"reguler_price"`
	Unit         string                    `gorm:"column:unit;default:'gram'"`
	Weight       int                       `gorm:"column:weight;default:0"`
	CreatedAt    time.Time                 `json:"created_at"`
	Child        []ProductCustomerResponse `json:"child"`
}
