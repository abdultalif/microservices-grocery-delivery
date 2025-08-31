package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentLog struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PaymentID uuid.UUID `gorm:"type:uuid;not null" json:"payment_id"`
	Status    string    `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
