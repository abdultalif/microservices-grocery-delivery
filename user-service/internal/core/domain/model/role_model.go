package model

import "time"

type Role struct {
	ID        int64    `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Users     []User `gorm:"many2many:user_role;"`
}