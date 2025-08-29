package entity

import "github.com/google/uuid"

type CartItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}
