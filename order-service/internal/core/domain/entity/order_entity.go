package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderEntity struct {
	ID uuid.UUID `json:"id"`
	OrderCode string `json:"order_code"`
	BuyerID int64 `json:"buyer_id"`
	OrderDate string `json:"order_date"`
	Status string `json:"status"`
	TotalAmount int64 `json:"total_amount"`
	ShippingType string `json:"shipping_type"`
	ShipingFee int64 `json:"shipping_fee"`
	OrderTime string `json:"order_time"`
	Remarks string `json:"remarks"`
	BuyerName string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	OrderItems []OrderItemEntity
}