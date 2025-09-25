package entity

import "github.com/google/uuid"

type PaymentEntity struct {
	ID                uuid.UUID
	OrderID           uuid.UUID
	UserID            int64
	PaymentMethod     string
	PaymentStatus     string
	PaymentGatewayID  string
	GrossAmount       float64
	PaymentURL        string
	PaymentAt         string
	PaymentLog        []PaymentLogEntity
	Remarks           string
	OrderCode         string
	OrderShippingType string
	OrderAt           string
	OrderRemarks      string
	OrderStatus       string
	CustomerName      string
	CustomerEmail     string
	CustomerAddress   string
}

type PaymentQueryStringRequest struct {
	Limit     int64
	Page      int64
	UserID    int64
	Status    string
	OrderType string
	OrderBy   string
	Search    string
}
