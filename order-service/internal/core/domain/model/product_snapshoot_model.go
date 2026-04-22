// internal/adapter/repository/model/product_snapshot.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name         string    `gorm:"type:varchar(255)"`
	Stock        int
	Image        string `gorm:"type:text"`
	RegulerPrice int64  `gorm:"column:reguler_price"`
	SalePrice    int64  `gorm:"column:sale_price"`
	Unit         string `gorm:"type:varchar(50)"`
	Weight       int
	CreatedAt    time.Time
}
