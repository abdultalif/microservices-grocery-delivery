package model

import (
	"time"
)

type UserSnapshoot struct {
	ID        int64      `gorm:"primaryKey"`
	Name      string     `gorm:"column:name;type:varchar(255);not null;"`
	Email     string     `gorm:"column:email;type:varchar(255);not null;"`
	Phone     string     `gorm:"column:phone;type:varchar(20);not null;"`
	Address   string     `gorm:"column:address;type:text;not null;"`
	Lat       string     `gorm:"column:lat;type:varchar(50);"`
	Lng       string     `gorm:"column:lng;type:varchar(50);"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
