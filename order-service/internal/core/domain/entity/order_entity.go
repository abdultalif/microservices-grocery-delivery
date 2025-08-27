package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderEntity struct {
	ID            uuid.UUID
	OrderCode     string
	BuyerID       int64
	OrderDate     string
	Status        string
	TotalAmount   int64
	ShippingType  string
	ShippingFee   int64
	PaymentMethod string
	OrderTime     string
	Remarks       string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	OrderItems    []OrderItemEntity
	BuyerName     string
	BuyerEmail    string
	BuyerPhone    string
	BuyerAddress  string
	BuyerLat      string
	BuyerLng      string
}

type QueryStringEntity struct {
	Page    int64  `json:"page"`
	Limit   int64  `json:"limit"`
	Status  string `json:"status"`
	Search  string `json:"search"`
	BuyerID int64  `json:"buyer_id"`
}
