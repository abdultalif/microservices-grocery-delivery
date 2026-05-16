package entity

import (
	"time"

	"github.com/google/uuid"
)

type PaymentEvent struct {
	OrderCode     string `json:"order_code"`
	PaymentMethod string `json:"payment_method,omitempty"`
	Status        string `json:"status"`
}

type OrderEntity struct {
	ID            uuid.UUID         `json:"id"`
	OrderCode     string            `json:"order_code"`
	BuyerID       int64             `json:"buyer_id"`
	OrderDate     string            `json:"order_date"`
	Status        string            `json:"status"`
	TotalAmount   int64             `json:"total_amount"`
	ShippingType  string            `json:"shipping_type"`
	ShippingFee   int64             `json:"shipping_fee"`
	PaymentMethod string            `json:"payment_method"`
	OrderTime     string            `json:"order_time"`
	Remarks       string            `json:"remaks"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     *time.Time        `json:"updated_at"`
	OrderItems    []OrderItemEntity `json:"order_items"`
	BuyerName     string            `json:"buyer_name"`
	BuyerEmail    string            `json:"buyer_email"`
	BuyerPhone    string            `json:"buyer_phone"`
	BuyerAddress  string            `json:"buyer_address"`
	BuyerLat      string            `json:"buyer_lat"`
	BuyerLng      string            `json:"buyer_lng"`
}

type QueryStringEntity struct {
	Page    int64  `json:"page"`
	Limit   int64  `json:"limit"`
	Status  string `json:"status"`
	Search  string `json:"search"`
	BuyerID int64  `json:"buyer_id"`
}
