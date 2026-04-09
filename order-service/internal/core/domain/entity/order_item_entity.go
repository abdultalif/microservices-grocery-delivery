package entity

import (
	"github.com/google/uuid"
)

type OrderItemEntity struct {
	ID            uuid.UUID `json:"id"`
	OrderID       uuid.UUID `json:"order_id"`
	ProductID     uuid.UUID `json:"product_id"`
	Quantity      int64     `json:"quantity"`
	OrderCode     string    `json:"order_code"`
	ProductName   string    `json:"product_name"`
	ProductImage  string    `json:"product_image"`
	ProductWeight int64     `json:"product_weight"`
	Price         int64     `json:"price"`
	ProductUnit   string    `json:"product_unit"`
	TotalPrice    int64     `json:"total_price"`
}

type PublishOrderItemEntity struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}
