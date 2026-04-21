package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	Name         string     `json:"name"`
	Stock        int        `json:"stock"`
	Image        string     `json:"image"`
	RegulerPrice int64      `json:"price"`
	SalePrice    int64      `json:"sale_price"`
	Unit         string     `json:"unit"`
	Weight       int        `json:"weight"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
