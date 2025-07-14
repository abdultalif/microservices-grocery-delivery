package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ParentID *uuid.UUID `gorm:"index"`
	Parent      *Category  `gorm:"foreignKey:ParentID"`
	Children    []Category `gorm:"foreignKey:ParentID"`
	Name string
	Icon string
	Status bool `gorm:"index"`
	Slug string `gorm:"index"`
	Description string
	CreatedAt time.Time
	UpdatedAt time.Time
	Products []Product `gorm:"foreignKey:CategorySlug;references:slug"`
}