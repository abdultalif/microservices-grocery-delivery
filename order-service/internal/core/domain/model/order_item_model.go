package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID 			uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID 	*uuid.UUID `gorm:"order_id"`
	ProductID 	*uuid.UUID `gorm:"product_id"`
	Quantity 	int64 `gorm:"quantity"`
	CreatedAt   time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at"`
	Order 		Order `gorm:"foreignKey:OrderId;referances:ID"`
}