package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderCode string `gorm:"order_code"`
	BuyerID int64 `gorm:"buyer_id"`
	OrderDate time.Time `gorm:"order_date"`
	Status string `gorm:"status"`
	TotalAmount float64 `gorm:"total_amount"`
	ShippingType string `gorm:"shipping_type"`
	ShippingFee float64 `gorm:"shipping_fee"`
	OrderTime string `gorm:"order_time"`
	Remarks string `gorm:"remarks"`
	CreatedAt    time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID;referances:ID"`
}