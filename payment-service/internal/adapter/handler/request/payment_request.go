package request

import "github.com/google/uuid"

type PaymentRequest struct {
	OrderID       uuid.UUID `json:"order_id" validate:"required"`
	PaymentMethod string    `json:"payment_method" validate:"required"`
	GrossAmount   int64     `json:"gross_amount" validate:"required"`
	UserID        int64     `json:"user_id" validate:"required"`
	Remarks       string    `json:"remarks"`
}
