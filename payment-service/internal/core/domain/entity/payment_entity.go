package entity

import "github.com/google/uuid"

type PaymentEntity struct {
	ID               uuid.UUID
	OrderID          uuid.UUID
	UserID           int64
	PaymentMethod    string
	PaymentStatus    string
	PaymentGatewayID string
	GrossAmount      float64
	PaymentURL       string
	PaymentLog       []PaymentLogEntity
	Remarks          string
	CustomerName     string
	CustomerEmail    string
}
