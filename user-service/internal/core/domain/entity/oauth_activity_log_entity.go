package entity

import "time"

type OAuthActivityLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Provider  string    `json:"provider"`
	Action    string    `json:"action"` // login, logout, link, unlink, refresh, revoke
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Status    string    `json:"status"` // success, failed
	ErrorMsg  string    `json:"error_msg,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
