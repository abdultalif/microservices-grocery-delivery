package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID   uuid.UUID  `gorm:"column:order_id;type:uuid;not null;references:orders.id;onDelete:CASCADE"`
	ProductID uuid.UUID  `gorm:"column:product_id;type:uuid;not null"`
	Quantity  int64      `gorm:"column:quantity;not null;default:1"`
	CreatedAt time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	Order     Order      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
