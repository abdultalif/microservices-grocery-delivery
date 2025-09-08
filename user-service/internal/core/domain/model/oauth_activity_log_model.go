package model

import "time"

type OAuthActivityLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Provider  string    `gorm:"not null;size:50" json:"provider"`
	Action    string    `gorm:"not null;size:50" json:"action"` // login, logout, link, unlink, refresh, revoke
	IPAddress string    `gorm:"size:45" json:"ip_address"`      // Support IPv6
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	Status    string    `gorm:"not null;size:20;index" json:"status"` // success, failed
	ErrorMsg  string    `gorm:"type:text" json:"error_msg"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// Foreign key relationship
	User User `gorm:"foreignKey:UserID;references:ID"`
}

func (OAuthActivityLog) TableName() string {
	return "oauth_activity_logs"
}
