package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationEntity struct {
	Email            *string
	Message          string
	NotificationType string
	ReceiverID       *int
	Subject          *string
	SentAt           *time.Time
	ReadAt           *time.Time
	Status           string
	ID               uuid.UUID
}

type NotifyQuerySting struct {
	Page      int64
	Limit     int64
	Status    string
	Search    string
	OrderBy   string
	OrderType string
	UserID    int64
	IsRead    bool
}
