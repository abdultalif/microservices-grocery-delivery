// internal/adapter/repository/model/product_snapshot.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid"`
	ParentID     *uuid.UUID `gorm:"index;constraint:OnDelete:CASCADE;"`
	Name         string     `gorm:"type:varchar(255)"`
	Image        string     `gorm:"type:text"`
	RegulerPrice float64    `gorm:"column:reguler_price"`
	SalePrice    float64    `gorm:"column:sale_price"`
	Unit         string     `gorm:"type:varchar(50)"`
	Weight       int        `gorm:"type:int"`
	Stock        int        `gorm:"type:int"`
	CreatedAt    time.Time  `gorm:"type:timestamp"`
	UpdatedAt    *time.Time `gorm:"type:timestamp"`

	Childs []ProductSnapshot `gorm:"foreignKey:ParentID;references:ID"`
}
