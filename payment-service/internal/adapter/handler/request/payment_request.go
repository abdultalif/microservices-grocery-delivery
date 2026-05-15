package request

type PaymentRequest struct {
	OrderCode     string `json:"order_code" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"`
	GrossAmount   int64  `json:"gross_amount" validate:"required"`
	UserID        int64  `json:"user_id" validate:"required"`
	Remarks       string `json:"remarks"`
}

type CancelTransactionRequest struct {
	OrderCode string `json:"order_code" validate:"required"`
}
