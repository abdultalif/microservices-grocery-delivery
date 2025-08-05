package message

import "product-service/internal/core/domain/entity"

type ProductMessage struct {
	ID           string  `json:"id"`
	CategorySlug string  `json:"category_slug"`
	ParentID     *string `json:"parent_id"`
	Name         string  `json:"name"`
	Image        string  `json:"image"`
	Description  string  `json:"description"`
	RegulerPrice float64 `json:"reguler_price"`
	SalePrice    float64 `json:"sale_price"`
	Unit         string  `json:"unit"`
	Weight       int     `json:"weight"`
	Stock        int     `json:"stock"`
	Variant      string  `json:"variant"`
	Status       string  `json:"status"`
	CategoryName string  `json:"category_name"`
	Category     entity.CategoryEntity `json:"category"`
	Child        []ProductChildMessage `json:"child"`
	CreatedAt    string                `json:"created_at"`
}

type ProductChildMessage struct {
	ID           string  `json:"ID"`
	Image        string  `json:"Image"`
	Weight       int     `json:"Weight"`
	Stock        int     `json:"Stock"`
	RegulerPrice float64 `json:"RegulerPrice"`
	SalePrice    float64 `json:"SalePrice"`
}
