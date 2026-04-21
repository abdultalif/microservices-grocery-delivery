package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string     `gorm:"column:name;not null;size:255"`
	Stock        int        `gorm:"column:stock;not null;default:1"`
	Image        string     `gorm:"column:image;size:255"`
	RegulerPrice float64    `gorm:"column:reguler_price;not null;default:0"`
	SalePrice    float64    `gorm:"column:sale_price;not null;default:0"`
	Unit         string     `gorm:"column:unit;default:'gram'"`
	Weight       int        `gorm:"column:weight"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	LastUsed     *time.Time `gorm:"column:last_used"`
}
