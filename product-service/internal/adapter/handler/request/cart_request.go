package request

import "github.com/google/uuid"

type CartRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int64     `json:"quantity" validate:"required,min=1"`
}
