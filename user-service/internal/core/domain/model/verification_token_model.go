package model

import "time"

type VerificationToken struct {
	ID        int64  `gorm:"primaryKey"`
	UserID    int64  `gorm:"index;not null"`
	Token     string `gorm:"index;not null"`
	TokenType string `gorm:"index;not null"`
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	User  	  User `gorm:"foreignKey=UserID"`
}