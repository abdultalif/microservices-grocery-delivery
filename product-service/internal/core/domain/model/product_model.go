package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ParentID *uuid.UUID `gorm:"index"`
	CategorySlug string `gorm:"index;unique;not null"`
	Name string
	Image string
	Description string
	RegulerPrice float64
	SalePrice float64 `gorm:"index"`
	Unit string
	Weight int
	Variant int
	Status string `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}