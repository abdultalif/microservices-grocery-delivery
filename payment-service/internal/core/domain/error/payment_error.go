package error

import "errors"

var (
	ErrNotFoundBuyer   = errors.New("buyer not found")
	ErrNotFoundOrder   = errors.New("order not found")
	ErrInvalidMethod   = errors.New("Invalid payment method")
	ErrPaymentExist    = errors.New("Payment already exists")
	ErrNotFoundPayment = errors.New("Payment not found")
)
