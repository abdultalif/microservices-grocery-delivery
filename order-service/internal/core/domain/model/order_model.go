package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID           uuid.UUID   `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderCode    string      `gorm:"column:order_code;unique;not null;size:64"`
	BuyerID      int64       `gorm:"column:buyer_id;not null"`
	OrderDate    time.Time   `gorm:"column:order_date;not null;default:CURRENT_TIMESTAMP"`
	Status       string      `gorm:"column:status;not null;default:'pending';size:20"`
	TotalAmount  float64     `gorm:"column:total_amount;not null;default:0"`
	ShippingType string      `gorm:"column:shipping_type;not null;default:'PICKUP';size:20"`
	ShippingFee  float64     `gorm:"column:shipping_fee;not null;default:0"`
	OrderTime    string      `gorm:"column:order_time"`
	Remarks      string      `gorm:"column:remarks"`
	CreatedAt    time.Time   `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt    *time.Time  `gorm:"column:updated_at"`
	BuyerName    string      `gorm:"column:buyer_name"`
	BuyerEmail   string      `gorm:"column:buyer_email"`
	BuyerPhone   string      `gorm:"column:buyer_phone"`
	BuyerAddress string      `gorm:"column:buyer_address"`
	BuyerLat     string      `gorm:"column:buyer_lat"`
	BuyerLng     string      `gorm:"column:buyer_lng"`
	OrderItems   []OrderItem `gorm:"foreignKey:OrderID;referances:ID"`
}
