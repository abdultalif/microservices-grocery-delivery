package entity

import (
	"time"
)

type UserSnapshot struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name"`
	Email     string    `gorm:"column:email"`
	Phone     string    `gorm:"column:phone"`
	Address   string    `gorm:"column:address"`
	Lat       string    `gorm:"column:lat"`
	Lng       string    `gorm:"column:lng"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}
