package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID           uuid.UUID                    `gorm:"primaryKey;type:uuid" json:"id"`
	Name         string                       `gorm:"type:varchar(255)" json:"name"`
	Stock        int                          `json:"stock"`
	Image        string                       `gorm:"type:text" json:"image"`
	RegulerPrice int64                        `gorm:"column:reguler_price" json:"reguler_price"`
	SalePrice    int64                        `gorm:"column:sale_price" json:"sale_price"`
	Unit         string                       `gorm:"type:varchar(50)" json:"unit"`
	Weight       int                          `json:"weight"`
	CreatedAt    time.Time                    `json:"created_at"`
	UpdatedAt    *time.Time                   `json:"updated_at"`
	Child        []ChildProductResponseEntity `json:"child"`
}

type ProductSnapshotPayload struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Stock        int       `json:"stock"`
	Image        string    `json:"image"`
	RegulerPrice int64     `json:"reguler_price"`
	SalePrice    int64     `json:"sale_price"`
	Unit         string    `json:"unit"`
	Weight       int       `json:"weight"`
	CreatedAt    time.Time `json:"created_at"`
}
