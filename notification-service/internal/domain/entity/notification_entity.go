package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationEntity struct {
	ReceiverEmail    *string    `json:"receiver_email"`
	Message          string     `json:"message"`
	NotificationType string     `json:"notification_type"`
	ReceiverID       *int       `json:"receiver_id"`
	Subject          *string    `json:"subject"`
	SentAt           *time.Time `json:"sent_at"`
	ReadAt           *time.Time `json:"read_at"`
	Status           string     `json:"status"`
	ID               uuid.UUID  `json:"id"`
}

type NotifyQuerySting struct {
	Page      int64  `json:"page"`
	Limit     int64  `json:"limit"`
	Status    string `json:"status"`
	Search    string `json:"search"`
	OrderBy   string `json:"order_by"`
	OrderType string `json:"order_type"`
	UserID    int64  `json:"user_id"`
	IsRead    bool   `json:"is_read"`
}
