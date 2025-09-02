package model

import (
	"time"
)

type OAuthProvider struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64      `gorm:"not null" json:"user_id"`
	Provider        string     `gorm:"not null;size:50" json:"provider"`
	ProviderUserID  string     `gorm:"not null;size:255;uniqueIndex:idx_provider_user" json:"provider_user_id"`
	ProviderEmail   string     `gorm:"not null;size:255" json:"provider_email"`
	ProviderName    string     `gorm:"not null;size:255" json:"provider_name"`
	ProviderPicture *string    `gorm:"size:500" json:"provider_picture"`
	AccessToken     *string    `gorm:"type:text" json:"access_token"`
	RefreshToken    *string    `gorm:"type:text" json:"refresh_token"`
	TokenExpiresAt  *time.Time `json:"token_expires_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

func (OAuthProvider) TableName() string {
	return "oauth_providers"
}
