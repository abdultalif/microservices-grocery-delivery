package model

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID               uuid.UUID    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrderID          uuid.UUID    `gorm:"type:uuid;not null;index" json:"order_id"`
	UserID           int64        `gorm:"type:bigint;not null;index" json:"user_id"`
	PaymentMethod    string       `gorm:"type:varchar(50);not null" json:"payment_method"`
	PaymentStatus    string       `gorm:"type:varchar(50);not null;index" json:"payment_status"`
	PaymentGatewayID *string      `gorm:"type:varchar(50);null;index" json:"payment_gateway_id,omitempty"`
	GrossAmount      float64      `gorm:"type:decimal(10,2);not null" json:"gross_amount"`
	PaymentURL       *string      `gorm:"type:text;null" json:"payment_url,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	PaymentLogs      []PaymentLog `gorm:"foreignKey:PaymentID;references:ID;constraint:OnDelete:CASCADE" json:"payment_log"`
}
