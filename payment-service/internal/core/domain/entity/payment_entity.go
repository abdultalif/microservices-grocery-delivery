package entity

import "github.com/google/uuid"

type MidtransCancelResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
}

type CancelTransaction struct {
	OrderCode string `json:"order_code"`
	UserID    int64  `json:"user_id"`
}

type PaymentEvent struct {
	OrderCode     string `json:"order_code"`
	PaymentMethod string `json:"payment_method,omitempty"`
	Status        string `json:"status"`
}

type PaymentEntity struct {
	ID                uuid.UUID
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
