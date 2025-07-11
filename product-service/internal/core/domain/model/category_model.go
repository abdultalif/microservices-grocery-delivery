package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ParentID *uuid.UUID `gorm:"index"`
	Name string
	Icon string
	Status string `gorm:"index"`
	Slug string `gorm:"index"`
	Description string
	CreatedAt time.Time
	UpdatedAt time.Time
	Products []Product `gorm:"foreignKey:CategorySlug;references:slug"`
}