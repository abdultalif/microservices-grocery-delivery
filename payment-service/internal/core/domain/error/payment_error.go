package error

import "errors"

var (
	ErrNotFoundBuyer   = errors.New("buyer not found")
	ErrNotFoundOrder   = errors.New("order not found")
	ErrInvalidMethod   = errors.New("Invalid payment method")
	ErrPaymentExist    = errors.New("Payment already exists")
	ErrNotFoundPayment = errors.New("Payment not found")

	ErrStatusCodeNotFound   = errors.New("status code not found")
	ErrOrderIDNotFound      = errors.New("order id not found")
	ErrGrossAmountNotFound  = errors.New("gross amount not found")
	ErrSignatureKeyNotFound = errors.New("signature key not found")
)
