package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshoot struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Stock        int        `json:"stock"`
	Image        string     `json:"image"`
	RegulerPrice int64      `json:"reguler_price"`
	SalePrice    int64      `json:"sale_price"`
	Unit         string     `json:"unit"`
	Weight       int        `json:"weight"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsed     *time.Time `json:"last_used"`
}
