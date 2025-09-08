package model

import "time"

type User struct {
	ID         int64  `gorm:"primaryKey"`
	Name       string `gorm:"not null"`
	Email      string `gorm:"not null"`
	Password   string `gorm:"not null"`
	Phone      string `gorm:"not null"`
	Photo      string `gorm:"not null"`
	Address    string `gorm:"not null"`
	Lat        string `gorm:"not null"`
	Lng        string `gorm:"not null"`
	IsVerified bool   `gorm:"not null"`
	OauthOnly  bool   `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	Roles      []Role `gorm:"many2many:user_role;"`
}
