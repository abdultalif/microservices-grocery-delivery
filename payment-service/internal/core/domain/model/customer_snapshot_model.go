package model

import (
	"time"
)

type UserSnapshot struct {
	ID        int64      `gorm:"primaryKey"`
	Name      string     `gorm:"column:name;type:varchar(255);not null;"`
	Email     string     `gorm:"column:email;type:varchar(255);not null;"`
	Address   string     `gorm:"column:address;type:text;not null;"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
